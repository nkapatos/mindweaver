// Package markdown provides a reusable goldmark-based markdown parser
// with common extensions pre-configured.
//
// # Supported Extensions
//
// The parser supports the following goldmark extensions:
//
//   - GitHub Flavored Markdown (GFM) - tables, task lists, strikethrough, autolinks
//   - YAML Frontmatter (goldmark-meta) - metadata extraction
//   - WikiLinks (goldmark/wikilink) - [[wiki-link]] and ![[embed]] syntax
//   - Hashtags (goldmark/hashtag) - #hashtag syntax with Obsidian variant
//
// # GitHub Flavored Markdown (GFM) Features
//
// When EnableGFM is true (default), the following features are available:
//
// Task Lists:
//   - Syntax: - [x] completed task, - [ ] incomplete task
//   - AST nodes: TaskCheckBox [GFM]
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Consider extracting task lists with completion status
//
// Tables:
//   - Syntax: | Header | Header | with alignment using :---|:---:|---:
//   - AST nodes: Table, TableHeader, TableRow, TableCell [GFM]
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Consider extracting table structure and data
//
// Strikethrough:
//   - Syntax: ~~strikethrough text~~
//   - AST nodes: Strikethrough [GFM]
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Could extract for change tracking or document diff features
//
// Autolinks:
//   - Syntax: https://example.com (bare URLs)
//   - AST nodes: AutoLink
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Extract external links for link management/validation
//
// Regular Links:
//   - Syntax: [text](url)
//   - AST nodes: Link
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Extract external links separately from WikiLinks
//
// Code Blocks:
//   - Syntax: ```language with optional language identifier
//   - AST nodes: FencedCodeBlock
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Extract code blocks for syntax highlighting, execution, or search
//
// Blockquotes:
//   - Syntax: > quoted text
//   - AST nodes: Blockquote
//   - Status: PARSED (not extracted to ParseResult yet)
//   - Future: Extract for callout/admonition support (like Obsidian)
//
// # Currently Extracted Features
//
// The following features are currently extracted to ParseResult:
//
//   - Metadata: Frontmatter YAML as map[string]any
//   - WikiLinks: [[target]] and [[target|display]] with embed support ![[target]]
//   - Hashtags: #hashtag syntax (deduplicated)
//   - RawFrontmatter: YAML text without delimiters
//   - BodyWithoutFrontmatter: Markdown body without frontmatter block
//
// # Implementation Notes
//
// To add extraction for GFM features:
//  1. Add new fields to ParseResult struct
//  2. Create extraction functions similar to extractWikiLinks() and extractHashtags()
//  3. Walk the AST looking for specific node types (see parser_exploration_test.go)
//  4. Use ast.Walk() with type assertions to identify nodes
//
// Example GFM node type checking:
//   - import "github.com/yuin/goldmark/extension/ast" for GFM node types
//   - Check node.Kind().String() to identify GFM-specific nodes
//   - See parser_exploration_test.go TestParserWithGFMFeatures for examples
//
// # References
//
// Test files demonstrating all features:
//   - parser_exploration_test.go: Shows ParseResult structure and all extracted data
//   - parser_tags_test.go: Shows tag merging behavior (frontmatter vs body)
//   - testdata/markdown/project-alpha.md: Comprehensive example with all GFM features
package markdown

import (
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/wikilink"
)

// Parser wraps goldmark with configured extensions
type Parser struct {
	markdown goldmark.Markdown
	options  Options
}

// Options configure the markdown parser
type Options struct {
	// EnableWikiLinks enables [[wiki-link]] syntax
	EnableWikiLinks bool
	// EnableHashtags enables #hashtag syntax
	EnableHashtags bool
	// EnableMeta enables YAML frontmatter parsing
	EnableMeta bool
	// EnableGFM enables GitHub Flavored Markdown (tables, strikethrough, etc)
	EnableGFM bool
	// WikiLinkResolver resolves wikilink targets to URLs
	WikiLinkResolver wikilink.Resolver
	// HashtagResolver resolves hashtags to URLs
	HashtagResolver hashtag.Resolver
}

// ParseResult contains the results of parsing markdown
type ParseResult struct {
	// AST is the parsed abstract syntax tree
	AST ast.Node
	// Metadata from frontmatter (if enabled)
	Metadata map[string]any
	// RawFrontmatter is the YAML frontmatter text without --- delimiters
	RawFrontmatter string
	// BodyWithoutFrontmatter is the markdown body without the frontmatter block
	BodyWithoutFrontmatter string
	// WikiLinks extracted from the document
	WikiLinks []WikiLink
	// Hashtags extracted from the document
	Hashtags []string
}

// WikiLink represents a [[wiki-link]] in the document
type WikiLink struct {
	Target      string // Target page name
	DisplayText string // Display text (if using [[target|display]] syntax)
	Embed       bool   // Whether this is an embedded link (![[...]])
}

// DefaultOptions returns sensible defaults for markdown parsing
func DefaultOptions() Options {
	return Options{
		EnableWikiLinks: true,
		EnableHashtags:  true,
		EnableMeta:      true,
		EnableGFM:       true,
	}
}

