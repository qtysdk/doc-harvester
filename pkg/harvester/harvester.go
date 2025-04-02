package harvester

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"github.com/qrtt1/doc-harvester/pkg/crawler"
	"github.com/qrtt1/doc-harvester/pkg/extractor"
	"github.com/qrtt1/doc-harvester/pkg/node"
	"github.com/qrtt1/doc-harvester/pkg/storage"
	"github.com/qrtt1/doc-harvester/pkg/tree"
)

// Storage defines the storage interface
type Storage interface {
	// SaveNodeContent saves the content of a node
	SaveNodeContent(node *node.WebNode, content string) error
	// CreateIndexFile creates an index file
	CreateIndexFile(path string) error
}

// NullStorage is used for exploration mode, doesn't actually store content
type NullStorage struct{}

// SaveNodeContent implements empty operation
func (s *NullStorage) SaveNodeContent(node *node.WebNode, content string) error {
	// Does not actually save any content
	return nil
}

// CreateIndexFile implements empty operation
func (s *NullStorage) CreateIndexFile(path string) error {
	// Does not actually create any file
	return nil
}

// HarvesterContext encapsulates all components and operations related to website exploration and downloading
type HarvesterContext struct {
	Crawler     *crawler.Crawler
	WebTree     *tree.WebTree
	Extractor   *extractor.ContentExtractor
	Storage     Storage
	RootURL     string
	BaseURL     string
	MaxDepth    int
	Debug       bool
	DownloadAll bool            // Whether to download all pages
	PrintedURLs map[string]bool // Used to track URLs that have been output
}

// NewExplorerContext creates a new exploration context (without downloading content)
func NewExplorerContext(rootURL string, maxDepth int, debug bool) (*HarvesterContext, error) {
	// Create crawler
	c := crawler.NewCrawler()

	// Create web tree
	webTree, err := tree.NewWebTree(rootURL, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to create web tree: %w", err)
	}

	// Create content extractor
	e := extractor.NewContentExtractor()

	// Create null storage (does not actually save files)
	s := &NullStorage{}

	return &HarvesterContext{
		Crawler:     c,
		WebTree:     webTree,
		Extractor:   e,
		Storage:     s,
		RootURL:     rootURL,
		BaseURL:     rootURL,
		MaxDepth:    maxDepth,
		Debug:       debug,
		PrintedURLs: make(map[string]bool), // Initialize printed URLs map
	}, nil
}

// NewDownloaderContext creates a new download context (actually downloads content)
func NewDownloaderContext(rootURL string, outputFilePath string, baseURL string, maxDepth int, debug bool) (*HarvesterContext, error) {
	// Create crawler
	c := crawler.NewCrawler()

	// Create web tree
	webTree, err := tree.NewWebTree(rootURL, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to create web tree: %w", err)
	}

	// Create content extractor
	e := extractor.NewContentExtractor()

	// Use XML storage instead of original LocalStorage
	s, err := storage.NewXMLStorage(outputFilePath, rootURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create XML storage: %w", err)
	}

	return &HarvesterContext{
		Crawler:     c,
		WebTree:     webTree,
		Extractor:   e,
		Storage:     s,
		RootURL:     rootURL,
		BaseURL:     baseURL,
		MaxDepth:    maxDepth,
		Debug:       debug,
		PrintedURLs: make(map[string]bool), // Initialize printed URLs map
	}, nil
}

// NewXMLDownloaderContext creates a download context using XML storage
func NewXMLDownloaderContext(rootURL string, xmlFilePath string, baseURL string, maxDepth int, debug bool) (*HarvesterContext, error) {
	// Create crawler
	c := crawler.NewCrawler()

	// Create web tree
	webTree, err := tree.NewWebTree(rootURL, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("failed to create web tree: %w", err)
	}

	// Create content extractor
	e := extractor.NewContentExtractor()

	// Create XML storage
	s, err := storage.NewXMLStorage(xmlFilePath, rootURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create XML storage: %w", err)
	}

	return &HarvesterContext{
		Crawler:     c,
		WebTree:     webTree,
		Extractor:   e,
		Storage:     s,
		RootURL:     rootURL,
		BaseURL:     baseURL,
		MaxDepth:    maxDepth,
		Debug:       debug,
		PrintedURLs: make(map[string]bool),
	}, nil
}

// Cleanup performs cleanup tasks, such as stopping auto-save
func (hc *HarvesterContext) Cleanup() {
	// Check if it's XMLStorage
	if xmlStorage, ok := hc.Storage.(*storage.XMLStorage); ok {
		// Stop auto-save
		xmlStorage.StopAutoSave()

		// Save one last time
		if err := xmlStorage.SaveToFile(); err != nil {
			fmt.Printf("Error saving XML file during cleanup: %v\n", err)
		}
	}
}

// isParentURL determines if a URL is a parent URL
func (hc *HarvesterContext) isParentURL(link string) bool {
	currentURL, err := url.Parse(hc.RootURL)
	if err != nil {
		return false
	}

	linkURL, err := url.Parse(link)
	if err != nil {
		return false
	}

	// Must be the same host
	if currentURL.Host != linkURL.Host {
		return false
	}

	// Full path processing
	currentPath := strings.TrimRight(currentURL.Path, "/")
	linkPath := strings.TrimRight(linkURL.Path, "/")

	// Get the parent path of the current URL
	lastSlash := strings.LastIndex(currentPath, "/")
	if lastSlash == -1 {
		return false
	}

	parentPath := currentPath[:lastSlash]

	// Debug information
	if hc.Debug {
		fmt.Printf("Current path: %s\n", currentPath)
		fmt.Printf("Parent path: %s\n", parentPath)
		fmt.Printf("Link path: %s\n", linkPath)
	}

	// Relaxed condition: Check if it's a parent path or contains parent path characteristics
	if linkPath == parentPath {
		return true
	}

	// Check if it's a prompt-engineering page
	if strings.Contains(parentPath, "prompt-engineering") && strings.Contains(linkPath, "prompt-engineering") {
		return true
	}

	return false
}

