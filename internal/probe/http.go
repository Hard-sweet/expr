package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type HTTPResult struct {
	Scheme   string
	Status   string
	Server   string
	Title    string
	Location string
}

func HTTP(ctx context.Context, timeout time.Duration, host string, port int, path string) (HTTPResult, error) {
	schemes := []string{"http", "https"}
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	for _, scheme := range schemes {
		url := fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return HTTPResult{}, err
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		result := HTTPResult{
			Scheme: scheme,
			Status: resp.Status,
			Server: resp.Header.Get("Server"),
			Title:  extractTitle(string(body)),
		}
		if location := resp.Header.Get("Location"); location != "" {
			result.Location = location
		}
		return result, nil
	}
	return HTTPResult{}, fmt.Errorf("http probe failed")
}

var titlePattern = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func extractTitle(body string) string {
	matches := titlePattern.FindStringSubmatch(body)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(stripTags(matches[1]))
}

func stripTags(value string) string {
	replacer := strings.NewReplacer("\n", " ", "\r", " ", "\t", " ")
	return replacer.Replace(value)
}
