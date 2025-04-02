# DocHarvester

DocHarvester is a lightweight tool for exploring website structures and saving content. It's designed for collecting technical documentation, creating offline references, and preparing training data for LLMs.

## Quick Start

### Build from source

```bash
# Clone the repository
git clone https://github.com/qrtt1/doc-harvester.git
cd doc-harvester

# Build the binary
make build
```

## Usage

DocHarvester has two primary modes of operation:

### 1. Explore Mode - Map website structure without downloading

```bash
./harvester --explore-only https://docs.anthropic.com
```

### 2. Download Mode - Fetch content and save as XML

```bash
./harvester https://docs.anthropic.com
```

## Command Options

```
Usage: harvester [options] <URL>

Options:
  --explore-only       Only explore the website structure without downloading content
  --xml-output string  Path to save content as a single XML file (default: docs.xml)
  --debug              Enable debug messages
  --max-depth int      Maximum depth for web crawling (default: 2)
```

## Examples

### Explore a documentation site with depth limit

```bash
./harvester --explore-only --max-depth 3 https://docs.anthropic.com
```

### Download content to a custom file

```bash
./harvester --xml-output ./output/site-docs.xml https://docs.anthropic.com
```

### Download Anthropic's documentation

```bash
./harvester --max-depth 3 https://docs.anthropic.com/en/docs/build-with-claude/prompt-engineering/overview
```

## Features

- 🔍 **Website Structure Analysis**: Visualize and understand website hierarchies
- 📥 **Selective Content Download**: Save content while preserving structure
- 🧹 **Content Cleaning**: Automatically remove ads and navigation elements
- 🔒 **Depth Control**: Limit crawling to prevent excessive downloads
- 🌐 **Domain Control**: Respect website boundaries and robots.txt

## Output XML Structure

DocHarvester saves harvested content in an XML file with the following structure:

```xml
<document rootUrl="https://example.org" createdAt="2025-04-03T10:15:30Z">
  <page url="https://example.org/path" title="Page Title" path="/path" lastFetched="2025-04-03T10:15:30Z">
    <content>
      <!-- Cleaned HTML content of the page -->
    </content>
    <links>
      <link>https://example.org/path/subpage1</link>
      <link>https://example.org/path/subpage2</link>
      <!-- More links found on the page -->
    </links>
  </page>
  <!-- More pages -->
</document>
```

Key elements:
- `<document>`: Root element with metadata about the harvest
- `<page>`: Individual webpages with their attributes
- `<content>`: Cleaned HTML content from the page
- `<links>`: List of all links found on the page

This XML format makes it easy to process the content with other tools or import into databases.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.