// removeFragment removes the fragment part from a URL
func (hc *HarvesterContext) removeFragment(linkStr string) string {
	parsedURL, err := url.Parse(linkStr)
	if err != nil {
		return linkStr // If parsing fails, return the original link
	}

	// Clear fragment
	parsedURL.Fragment = ""
	return parsedURL.String()
}

// processLink processes a single link (exploration mode)
func (hc *HarvesterContext) processLink(link string) {
	// Only show parent URLs and remove fragments
	if hc.isParentURL(link) {
		cleanLink := hc.removeFragment(link)

		// Check if URL has already been output
		if !hc.PrintedURLs[cleanLink] {
			fmt.Printf("<a href=\"%s\">\n", cleanLink)
			// Mark as output
			hc.PrintedURLs[cleanLink] = true
		}
	} else if hc.Debug {
		// Filtered links, only show in debug mode
		if hc.WebTree.IsVisited(link) {
			fmt.Printf("Filtered (duplicated): %s\n", link)
		} else {
			fmt.Printf("Filtered (not parent): %s\n", link)
		}
	}
}

// Explore explores the website structure without downloading content
func (hc *HarvesterContext) Explore() error {
	// Get the HTML content of the initial page
	doc, err := hc.Crawler.FetchPage(hc.RootURL)
	if err != nil {
		return fmt.Errorf("failed to fetch the URL: %w", err)
	}

	// Extract title
	title := hc.Crawler.ExtractTitle(doc)
	rootNode := hc.WebTree.RootNode
	rootNode.Title = title

	// Extract all links
	links, err := hc.Crawler.ExtractLinks(doc, hc.RootURL)
	if err != nil {
		return fmt.Errorf("failed to extract links: %w", err)
	}

	// Process each link
	for _, link := range links {
		hc.processLink(link)
	}

	return nil
}

// Download downloads website content
func (hc *HarvesterContext) Download() error {
	fmt.Printf("Downloading content from URL: %s\n", hc.RootURL)

	// Get the HTML content of the initial page
	doc, err := hc.Crawler.FetchPage(hc.RootURL)
	if err != nil {
		return fmt.Errorf("failed to fetch the URL: %w", err)
	}

	// Extract title
	title := hc.Crawler.ExtractTitle(doc)
	rootNode := hc.WebTree.RootNode
	rootNode.Title = title

	// Extract content
	content, err := hc.Extractor.ExtractContent(doc)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	// Save content
	if err := hc.Storage.SaveNodeContent(rootNode, content); err != nil {
		return fmt.Errorf("failed to save content: %w", err)
	}

	// Extract all links
	links, err := hc.Crawler.ExtractLinks(doc, hc.RootURL)
	if err != nil {
		return fmt.Errorf("failed to extract links: %w", err)
	}

	fmt.Printf("Found %d links on the page.\n", len(links))

	// Process each link
	for _, link := range links {
		hc.processLinkAndDownload(link)
	}

	// Create index file
	if rootNode.URL != nil {
		indexPath := rootNode.URL.Path
		if err := hc.Storage.CreateIndexFile(indexPath); err != nil && hc.Debug {
			fmt.Printf("Failed to create index file: %s\n", err)
		}
	}

	return nil
}

// processLinkAndDownload processes a single link and downloads it (download mode)
func (hc *HarvesterContext) processLinkAndDownload(link string) {
	// Only process parent URLs
	if hc.isParentURL(link) {
		cleanLink := hc.removeFragment(link)

		// Check if URL has already been output
		if !hc.PrintedURLs[cleanLink] {
			fmt.Printf("<a href=\"%s\">\n", cleanLink)
			// Mark as output
			hc.PrintedURLs[cleanLink] = true
		}

		// If download all pages is enabled
		if hc.DownloadAll {
			// Parse link
			parsedURL := hc.WebTree.FindNode(hc.RootURL)
			parsedLink, _ := hc.WebTree.AddURL(link, parsedURL)

			if parsedLink != nil && parsedLink.URL != nil {
				// Get page content
				doc, err := hc.Crawler.FetchPage(parsedLink.URL.String())
				if err != nil {
					fmt.Printf("Failed to fetch: %s - %s\n", parsedLink.URL.String(), err)
					return
				}

				// Extract title
				title := hc.Crawler.ExtractTitle(doc)
				parsedLink.Title = title

				// Extract content
				content, err := hc.Extractor.ExtractContent(doc)
				if err != nil {
					fmt.Printf("Failed to extract content: %s - %s\n", parsedLink.URL.String(), err)
					return
				}

				// Save content
				if err := hc.Storage.SaveNodeContent(parsedLink, content); err != nil {
					fmt.Printf("Failed to save content: %s - %s\n", parsedLink.URL.String(), err)
				}
			}
		}
	} else if hc.Debug {
		// Filtered links, only show in debug mode
		if hc.WebTree.IsVisited(link) {
			fmt.Printf("Filtered (duplicated): %s\n", link)
		} else {
			fmt.Printf("Filtered (not parent): %s\n", link)
		}
	}
}

// GetTree returns the website tree structure
func (hc *HarvesterContext) GetTree() *tree.WebTree {
	return hc.WebTree
}

// FetchDocument gets the document for a specified URL
func (hc *HarvesterContext) FetchDocument(url string) (*html.Node, error) {
	return hc.Crawler.FetchPage(url)
}
