package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type RedirectHop struct {
	Step       int
	StatusCode int
	StatusText string
	URL        string
}

type RedirectResult struct {
	InitialURL string
	FinalURL   string
	Hops       []RedirectHop
	TotalHops  int
	IsRedirect bool
}

type URLMetadata struct {
	URL         string
	Title       string
	Description string
	SiteName    string
	ImageURL    string
	Canonical   string
	ContentType string
	StatusCode  int
	ContentSize int64
	Server      string
}

func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return "https://" + raw
	}
	return raw
}

func ShortenURL(ctx context.Context, longURL string) (string, error) {
	normURL := NormalizeURL(longURL)
	if _, err := url.ParseRequestURI(normURL); err != nil {
		return "", fmt.Errorf("URL tidak valid: %s", longURL)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	tinyReqURL := "https://tinyurl.com/api-create.php?url=" + url.QueryEscape(normURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tinyReqURL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "PikoBot/1.0")
		resp, err := client.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					res := strings.TrimSpace(string(body))
					if strings.HasPrefix(res, "http") {
						return res, nil
					}
				}
			}
		}
	}

	isgdReqURL := "https://is.gd/create.php?format=simple&url=" + url.QueryEscape(normURL)
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, isgdReqURL, nil)
	if err != nil {
		return "", fmt.Errorf("gagal membuat request shortlink: %w", err)
	}
	req2.Header.Set("User-Agent", "PikoBot/1.0")
	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("gagal menghubungi server pemendek tautan: %w", err)
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", fmt.Errorf("gagal membaca balasan pemendek tautan: %w", err)
	}

	res2 := strings.TrimSpace(string(body2))
	if !strings.HasPrefix(res2, "http") {
		return "", fmt.Errorf("gagal membuat shortlink: %s", res2)
	}

	return res2, nil
}

func CheckRedirects(ctx context.Context, targetURL string) (*RedirectResult, error) {
	currURL := NormalizeURL(targetURL)
	if _, err := url.ParseRequestURI(currURL); err != nil {
		return nil, fmt.Errorf("URL tidak valid: %s", targetURL)
	}

	client := &http.Client{
		Timeout: 12 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {

			return http.ErrUseLastResponse
		},
	}

	result := &RedirectResult{
		InitialURL: currURL,
		Hops:       make([]RedirectHop, 0),
	}

	visited := make(map[string]bool)
	step := 1

	for step <= 15 {
		if visited[currURL] {
			return nil, fmt.Errorf("terdeteksi loop redirect tak berujung pada: %s", currURL)
		}
		visited[currURL] = true

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, currURL, nil)
		if err != nil {
			return nil, fmt.Errorf("gagal membuat request redirect: %w", err)
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("gagal mengakses URL di tahap %d: %w", step, err)
		}
		_ = resp.Body.Close()

		hop := RedirectHop{
			Step:       step,
			StatusCode: resp.StatusCode,
			StatusText: http.StatusText(resp.StatusCode),
			URL:        currURL,
		}
		result.Hops = append(result.Hops, hop)

		if resp.StatusCode >= 300 && resp.StatusCode < 400 {
			loc := resp.Header.Get("Location")
			if loc == "" {
				break
			}

			parsedLoc, err := url.Parse(loc)
			if err != nil {
				break
			}
			baseURL, _ := url.Parse(currURL)
			nextURL := baseURL.ResolveReference(parsedLoc).String()

			currURL = nextURL
			step++
		} else {
			break
		}
	}

	result.FinalURL = currURL
	result.TotalHops = len(result.Hops)
	result.IsRedirect = result.TotalHops > 1 || (len(result.Hops) == 1 && result.Hops[0].StatusCode >= 300 && result.Hops[0].StatusCode < 400)

	return result, nil
}

func FetchURLMetadata(ctx context.Context, targetURL string) (*URLMetadata, error) {
	normURL := NormalizeURL(targetURL)
	if _, err := url.ParseRequestURI(normURL); err != nil {
		return nil, fmt.Errorf("URL tidak valid: %s", targetURL)
	}

	client := &http.Client{Timeout: 12 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, normURL, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request URL: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PikoBot/1.0; +https://t.me/pikobot)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengakses URL: %w", err)
	}
	defer resp.Body.Close()

	meta := &URLMetadata{
		URL:         resp.Request.URL.String(),
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		ContentSize: resp.ContentLength,
		Server:      resp.Header.Get("Server"),
	}

	limitReader := io.LimitReader(resp.Body, 2*1024*1024)
	doc, err := html.Parse(limitReader)
	if err != nil {
		return meta, nil
	}

	parseHTMLNodes(doc, meta)

	if meta.Title == "" && meta.SiteName != "" {
		meta.Title = meta.SiteName
	}

	return meta, nil
}

func parseHTMLNodes(n *html.Node, meta *URLMetadata) {
	if n.Type == html.ElementNode {
		switch strings.ToLower(n.Data) {
		case "title":
			if n.FirstChild != nil && meta.Title == "" {
				meta.Title = strings.TrimSpace(n.FirstChild.Data)
			}
		case "meta":
			var name, prop, content string
			for _, attr := range n.Attr {
				switch strings.ToLower(attr.Key) {
				case "name":
					name = strings.ToLower(attr.Val)
				case "property":
					prop = strings.ToLower(attr.Val)
				case "content":
					content = strings.TrimSpace(attr.Val)
				}
			}

			if content != "" {
				if (name == "description" || prop == "og:description") && meta.Description == "" {
					meta.Description = content
				}
				if prop == "og:title" && (meta.Title == "" || len(content) > len(meta.Title)) {
					meta.Title = content
				}
				if prop == "og:site_name" && meta.SiteName == "" {
					meta.SiteName = content
				}
				if prop == "og:image" && meta.ImageURL == "" {
					meta.ImageURL = content
				}
			}
		case "link":
			var rel, href string
			for _, attr := range n.Attr {
				if strings.ToLower(attr.Key) == "rel" {
					rel = strings.ToLower(attr.Val)
				}
				if strings.ToLower(attr.Key) == "href" {
					href = strings.TrimSpace(attr.Val)
				}
			}
			if rel == "canonical" && href != "" && meta.Canonical == "" {
				meta.Canonical = href
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parseHTMLNodes(c, meta)
	}
}
