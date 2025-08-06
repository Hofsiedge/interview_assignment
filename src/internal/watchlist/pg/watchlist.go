package pg

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"interview/src/internal/client"
	"interview/src/internal/watchlist"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresWatchList struct {
	sync.RWMutex

	dbpool  *pgxpool.Pool
	client  client.Client
	symbols map[string]struct{}
}

func New(pool *pgxpool.Pool, client client.Client) (*PostgresWatchList, error) {
	list := PostgresWatchList{
		dbpool:  pool,
		client:  client,
		symbols: map[string]struct{}{},
	}

	err := list.loadWatchList()

	return &list, err
}

func (wl *PostgresWatchList) loadWatchList() error {
	wl.Lock()
	defer wl.Unlock()

	rows, err := wl.dbpool.Query(context.Background(), `SELECT * FROM tracking`)
	if err != nil {
		return err
	}

	symbols, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return err
	}

	for _, symbol := range symbols {
		wl.symbols[symbol] = struct{}{}
	}

	slog.Info("loaded tracking symbols", slog.Any("symbols", symbols))

	return nil
}

// Add implements watchlist.WatchList.
func (wl *PostgresWatchList) Add(coin string) error {
	wl.Lock()
	defer wl.Unlock()

	if _, found := wl.symbols[coin]; found {
		return fmt.Errorf("duplicate coin symbol")
	}

	_, err := wl.dbpool.Exec(context.Background(),
		`INSERT INTO tracking (coin) VALUES ($1)`, coin)
	if err != nil {
		return err
	}

	wl.symbols[coin] = struct{}{}

	return nil
}

// GetClosestPrice implements watchlist.WatchList.
func (wl *PostgresWatchList) GetClosestPrice(coin string, timestamp uint64) (float64, error) {
	wl.RLock()
	defer wl.RUnlock()

	var price float64

	err := wl.dbpool.QueryRow(context.Background(),
		`SELECT price FROM prices WHERE coin = $1 ORDER BY ABS(timestamp - $2) LIMIT 1`,
		coin, uint64(time.Now().Unix())).Scan(&price)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, watchlist.ErrNotFound
		} else {
			slog.Error("query error", slog.String("error", err.Error()))

			return 0, err
		}
	}

	return price, nil
}

// Remove implements watchlist.WatchList.
func (wl *PostgresWatchList) Remove(coin string) error {
	wl.Lock()
	defer wl.Unlock()

	if _, found := wl.symbols[coin]; !found {
		return watchlist.ErrNotFound
	}

	_, err := wl.dbpool.Exec(context.Background(),
		`DELETE FROM tracking WHERE coin=$1`, coin)
	if err != nil {
		return err
	}

	delete(wl.symbols, coin)

	return nil
}

// Poll implements watchlist.WatchList.
func (wl *PostgresWatchList) Poll(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("started polling", slog.Any("interval", interval))

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopped polling")
			return
		case <-ticker.C:
			if err := wl.getPrices(); err != nil {
				slog.Error("error fetching prices", slog.Any("error", err))
			}
		}
	}
}

func (wl *PostgresWatchList) getPrices() error {
	wl.Lock()
	defer wl.Unlock()

	var (
		batches [][]string
		batch   []string
	)

	// split items into batches of up to 50 items
	for symbol := range wl.symbols {
		batch = append(batch, symbol)
		if len(batch) == 50 {
			batches = append(batches, batch)
			batch = nil // reset batch
		}
	}

	if len(batch) > 0 {
		batches = append(batches, batch)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(batches))

	for _, batch := range batches {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := wl.processBatch(batch); err != nil {
				errCh <- err
			}
		}()

	}

	wg.Wait()
	close(errCh)

	var goroutineErrs []error
	for err := range errCh {
		goroutineErrs = append(goroutineErrs, err)
	}

	if len(goroutineErrs) > 0 {
		return fmt.Errorf("some batches failed: %v", goroutineErrs)
	}

	return nil
}

func (wl *PostgresWatchList) processBatch(batch []string) error {
	prices, err := wl.client.GetPrices(batch...)
	if err != nil {
		return err
	}

	timestamp := uint64(time.Now().Unix())
	data := make([][]any, 0)

	for symbol, price := range prices {
		data = append(data, []any{symbol, price, timestamp})
	}

	_, err = wl.dbpool.CopyFrom(context.Background(),
		pgx.Identifier{"prices"},
		[]string{"coin", "price", "timestamp"},
		pgx.CopyFromRows(data),
	)

	return err
}
