package client

type Client interface {
	GetPrices(coins []string) (map[string]float64, error)
}
