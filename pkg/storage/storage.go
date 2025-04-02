package storage

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/qrtt1/doc-harvester/pkg/node"
)

// XMLDocument represents the entire XML document structure
type XMLDocument struct {
	XMLName    xml.Name       `xml:"document"`
	RootURL    string         `xml:"rootUrl,attr"`
	CreatedAt  string         `xml:"createdAt,attr"`
	Pages      []XMLPage      `xml:"page"`
	pagesByURL map[string]int // Maps URL -> Pages array index for fast lookup
	mutex      sync.Mutex     // Ensures thread safety
}

// XMLPage represents the content of a single page
type XMLPage struct {
	URL         string   `xml:"url,attr"`
	Title       string   `xml:"title,attr"`
	Path        string   `xml:"path,attr"`
	LastFetched string   `xml:"lastFetched,attr"`
	Content     string   `xml:"content"`
	Links       []string `xml:"links>link,omitempty"`
}

// XMLStorage manages downloaded content as a single XML file
type XMLStorage struct {
	FilePath     string        // Path to the XML file
	Document     *XMLDocument  // XML document object
	SaveInterval time.Duration // Auto-save interval
	stopAutoSave chan bool     // Channel to stop auto-save
}

// NewXMLStorage creates a new XML storage manager
func NewXMLStorage(filePath string, rootURL string) (*XMLStorage, error) {
	// Ensure directory exists
	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Initialize XML document
	doc := &XMLDocument{
		RootURL:    rootURL,
		CreatedAt:  time.Now().Format(time.RFC3339),
		Pages:      make([]XMLPage, 0),
		pagesByURL: make(map[string]int),
	}

	storage := &XMLStorage{
		FilePath:     filePath,
		Document:     doc,
		SaveInterval: 5 * time.Minute, // Default auto-save every 5 minutes
		stopAutoSave: make(chan bool),
	}

	// Start auto-save
	go storage.autoSaveLoop()

	return storage, nil
}

// autoSaveLoop periodically auto-saves the XML document
func (s *XMLStorage) autoSaveLoop() {
	ticker := time.NewTicker(s.SaveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.SaveToFile(); err != nil {
				fmt.Printf("Error during auto-save: %v\n", err)
			}
		case <-s.stopAutoSave:
			return
		}
	}
}

// StopAutoSave stops the auto-save process
func (s *XMLStorage) StopAutoSave() {
	s.stopAutoSave <- true
}

// SaveToFile saves the XML document to a file
func (s *XMLStorage) SaveToFile() error {
	s.Document.mutex.Lock()
	defer s.Document.mutex.Unlock()

	// Encode document as XML
	xmlData, err := xml.MarshalIndent(s.Document, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %v", err)
	}

	// Add XML header
	// Add prompt reference data tag
	xmlData = append([]byte("<!-- PROMPT_REFERENCE_DATA: Web documentation harvested by DocHarvester, intended for use as reference material in prompts and context windows -->\n"), xmlData...)
	xmlData = append([]byte(xml.Header), xmlData...)

	// Write to file
	if err := os.WriteFile(s.FilePath, xmlData, 0644); err != nil {
		return fmt.Errorf("failed to write XML file: %v", err)
	}

	return nil
}

// SaveNodeContent saves node content to the XML document
func (s *XMLStorage) SaveNodeContent(webNode *node.WebNode, content string) error {
	if webNode == nil || webNode.URL == nil {
		return fmt.Errorf("invalid node or URL")
	}

	urlStr := webNode.URL.String()
	path := webNode.URL.Path

	s.Document.mutex.Lock()
	defer s.Document.mutex.Unlock()

	// Extract all links from the current page
	var links []string
	if webNode.Children != nil {
		for _, child := range webNode.Children {
			if child.URL != nil {
				links = append(links, child.URL.String())
			}
		}
	}

	// Create page object
	page := XMLPage{
		URL:         urlStr,
		Title:       webNode.Title,
		Path:        path,
		LastFetched: time.Now().Format(time.RFC3339),
		Content:     content,
		Links:       links,
	}

	// Check if page already exists
	if idx, exists := s.Document.pagesByURL[urlStr]; exists {
		// Update existing page
		s.Document.Pages[idx] = page
	} else {
		// Add new page
		s.Document.Pages = append(s.Document.Pages, page)
		s.Document.pagesByURL[urlStr] = len(s.Document.Pages) - 1
	}

	return nil
}

// CreateIndexFile implements an empty method for XML format, as index files are not needed
func (s *XMLStorage) CreateIndexFile(path string) error {
	// XML format does not need to create index files
	return nil
}
