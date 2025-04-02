package crawler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

// Crawler handles web crawling logic
type Crawler struct {
	UserAgent      string        // Simulated browser information
	RequestTimeout time.Duration // Request timeout
	Client         *http.Client  // HTTP client
}

// NewCrawler creates a new Crawler instance
func NewCrawler() *Crawler {
	return &Crawler{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		RequestTimeout: 10 * time.Second,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchPage fetches HTML content of a single page
func (c *Crawler) FetchPage(urlStr string) (*html.Node, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %d %s", resp.StatusCode, resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	return doc, nil
}

// ExtractLinks extracts all links from HTML
func (c *Crawler) ExtractLinks(doc *html.Node, baseURLStr string) ([]string, error) {
	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	var links []string
	var extractFunc func(*html.Node)

	extractFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					hrefURL, err := url.Parse(attr.Val)
					if err != nil {
						continue
					}
					fullURL := baseURL.ResolveReference(hrefURL)
					links = append(links, fullURL.String())
					break
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			extractFunc(child)
		}
	}

	extractFunc(doc)
	return links, nil
}

// IsSameDomain checks if two URLs belong to the same domain
func (c *Crawler) IsSameDomain(url1, url2 string) bool {
	u1, err := url.Parse(url1)
	if err != nil {
		return false
	}

	u2, err := url.Parse(url2)
	if err != nil {
		return false
	}

	return u1.Host == u2.Host
}

// ExtractTitle extracts the title from HTML
func (c *Crawler) ExtractTitle(doc *html.Node) string {
	var title string
	var extractTitleFunc func(*html.Node)

	extractTitleFunc = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				title = n.FirstChild.Data
				return
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			extractTitleFunc(child)
			if title != "" {
				return
			}
		}
	}

	extractTitleFunc(doc)
	return title
}
