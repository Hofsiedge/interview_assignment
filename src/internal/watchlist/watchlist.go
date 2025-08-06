package watchlist

import (
	"context"
	"fmt"
	"time"
)

type WatchList interface {
	Add(coin string) error
	Remove(coin string) error
	GetClosestPrice(coin string, timestamp uint64) (float64, error)
	Poll(ctx context.Context, interval time.Duration)
}

var ErrNotFound = fmt.Errorf("not found")
