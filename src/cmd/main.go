package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"interview/src/internal/api"
	"interview/src/internal/client/coingecko"
	"interview/src/internal/config"
	pgwatchlist "interview/src/internal/watchlist/pg"
)

func main() {
	// config
	cfg, err := config.ReadConfig()
	if err != nil {
		slog.Error("could not load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// postgres connection pool
	dbpool, err := pgxpool.New(context.Background(), cfg.PostgresURL)
	if err != nil {
		slog.Error("unable to create connection pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbpool.Close()

	// controller
	coinGeckoClient := coingecko.New(cfg.CoinGeckoToken, http.DefaultClient)

	pgwatch, err := pgwatchlist.New(dbpool, coinGeckoClient)
	if err != nil {
		slog.Error("could not create a watch list", slog.String("error", err.Error()))
		os.Exit(1)
	}

	controller := api.New(pgwatch, coinGeckoClient)

	currencySubrouter := http.NewServeMux()
	controller.RegisterHandlers(currencySubrouter)

	mainRouter := http.NewServeMux()
	mainRouter.Handle("/currency/",
		http.StripPrefix("/currency", currencySubrouter))

	go pgwatch.Poll(context.Background(), time.Second*time.Duration(cfg.PollingInterval))

	_ = http.ListenAndServe(":80", mainRouter)
}