// NewParser creates a new markdown parser with the given options
func NewParser() *Parser {
	options := DefaultOptions()
	var extensions []goldmark.Extender

	// Add GFM if enabled
	if options.EnableGFM {
		extensions = append(extensions, extension.GFM)
	}

	// Add meta (frontmatter) if enabled
	if options.EnableMeta {
		extensions = append(extensions, meta.Meta)
	}

	// Add wikilinks if enabled
	if options.EnableWikiLinks {
		wikilinkExt := &wikilink.Extender{}
		if options.WikiLinkResolver != nil {
			wikilinkExt.Resolver = options.WikiLinkResolver
		}
		extensions = append(extensions, wikilinkExt)
	}

	// Add hashtags if enabled
	if options.EnableHashtags {
		hashtagExt := &hashtag.Extender{
			Variant: hashtag.ObsidianVariant, // Support emoji and more flexible syntax
		}
		if options.HashtagResolver != nil {
			hashtagExt.Resolver = options.HashtagResolver
		}
		extensions = append(extensions, hashtagExt)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return &Parser{
		markdown: md,
		options:  options,
	}
}

// Parse parses markdown content and returns a ParseResult
func (p *Parser) Parse(source []byte) (*ParseResult, error) {
	// Parse the document
	reader := text.NewReader(source)
	ctx := parser.NewContext()
	doc := p.markdown.Parser().Parse(reader, parser.WithContext(ctx))

	result := &ParseResult{
		AST: doc,
	}

	// Extract metadata if enabled
	if p.options.EnableMeta {
		if metaData := meta.Get(ctx); metaData != nil {
			result.Metadata = metaData
		}
		// Extract raw frontmatter and body
		result.RawFrontmatter = ExtractRawFrontmatter(source)
		result.BodyWithoutFrontmatter = ExtractBodyWithoutFrontmatter(source)
	}

	// Extract wikilinks
	if p.options.EnableWikiLinks {
		result.WikiLinks = extractWikiLinks(doc, source)
	}

	// Extract hashtags
	if p.options.EnableHashtags {
		result.Hashtags = extractHashtags(doc, source)
	}

	return result, nil
}

// extractWikiLinks walks the AST and collects all wikilinks
func extractWikiLinks(node ast.Node, source []byte) []WikiLink {
	var links []WikiLink
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if link, ok := n.(*wikilink.Node); ok {
			// Get display text from node's text content
			displayText := string(link.Target)
			if n.ChildCount() > 0 {
				// If node has children, extract their text
				var textBuf []byte
				for child := n.FirstChild(); child != nil; child = child.NextSibling() {
					if textNode, ok := child.(*ast.Text); ok {
						textBuf = append(textBuf, textNode.Segment.Value(source)...)
					}
				}
				if len(textBuf) > 0 {
					displayText = string(textBuf)
				}
			}

			links = append(links, WikiLink{
				Target:      string(link.Target),
				DisplayText: displayText,
				Embed:       link.Embed,
			})
		}
		return ast.WalkContinue, nil
	})
	return links
}

// extractHashtags walks the AST and collects all hashtags
func extractHashtags(node ast.Node, source []byte) []string {
	tagMap := make(map[string]struct{})
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if tag, ok := n.(*hashtag.Node); ok {
			tagMap[string(tag.Tag)] = struct{}{}
		}
		return ast.WalkContinue, nil
	})

	// Convert map to slice
	var tags []string
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	return tags
}

// ExtractRawFrontmatter extracts the YAML frontmatter from markdown source
// without the --- delimiters. Returns empty string if no frontmatter exists.
func ExtractRawFrontmatter(source []byte) string {
	if len(source) < 4 || string(source[0:3]) != "---" {
		return ""
	}

	// Find the closing ---
	start := 3
	// Skip the newline after opening ---
	if source[start] == '\n' {
		start++
	} else if source[start] == '\r' && start+1 < len(source) && source[start+1] == '\n' {
		start += 2
	}

	// Find the closing delimiter
	end := start
	for end < len(source) {
		if source[end] == '\n' {
			// Check if next line starts with ---
			lineStart := end + 1
			if lineStart+3 <= len(source) && string(source[lineStart:lineStart+3]) == "---" {
				// Found closing delimiter
				return string(source[start:end])
			}
		}
		end++
	}

	// No closing delimiter found
	return ""
}

// ExtractBodyWithoutFrontmatter extracts the markdown body without frontmatter.
// Returns the original source if no frontmatter exists.
func ExtractBodyWithoutFrontmatter(source []byte) string {
	if len(source) < 4 || string(source[0:3]) != "---" {
		return string(source)
	}

	// Find the closing ---
	pos := 3
	// Skip the newline after opening ---
	if source[pos] == '\n' {
		pos++
	} else if source[pos] == '\r' && pos+1 < len(source) && source[pos+1] == '\n' {
		pos += 2
	}

	// Find the closing delimiter
	for pos < len(source) {
		if source[pos] == '\n' {
			// Check if next line starts with ---
			lineStart := pos + 1
			if lineStart+3 <= len(source) && string(source[lineStart:lineStart+3]) == "---" {
				// Found closing delimiter, skip past it and any newlines
				bodyStart := lineStart + 3
				if bodyStart < len(source) && source[bodyStart] == '\n' {
					bodyStart++
				} else if bodyStart+1 < len(source) && source[bodyStart] == '\r' && source[bodyStart+1] == '\n' {
					bodyStart += 2
				}
				return string(source[bodyStart:])
			}
		}
		pos++
	}

	// No closing delimiter found, return original
	return string(source)
}
