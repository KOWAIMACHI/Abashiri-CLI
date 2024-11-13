package discovery

import (
	"abashiri-cli/storage"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/buger/jsonparser"
)

type URLEnumerationService struct {
	urlStorage storage.URLStorage
	httpClient *HTTPClient
	option     *Option
}

func NewURLEumerationService(us storage.URLStorage) *URLEnumerationService {
	return &URLEnumerationService{
		httpClient: newHTTPClient(),
		urlStorage: us,
	}
}

func (ues *URLEnumerationService) StartScan(ctx context.Context, domain string) error {
	var wg sync.WaitGroup
	scanFunctions := map[string](func(string) ([]string, error)){
		"waybackmachine": ues.enumURLFromWayBackMachine,
	}

	f := func(method string, scanfunc func(string) ([]string, error), ch chan<- []string) {
		defer wg.Done()
		result, err := scanfunc(domain)
		if err != nil {
			log.Printf("Error in %s:%v", method, err)
			return
		}
		ch <- result
	}

	resultChan := make(chan []string, len(scanFunctions))
	for method, scanFunc := range scanFunctions {
		wg.Add(1)
		go f(method, scanFunc, resultChan)
	}

	wg.Wait()
	close(resultChan)

	var results []string
	for result := range resultChan {
		results = append(results, result...)
	}

	return ues.urlStorage.RegisterURLs(ctx, domain, results)
}

func (ues *URLEnumerationService) enumURLFromWayBackMachine(domain string) ([]string, error) {
	log.Printf("[+] WaybackMachine enumeration for %v", domain)
	apiURL := fmt.Sprintf("https://web.archive.org/web/timemap/json?url=%s&matchType=prefix&collapse=urlkey&output=json&fl=original&filter=&limit=10000", domain)
	resp, err := ues.httpClient.GET(apiURL)
	if err != nil {
		log.Printf("faled to request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to parse response: %v", err)
		return nil, err
	}

	var results []string
	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		url, err := jsonparser.GetString(value, "[0]")
		if err != nil {
			log.Printf("failed to parse response: %v", err)
			return
		}
		// waybackmachineのAPIレスポンスにゴミデータが入ることがあるので削除
		if url != "original" {
			results = append(results, url)
		}
	})

	return results, nil
}

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

func (c *HTTPClient) GET(url string) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i < c.maxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

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
