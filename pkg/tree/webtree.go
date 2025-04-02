package tree

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/qrtt1/doc-harvester/pkg/node"
)

// WebTree manages the entire website structure
type WebTree struct {
	RootNode    *node.WebNode   // Root node
	MaxDepth    int             // Maximum exploration depth
	VisitedURLs map[string]bool // Set of visited URLs
}

// NewWebTree creates a new WebTree instance
func NewWebTree(rootURL string, maxDepth int) (*WebTree, error) {
	rootNode, err := node.NewWebNode(rootURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create root node: %v", err)
	}

	return &WebTree{
		RootNode:    rootNode,
		MaxDepth:    maxDepth,
		VisitedURLs: make(map[string]bool),
	}, nil
}

// AddURL adds a URL to the appropriate position in the tree
func (t *WebTree) AddURL(urlStr string, parentNode *node.WebNode) (*node.WebNode, error) {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Check if URL has been visited
	urlKey := t.normalizeURL(parsedURL)
	if t.VisitedURLs[urlKey] {
		return nil, nil // URL already exists in the tree
	}

	// Create new node
	newNode, err := node.NewWebNode(urlStr, parentNode)
	if err != nil {
		return nil, err
	}

	// Add to parent node
	if parentNode != nil {
		parentNode.AddChild(newNode)
	}

	// Mark as visited
	t.VisitedURLs[urlKey] = true

	return newNode, nil
}

// IsVisited checks if a URL has been visited
func (t *WebTree) IsVisited(urlStr string) bool {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	urlKey := t.normalizeURL(parsedURL)
	return t.VisitedURLs[urlKey]
}

// IsAllowedDepth checks if exploration is allowed at the given depth
func (t *WebTree) IsAllowedDepth(depth int) bool {
	return t.MaxDepth <= 0 || depth <= t.MaxDepth
}

// FindNode finds a node corresponding to a specific URL in the tree
func (t *WebTree) FindNode(urlStr string) *node.WebNode {
	targetURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}

	return t.findNodeRecursive(t.RootNode, targetURL)
}

// Print prints the entire tree structure
func (t *WebTree) Print() {
	t.printNode(t.RootNode, 0)
}

// Helper methods

// normalizeURL standardizes a URL for comparison and deduplication
func (t *WebTree) normalizeURL(u *url.URL) string {
	if u == nil {
		return ""
	}

	result := *u
	result.Fragment = "" // Ignore fragment

	// Handle consistency of trailing slashes
	path := strings.TrimRight(result.Path, "/")
	result.Path = path

	return result.String()
}

// findNodeRecursive recursively searches for a node
func (t *WebTree) findNodeRecursive(current *node.WebNode, target *url.URL) *node.WebNode {
	if current == nil {
		return nil
	}

	// Check current node
	currentURL := current.URL
	if currentURL != nil {
		currentCopy := *currentURL
		currentCopy.Fragment = ""

		targetCopy := *target
		targetCopy.Fragment = ""

		if currentCopy.String() == targetCopy.String() {
			return current
		}
	}

	// Check child nodes
	for _, child := range current.Children {
		if found := t.findNodeRecursive(child, target); found != nil {
			return found
		}
	}

	return nil
}

// printNode prints a single node and its children
func (t *WebTree) printNode(n *node.WebNode, depth int) {
	if n == nil {
		return
	}

	// Print current node
	indent := strings.Repeat("  ", depth)
	if n.URL != nil {
		fmt.Printf("%s- %s\n", indent, n.URL.String())
	} else {
		fmt.Printf("%s- [Invalid URL]\n", indent)
	}

	// Print child nodes
	for _, child := range n.Children {
		t.printNode(child, depth+1)
	}
}
