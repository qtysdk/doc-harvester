package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/qrtt1/doc-harvester/pkg/harvester"
	"github.com/qrtt1/doc-harvester/pkg/tree"
)

// Global debug flag
var debug bool

// ExploreWebsite explores the website structure without downloading content
func ExploreWebsite(urlStr string, maxDepth int) {
	// Create website exploration context
	explorerCtx, err := harvester.NewExplorerContext(urlStr, maxDepth, debug)
	if err != nil {
		fmt.Printf("Failed to create explorer context: %s\n", err)
		return
	}

	// Perform website exploration
	if err := explorerCtx.Explore(); err != nil {
		fmt.Printf("Failed to explore website: %s\n", err)
	}
}

// DownloadWebsite downloads website content and saves it locally
func DownloadWebsite(url string, baseURL string, maxDepth int) {
	// Create download context using XML storage
	xmlFilePath := "docs.xml"
	fmt.Printf("Using default XML output file: %s\n", xmlFilePath)

	// Ensure directory exists
	dirPath := filepath.Dir(xmlFilePath)
	if dirPath != "." {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Printf("Failed to create directory for XML file: %v\n", err)
			return
		}
	}

	// Create download context using XML storage
	downloaderCtx, err := harvester.NewXMLDownloaderContext(url, xmlFilePath, baseURL, maxDepth, debug)
	if err != nil {
		fmt.Printf("Failed to create XML downloader context: %s\n", err)
		return
	}

	// Set to download all pages
	downloaderCtx.DownloadAll = true

	// Execute download
	if err := downloaderCtx.Download(); err != nil {
		fmt.Printf("Failed to download website: %s\n", err)
		return
	}

	// Cleanup work (save XML file)
	downloaderCtx.Cleanup()

	fmt.Printf("XML download completed successfully. File saved to: %s\n", xmlFilePath)
}

// getDomain extracts domain from URL
func getDomain(url string) string {
	webTree, err := tree.NewWebTree(url, 0)
	if err != nil {
		return ""
	}

	if webTree.RootNode != nil && webTree.RootNode.URL != nil {
		return webTree.RootNode.URL.Host
	}

	return ""
}

func main() {
	// Define CLI flags
	exploreOnly := flag.Bool("explore-only", false, "Only explore the website structure without downloading content")
	xmlOutput := flag.String("xml-output", "", "Path to save content as a single XML file")
	debugFlag := flag.Bool("debug", false, "Enable debug messages")
	maxDepth := flag.Int("max-depth", 2, "Maximum depth for web crawling (default: 2)")

	// Parse CLI flags
	flag.Parse()

	// Set global debug flag
	debug = *debugFlag

	// Validate arguments
	if len(flag.Args()) < 1 {
		fmt.Println("Usage: harvester [options] <URL>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	url := flag.Args()[0]

	// Determine the XML output file path
	xmlFilePath := "docs.xml"
	if *xmlOutput != "" {
		xmlFilePath = *xmlOutput
	}

	// Handle the download logic
	if *exploreOnly {
		fmt.Printf("Exploring website structure for URL: %s with max depth: %d\n", url, *maxDepth)
		ExploreWebsite(url, *maxDepth)
	} else {
		fmt.Printf("Downloading content from URL: %s to XML file: %s with max depth: %d\n", url, xmlFilePath, *maxDepth)
		DownloadWebsite(url, url, *maxDepth)
	}
}
