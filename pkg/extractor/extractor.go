package extractor

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// ContentExtractor is responsible for extracting useful content from web pages
type ContentExtractor struct {
	// Configuration items can be added here, such as specific selectors
}

// NewContentExtractor creates a new ContentExtractor instance
func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{}
}

// ExtractContent extracts the main content of a page
func (e *ContentExtractor) ExtractContent(doc *html.Node) (string, error) {
	body := e.findNode(doc, "body")
	if body == nil {
		return "", fmt.Errorf("no body tag found in HTML")
	}

	// Remove unwanted tags (such as ads, navigation bars, etc.)
	e.removeNodes(body, []string{"nav", "header", "footer", "aside", "script", "style", "iframe", "noscript"})

	// Get the cleaned content
	content := e.renderNode(body)

	return content, nil
}

// ExtractMainContent attempts to extract the main content part of the page, usually the article body
func (e *ContentExtractor) ExtractMainContent(doc *html.Node) (string, error) {
	// Try to extract content from common content container tags
	contentContainers := []string{
		"article",
		"main",
		"div[class*='content']",
		"div[id*='content']",
		"div[class*='article']",
		"div[id*='article']",
	}

	for _, selector := range contentContainers {
		node := e.findNodeBySelector(doc, selector)
		if node != nil {
			// Remove interfering elements
			e.removeNodes(node, []string{"script", "style", "iframe", "noscript", "nav"})
			return e.renderNode(node), nil
		}
	}

	// If no specific content container is found, fall back to extracting body content
	return e.ExtractContent(doc)
}

// ExtractMetadata extracts metadata (title, author, etc.)
func (e *ContentExtractor) ExtractMetadata(doc *html.Node) map[string]string {
	metadata := make(map[string]string)

	// Extract title
	titleNode := e.findNode(doc, "title")
	if titleNode != nil && titleNode.FirstChild != nil {
		metadata["title"] = titleNode.FirstChild.Data
	}

	// Extract information from meta tags
	metaTags := e.findNodes(doc, "meta")
	for _, meta := range metaTags {
		name := ""
		content := ""

		for _, attr := range meta.Attr {
			if attr.Key == "name" || attr.Key == "property" {
				name = attr.Val
			} else if attr.Key == "content" {
				content = attr.Val
			}
		}

		if name != "" && content != "" {
			metadata[name] = content
		}
	}

	return metadata
}

// ConvertToMarkdown converts HTML to Markdown format
func (e *ContentExtractor) ConvertToMarkdown(htmlContent string) string {
	// Simple conversion, actual implementation may require more complex logic or a dedicated HTML-to-Markdown library
	md := htmlContent

	// Replace common HTML tags with Markdown syntax
	replacements := []struct {
		from string
		to   string
	}{
		{"<h1>", "# "},
		{"</h1>", "\n\n"},
		{"<h2>", "## "},
		{"</h2>", "\n\n"},
		{"<h3>", "### "},
		{"</h3>", "\n\n"},
		{"<h4>", "#### "},
		{"</h4>", "\n\n"},
		{"<h5>", "##### "},
		{"</h5>", "\n\n"},
		{"<h6>", "###### "},
		{"</h6>", "\n\n"},
		{"<p>", ""},
		{"</p>", "\n\n"},
		{"<strong>", "**"},
		{"</strong>", "**"},
		{"<b>", "**"},
		{"</b>", "**"},
		{"<em>", "_"},
		{"</em>", "_"},
		{"<i>", "_"},
		{"</i>", "_"},
		{"<code>", "`"},
		{"</code>", "`"},
		{"<pre>", "```\n"},
		{"</pre>", "\n```\n"},
		{"<blockquote>", "> "},
		{"</blockquote>", "\n\n"},
		{"<ul>", "\n"},
		{"</ul>", "\n"},
		{"<ol>", "\n"},
		{"</ol>", "\n"},
		{"<li>", "- "},
		{"</li>", "\n"},
	}

	for _, r := range replacements {
		md = strings.ReplaceAll(md, r.from, r.to)
	}

	// Handle links
	// TODO: Use regular expressions to handle more complex cases

	return md
}

// Helper methods

// findNode finds the first node with the specified tag in the HTML document
func (e *ContentExtractor) findNode(n *html.Node, tagName string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tagName {
		return n
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if found := e.findNode(child, tagName); found != nil {
			return found
		}
	}

	return nil
}

// findNodes finds all nodes with the specified tag in the HTML document
func (e *ContentExtractor) findNodes(n *html.Node, tagName string) []*html.Node {
	var nodes []*html.Node

	if n.Type == html.ElementNode && n.Data == tagName {
		nodes = append(nodes, n)
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		childNodes := e.findNodes(child, tagName)
		if len(childNodes) > 0 {
			nodes = append(nodes, childNodes...)
		}
	}

	return nodes
}

// findNodeBySelector finds a node using a simple selector
// Note: This implementation only supports simple tag and attribute selectors, not full CSS selectors
func (e *ContentExtractor) findNodeBySelector(n *html.Node, selector string) *html.Node {
	parts := strings.Split(selector, "[")

	tagName := parts[0]
	var attrCondition string
	if len(parts) > 1 {
		attrCondition = strings.TrimSuffix(parts[1], "]")
	}

	if n.Type == html.ElementNode && n.Data == tagName {
		if attrCondition == "" {
			return n // Pure tag selector
		}

		// Handle attribute selectors, e.g., div[class*='content']
		if strings.Contains(attrCondition, "*=") {
			// Contains relationship
			keyValue := strings.Split(attrCondition, "*=")
			if len(keyValue) == 2 {
				attrName := strings.TrimSpace(keyValue[0])
				attrValue := strings.Trim(keyValue[1], "'\"")

				for _, attr := range n.Attr {
					if attr.Key == attrName && strings.Contains(attr.Val, attrValue) {
						return n
					}
				}
			}
		}
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if found := e.findNodeBySelector(child, selector); found != nil {
			return found
		}
	}

	return nil
}

// removeNodes removes nodes with specified tags
func (e *ContentExtractor) removeNodes(n *html.Node, tagNames []string) {
	var next *html.Node

	for child := n.FirstChild; child != nil; child = next {
		next = child.NextSibling

		// First recursively process child nodes
		e.removeNodes(child, tagNames)

		// Check if the current node needs to be removed
		if child.Type == html.ElementNode {
			for _, tag := range tagNames {
				if child.Data == tag {
					n.RemoveChild(child)
					break
				}
			}
		}
	}
}

// renderNode converts a node to an HTML string
func (e *ContentExtractor) renderNode(n *html.Node) string {
	var buf bytes.Buffer
	err := html.Render(&buf, n)
	if err != nil {
		return ""
	}
	return buf.String()
}
