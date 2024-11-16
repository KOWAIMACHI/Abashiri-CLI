package discovery

import (
	"abashiri-cli/helper"
	"abashiri-cli/storage"
	"context"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"

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
		"commoncrawl":    ues.enumURLFromCommonCrawl,
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

	return ues.urlStorage.RegisterURLs(ctx, domain, helper.RemoveDuplicatesFromArray(results))
}

func (ues *URLEnumerationService) enumURLFromCommonCrawl(domain string) ([]string, error) {
	log.Printf("[+] Common Crawl enumeration for %v", domain)
	apiURL := fmt.Sprintf("https://index.commoncrawl.org/CC-MAIN-2024-42-index?url=%s/*&output=json", domain)
	resp, err := ues.httpClient.GET(apiURL, nil)
	if err != nil {
		log.Printf("faled to request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to parse response: %v", err)
		return nil, err
	}

	// JSONとしてパースできるように文字列操作
	re := regexp.MustCompile(`(?m)}$`)
	b := re.ReplaceAllString(string(body), "},\n")
	json := "[" + b[:len(b)-2] + "]"

	var results []string
	jsonparser.ArrayEach([]byte(json), func(value []byte, dataType jsonparser.ValueType, offset int, _ error) {
		url, err := jsonparser.GetString(value, "url")
		if err != nil {
			m, err := jsonparser.GetString(value, "message")
			if strings.Contains(m, "No Captures found for:") {
				return
			}
			log.Printf("failed to parse response: %v", err)
			return
		}
		results = append(results, url)
	})
	return results, nil
}

func (ues *URLEnumerationService) enumURLFromWayBackMachine(domain string) ([]string, error) {
	log.Printf("[+] Wayback Machine enumeration for %v", domain)
	apiURL := fmt.Sprintf("https://web.archive.org/web/timemap/json?url=%s&matchType=prefix&collapse=urlkey&output=json&fl=original&filter=&limit=10000", domain)
	resp, err := ues.httpClient.GET(apiURL, nil)
	if err != nil {
		log.Printf("faled to request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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
