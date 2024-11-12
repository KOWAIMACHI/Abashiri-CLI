package helper

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

func ExtractSubdomains(filePath string, domain string) ([]string, error) {
	re := regexp.MustCompile(fmt.Sprintf(`[a-zA-Z0-9\.\-]+\.%s`, regexp.QuoteMeta(domain)))

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %v", err)
	}
	defer file.Close()

	var subdomains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindAllString(line, -1)
		subdomains = append(subdomains, matches...)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	if err = os.Remove(filePath); err != nil {
		return nil, fmt.Errorf("error removing file: %v", err)
	}

	return RemoveDuplicates(subdomains), nil
}
