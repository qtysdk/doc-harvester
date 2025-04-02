# DocHarvester Design Document

## Overview

DocHarvester is a Go-based web content harvesting tool designed to explore website structures and extract content for offline use. It features a streamlined architecture focused on efficient crawling, content extraction, and storage capabilities.

## Core Architecture

DocHarvester follows a modular design with clear separation of concerns:

```
DocHarvester
├── cmd/           # Command-line interface
└── pkg/           # Core functionality packages
    ├── crawler/   # Web page fetching
    ├── extractor/ # Content extraction
    ├── node/      # Web node representation
    ├── tree/      # Website tree structure
    ├── storage/   # Content storage
    └── harvester/ # High-level operations
```

## Key Components

### 1. WebNode

Represents a single page in the website structure:

```go
type WebNode struct {
    URL         *url.URL          // Full URL of the node
    Title       string            // Page title
    ContentType string            // Content type (HTML, PDF, etc.)
    Children    []*WebNode        // List of child nodes
    Parent      *WebNode          // Reference to parent node
    Depth       int               // Depth level in the tree
    Metadata    map[string]string // Additional information
}
```

### 2. WebTree

Manages the entire website structure:

```go
type WebTree struct {
    RootNode    *node.WebNode   // Root node
    MaxDepth    int             // Maximum exploration depth
    VisitedURLs map[string]bool // Set of visited URLs
}
```

### 3. ContentExtractor

Responsible for extracting and cleaning content from web pages:

```go
type ContentExtractor struct {
    // Configuration properties
}

// Key methods:
// - ExtractContent(): Get main content from HTML
// - ExtractMainContent(): Focus on article body
// - ExtractMetadata(): Get metadata like title, author
// - ConvertToMarkdown(): Format conversion
```

### 4. Crawler

Handles the mechanics of fetching web pages:

```go
type Crawler struct {
    UserAgent      string        // Browser identification
    RequestTimeout time.Duration // Timeout settings
    Client         *http.Client  // HTTP client
}

// Key methods:
// - FetchPage(): Retrieve a single page
// - ExtractLinks(): Parse links from HTML
// - IsSameDomain(): Domain comparison
```

### 5. XMLStorage

Manages the storage of downloaded content in XML format:

```go
type XMLStorage struct {
    FilePath     string        // Path to XML file
    Document     *XMLDocument  // XML document structure
    SaveInterval time.Duration // Auto-save timing
    stopAutoSave chan bool     // Control channel
}

// Key methods:
// - SaveToFile(): Write to disk
// - SaveNodeContent(): Add node content to XML
```

## Data Flow

### Exploration Flow

1. User initiates exploration with a starting URL and depth
2. WebTree creates a root node and initializes tracking structures
3. Crawler fetches the root page and extracts links
4. For each valid link (respecting depth/domain rules):
    - Create new WebNode
    - Add to WebTree
    - Recursively process if within depth limit
5. Return completed WebTree structure

```
CLI Arguments → Parse URL/Depth → Create WebTree → Crawl Pages → Build Tree Structure
```

### Download Flow

1. User initiates download with a starting URL and parameters
2. System explores website structure as above
3. For each node in the tree:
    - Fetch page content
    - Extract and clean main content
    - Store in XML format
4. Save final XML document to file

```
CLI Arguments → Explore Website → Process Each Node → Extract Content → Save to XML
```

## XML Output Structure

```xml
<document rootUrl="..." createdAt="...">
  <page url="..." title="..." path="..." lastFetched="...">
    <content>
      <!-- Cleaned HTML content -->
    </content>
    <links>
      <link>https://...</link>
      <!-- More links -->
    </links>
  </page>
  <!-- More pages -->
</document>
```

## Command Line Interface

```
harvester [options] <URL>

Options:
  --explore-only       Only explore without downloading
  --xml-output string  Path to save XML (default: docs.xml)
  --debug              Enable debug messages
  --max-depth int      Maximum crawling depth (default: 2)
```

## Implementation Notes

- Uses standard Go libraries where possible
- Employs goroutines for concurrent processing
- Implements a simple XML-based storage system
- Features respectful crawling with proper timing and robots.txt support
- Focuses on content extraction with HTML cleaning