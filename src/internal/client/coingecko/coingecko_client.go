package coingecko

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DemoAPIBaseUrl = "https://api.coingecko.com/api/v3/"

type Client struct {
	client   *http.Client
	token    string
	priceURL string
}

func New(token string, httpClient *http.Client) *Client {
	return &Client{
		token:  token,
		client: httpClient,
		// URL template
		priceURL: DemoAPIBaseUrl + "simple/price?vs_currencies=usd&" +
			"precision=full&symbols=%s",
	}
}

// GetPrices returns a map of prices (in USD) of the provided coin symbols.
//
// Symbols that could not be found are skipped.
func (c *Client) GetPrices(coins []string) (map[string]float64, error) {
	url := fmt.Sprintf(c.priceURL, strings.Join(coins, ","))
	// TODO: NewRequestWithContext
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		// TODO
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-cg-demo-api-key", c.token)

	res, err := c.client.Do(req)
	if err != nil {
		// TODO
		return nil, err
	}

	defer res.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(res.Body)
	if err != nil {
		// TODO
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code: %d", res.StatusCode)
	}

	return unmarshalResponseBody(body)
}

type responseEntry struct {
	USD float64 `json:"usd"`
}

type response map[string]responseEntry

func unmarshalResponseBody(body []byte) (map[string]float64, error) {
	var responseData response

	if err := json.Unmarshal(body, &responseData); err != nil {
		// TODO
		return nil, err
	}

	currentPrices := make(map[string]float64)

	for coin, prices := range responseData {
		currentPrices[coin] = prices.USD
	}

	return currentPrices, nil
}
