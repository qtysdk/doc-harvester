package node

import (
	"net/url"
)

// WebNode represents a single node in the website structure
type WebNode struct {
	URL         *url.URL          // Full URL of the node
	Title       string            // Page title
	ContentType string            // Content type (HTML, PDF, etc.)
	Children    []*WebNode        // List of child nodes
	Parent      *WebNode          // Reference to parent node
	Depth       int               // Depth level in the tree
	Metadata    map[string]string // Additional information (like size, last modified time)
}

// NewWebNode creates a new WebNode instance
func NewWebNode(urlStr string, parent *WebNode) (*WebNode, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	depth := 0
	if parent != nil {
		depth = parent.Depth + 1
	}

	return &WebNode{
		URL:         parsedURL,
		Children:    make([]*WebNode, 0),
		Parent:      parent,
		Depth:       depth,
		Metadata:    make(map[string]string),
		ContentType: "text/html", // Default type
	}, nil
}

// AddChild adds a child node
func (n *WebNode) AddChild(child *WebNode) {
	if child != nil {
		n.Children = append(n.Children, child)
	}
}

// URLWithoutFragment returns the URL without the fragment part
func (n *WebNode) URLWithoutFragment() string {
	if n.URL == nil {
		return ""
	}

	urlCopy := *n.URL
	urlCopy.Fragment = ""
	return urlCopy.String()
}

// IsAnchorOfSamePage determines if a given URL is an anchor of the current page
func (n *WebNode) IsAnchorOfSamePage(other *url.URL) bool {
	if n.URL == nil || other == nil {
		return false
	}

	urlCopy := *other
	urlCopy.Fragment = ""

	nodeCopy := *n.URL
	nodeCopy.Fragment = ""

	return urlCopy.String() == nodeCopy.String() && other.Fragment != ""
}

// IsSameOrNextLevel determines if a given URL is at the same level or next level
func (n *WebNode) IsSameOrNextLevel(other *url.URL) bool {
	if n.URL == nil || other == nil {
		return false
	}

	// Different domains, return false directly
	if n.URL.Host != other.Host {
		return false
	}

	basePath := n.URL.Path
	targetPath := other.Path

	// Remove trailing slashes
	basePath = trimRightSlash(basePath)
	targetPath = trimRightSlash(targetPath)

	// Get parent path from basePath
	parentPath := getParentPath(basePath)

	// Check if it's the same level (sibling node)
	if isPathPrefixed(targetPath, parentPath) {
		remainingPath := targetPath[len(parentPath):]
		remainingPath = trimLeftSlash(remainingPath)

		// Same level: no additional path segments or exactly one path segment
		segments := countPathSegments(remainingPath)
		if segments == 0 {
			return true
		}
	}

	// Check if it's a direct child node of the base URL
	if isPathPrefixed(targetPath, basePath) {
		remainingPath := targetPath[len(basePath):]
		remainingPath = trimLeftSlash(remainingPath)

		// Next level: exactly one path segment
		return countPathSegments(remainingPath) == 0
	}

	return false
}

// Helper functions

// trimRightSlash removes slashes from the right side of a string
func trimRightSlash(s string) string {
	return trimRight(s, '/')
}

// trimLeftSlash removes slashes from the left side of a string
func trimLeftSlash(s string) string {
	return trimLeft(s, '/')
}

// trimRight removes specified character from the right side of a string
func trimRight(s string, c byte) string {
	if len(s) == 0 {
		return s
	}

	end := len(s) - 1
	for end >= 0 && s[end] == c {
		end--
	}

	return s[:end+1]
}

// trimLeft removes specified character from the left side of a string
func trimLeft(s string, c byte) string {
	if len(s) == 0 {
		return s
	}

	start := 0
	for start < len(s) && s[start] == c {
		start++
	}

	return s[start:]
}

// isPathPrefixed determines if a path starts with a prefix
func isPathPrefixed(path, prefix string) bool {
	if prefix == "" {
		return true
	}

	if len(path) < len(prefix) {
		return false
	}

	return path[:len(prefix)] == prefix
}

// getParentPath gets the parent path of a path
func getParentPath(path string) string {
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			lastSlash = i
			break
		}
	}

	if lastSlash < 0 {
		return "/"
	}

	return path[:lastSlash]
}

// countPathSegments counts the number of segments in a path
func countPathSegments(path string) int {
	if path == "" {
		return 0
	}

	count := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			count++
		}
	}

	// If the path doesn't end with /, add the last segment
	if len(path) > 0 && path[len(path)-1] != '/' {
		count++
	}

	return count
}
