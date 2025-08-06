package api

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"interview/src/internal/client"
	"interview/src/internal/watchlist"
)

type Controller struct {
	watchlist watchlist.WatchList
	client    client.Client
}

func New(w watchlist.WatchList, c client.Client) *Controller {
	return &Controller{w, c}
}

func (c *Controller) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("POST /add/{coin}", c.makeAddHandler())
	mux.HandleFunc("POST /remove/{coin}", c.removeHandler)
	mux.HandleFunc("GET /price/{coin}", c.getPriceHandler)
}

func (c *Controller) makeAddHandler() func(w http.ResponseWriter, r *http.Request) {
	coinSymbolRegex := regexp.MustCompile(`^[a-z0-9]+$`)

	return func(w http.ResponseWriter, r *http.Request) {
		coin := r.PathValue("coin")

		if !coinSymbolRegex.MatchString(coin) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"error": "coin symbol does not match the ` +
				`expected format (digits and lowercase latin letters)."}`))

			return
		}

		err := c.watchlist.Add(coin)
		if err != nil {
			// TODO: handle errors
			return
		}

		slog.Info("coin added to watchlist", slog.String("coin", coin))

		w.WriteHeader(http.StatusOK)
	}
}

func (c *Controller) removeHandler(w http.ResponseWriter, r *http.Request) {
	coin := r.PathValue("coin")

	if err := c.watchlist.Remove(coin); err != nil {
		// TODO: handle errors
		return
	}

	slog.Info("coin removed from watchlist", slog.String("coin", coin))
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) getPriceHandler(w http.ResponseWriter, r *http.Request) {
	coin := r.PathValue("coin")

	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid query parameters format"}`))

		return
	}

	var timestamp uint64

	tsParameter := params.Get("timestamp")
	if tsParameter == "" {
		timestamp = uint64(time.Now().Unix())
	} else {
		timestamp, err = strconv.ParseUint(tsParameter, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error": "invalid timestamp format; ` +
				`expected a unix timestamp in seconds"}`))

			return
		}
	}

	price, err := c.watchlist.GetClosestPrice(coin, timestamp)
	if err != nil {
		if errors.Is(err, watchlist.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			slog.Error("unexpected error", slog.String("error", err.Error()))
		}

		return
	}

	slog.Info("requested coin price", slog.String("coin", coin))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatFloat(price, 'f', 6, 64)))
}
