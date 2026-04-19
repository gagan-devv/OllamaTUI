package components

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/ollama/ollama/api"
)

// SearchResult represents a single search match
type SearchResult struct {
	MessageIndex int    // Index of the message containing the match
	Position     int    // Character position within the message content
	Length       int    // Length of the matched text
	Context      string // Surrounding context for display
}

// SearchOptions configures search behavior
type SearchOptions struct {
	CaseSensitive bool
	WholeWord     bool
	Regex         bool
}

// SearchEngine handles conversation search functionality
type SearchEngine struct {
	query       string
	options     SearchOptions
	results     []SearchResult
	currentIdx  int
	messages    []api.Message
	highlighted map[int][]HighlightRange // Message index -> highlight ranges
}

// HighlightRange represents a range of text to highlight
type HighlightRange struct {
	Start  int
	End    int
	Active bool // Whether this is the currently active match
}

// NewSearchEngine creates a new search engine
func NewSearchEngine() *SearchEngine {
	return &SearchEngine{
		currentIdx:  -1,
		highlighted: make(map[int][]HighlightRange),
	}
}

// Search performs a search across all messages
func (se *SearchEngine) Search(query string, messages []api.Message, options SearchOptions) error {
	se.query = query
	se.options = options
	se.messages = messages
	se.results = nil
	se.currentIdx = -1
	se.highlighted = make(map[int][]HighlightRange)

	if query == "" {
		return nil
	}

	// Prepare search pattern
	pattern := query
	if !options.Regex {
		pattern = regexp.QuoteMeta(query)
	}
	
	if options.WholeWord {
		pattern = `\b` + pattern + `\b`
	}

	flags := ""
	if !options.CaseSensitive {
		flags = "(?i)"
	}

	regex, err := regexp.Compile(flags + pattern)
	if err != nil {
		return err
	}

	// Search through all messages
	for msgIdx, message := range messages {
		// Skip system messages for search
		if message.Role == "system" {
			continue
		}

		content := message.Content
		matches := regex.FindAllStringIndex(content, -1)

		var ranges []HighlightRange
		for _, match := range matches {
			start, end := match[0], match[1]
			
			// Create search result
			context := se.extractContext(content, start, end)
			result := SearchResult{
				MessageIndex: msgIdx,
				Position:     start,
				Length:       end - start,
				Context:      context,
			}
			se.results = append(se.results, result)

			// Create highlight range
			ranges = append(ranges, HighlightRange{
				Start:  start,
				End:    end,
				Active: false,
			})
		}

		if len(ranges) > 0 {
			se.highlighted[msgIdx] = ranges
		}
	}

	// Set first match as active if any results found
	if len(se.results) > 0 {
		se.currentIdx = 0
		se.updateActiveMatch()
	}

	return nil
}

// extractContext extracts surrounding context for a match
func (se *SearchEngine) extractContext(content string, start, end int) string {
	const contextLength = 50

	// Find context boundaries
	contextStart := start - contextLength
	if contextStart < 0 {
		contextStart = 0
	}

	contextEnd := end + contextLength
	if contextEnd > len(content) {
		contextEnd = len(content)
	}

	// Adjust to word boundaries if possible
	for contextStart > 0 && !unicode.IsSpace(rune(content[contextStart])) {
		contextStart--
	}

	for contextEnd < len(content) && !unicode.IsSpace(rune(content[contextEnd])) {
		contextEnd++
	}

	context := content[contextStart:contextEnd]
	
	// Add ellipsis if truncated
	if contextStart > 0 {
		context = "..." + context
	}
	if contextEnd < len(content) {
		context = context + "..."
	}

	return strings.TrimSpace(context)
}

// NextMatch moves to the next search result
func (se *SearchEngine) NextMatch() *SearchResult {
	if len(se.results) == 0 {
		return nil
	}

	se.currentIdx = (se.currentIdx + 1) % len(se.results)
	se.updateActiveMatch()
	return &se.results[se.currentIdx]
}

// PreviousMatch moves to the previous search result
func (se *SearchEngine) PreviousMatch() *SearchResult {
	if len(se.results) == 0 {
		return nil
	}

	se.currentIdx--
	if se.currentIdx < 0 {
		se.currentIdx = len(se.results) - 1
	}
	se.updateActiveMatch()
	return &se.results[se.currentIdx]
}

// updateActiveMatch updates which match is currently active
func (se *SearchEngine) updateActiveMatch() {
	// Clear all active flags
	for msgIdx, ranges := range se.highlighted {
		for i := range ranges {
			ranges[i].Active = false
		}
		se.highlighted[msgIdx] = ranges
	}

	// Set current match as active
	if se.currentIdx >= 0 && se.currentIdx < len(se.results) {
		result := se.results[se.currentIdx]
		if ranges, exists := se.highlighted[result.MessageIndex]; exists {
			for i, r := range ranges {
				if r.Start == result.Position {
					ranges[i].Active = true
					break
				}
			}
			se.highlighted[result.MessageIndex] = ranges
		}
	}
}

// GetCurrentMatch returns the current search result
func (se *SearchEngine) GetCurrentMatch() *SearchResult {
	if se.currentIdx >= 0 && se.currentIdx < len(se.results) {
		return &se.results[se.currentIdx]
	}
	return nil
}

// GetMatchCount returns the total number of matches
func (se *SearchEngine) GetMatchCount() int {
	return len(se.results)
}

// GetCurrentMatchIndex returns the current match index (1-based)
func (se *SearchEngine) GetCurrentMatchIndex() int {
	if se.currentIdx >= 0 {
		return se.currentIdx + 1
	}
	return 0
}

// HighlightText applies highlighting to message content
func (se *SearchEngine) HighlightText(content string, msgIdx int, highlightStyle, activeStyle lipgloss.Style) string {
	ranges, exists := se.highlighted[msgIdx]
	if !exists || len(ranges) == 0 {
		return content
	}

	// Sort ranges by start position (descending) to apply highlights from end to start
	// This prevents position shifts when inserting ANSI codes
	sortedRanges := make([]HighlightRange, len(ranges))
	copy(sortedRanges, ranges)
	
	for i := 0; i < len(sortedRanges)-1; i++ {
		for j := i + 1; j < len(sortedRanges); j++ {
			if sortedRanges[i].Start < sortedRanges[j].Start {
				sortedRanges[i], sortedRanges[j] = sortedRanges[j], sortedRanges[i]
			}
		}
	}

	result := content
	for _, r := range sortedRanges {
		style := highlightStyle
		if r.Active {
			style = activeStyle
		}

		before := result[:r.Start]
		match := result[r.Start:r.End]
		after := result[r.End:]

		result = before + style.Render(match) + after
	}

	return result
}

// ClearHighlights removes all search highlights
func (se *SearchEngine) ClearHighlights() {
	se.query = ""
	se.results = nil
	se.currentIdx = -1
	se.highlighted = make(map[int][]HighlightRange)
}

// HasResults returns true if there are search results
func (se *SearchEngine) HasResults() bool {
	return len(se.results) > 0
}

// GetQuery returns the current search query
func (se *SearchEngine) GetQuery() string {
	return se.query
}