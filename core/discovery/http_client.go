package discovery

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type HTTPClient struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
}

func newHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

func (c *HTTPClient) GET(url string, header http.Header) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < c.maxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
		for key, values := range header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
		resp, err = c.client.Do(req)
		if err != nil {
			log.Printf("Request error (attempt %d): %v\n", i+1, err)
			time.Sleep(c.retryDelay)
			continue
		}

		if resp.StatusCode == http.StatusBadRequest {
			log.Printf("Received 400 status (attempt %d). Retrying...\n", i+1)
			time.Sleep(c.retryDelay)
			resp.Body.Close()
			continue
		}
		return resp, nil
	}
	return nil, fmt.Errorf("request failed after %d retries: %v", c.maxRetries, err)
}
