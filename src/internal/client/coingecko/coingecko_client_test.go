package coingecko

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"interview/src/internal/client"
)

func TestUnmarshalResponseBody(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect map[string]float64
		err    error
	}{
		{
			"example from API",
			`{
  "bitcoin": {
    "usd": 67187.3358936566,
    "usd_market_cap": 1317802988326.25,
    "usd_24h_vol": 31260929299.5248,
    "usd_24h_change": 3.63727894677354,
    "last_updated_at": 1711356300
  }
}`,
			map[string]float64{"bitcoin": 67187.3358936566},
			nil,
		},
		{
			"multiple coins",
			`{"btc": {"usd": 67187.335}, "zoc": {"usd": 123.456}, "quux": {"usd": 678.901}}`,
			map[string]float64{"btc": 67187.335, "zoc": 123.456, "quux": 678.901},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := unmarshalResponseBody([]byte(tt.input))

			if !errors.Is(err, tt.err) {
				t.Errorf("got `%v` error, expected `%v`", err, tt.err)
			}

			if !reflect.DeepEqual(result, tt.expect) {
				t.Errorf("got %v result, expected %v", result, tt.expect)
			}
		})
	}
}

type mockRoundTripper struct {
	response  *http.Response
	validator func(*http.Request)
}

func (rt *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if rt.validator != nil {
		rt.validator(req)
	}

	return rt.response, nil
}

func TestGetPrices(t *testing.T) {
	tests := []struct {
		name     string
		coins    []string
		response string
		expect   map[string]float64
		err      error
	}{
		{
			"multiple coins",
			[]string{"btc", "zoc", "quux"},
			`{"btc": {"usd": 67187.335}, "zoc": {"usd": 123.456}, "quux": {"usd": 678.901}}`,
			map[string]float64{"btc": 67187.335, "zoc": 123.456, "quux": 678.901},
			nil,
		},
	}

	authToken := "api-auth-token"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			recorder.Header().Add("Content-Type", "application/json")
			_, _ = recorder.WriteString(tt.response)
			expectedResponse := recorder.Result()

			httpClient := http.Client{Transport: &mockRoundTripper{
				response: expectedResponse,
				validator: func(req *http.Request) {
					token := req.Header.Get("x-cg-demo-api-key")
					if token != authToken {
						t.Errorf("invalid API auth token: got %q, expected %q", token, authToken)
					}
				},
			}}

			apiClient := New(authToken, &httpClient)

			result, err := apiClient.GetPrices(tt.coins...)

			if !errors.Is(err, tt.err) {
				t.Errorf("got `%v` error, expected `%v`", err, tt.err)
			}

			if !reflect.DeepEqual(result, tt.expect) {
				t.Errorf("got %v result, expected %v", result, tt.expect)
			}
		})
	}
}

// ensure interface compliance
var _ client.Client = &Client{}
