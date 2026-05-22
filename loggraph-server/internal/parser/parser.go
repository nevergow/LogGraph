package parser

import (
	"regexp"
	"strings"

	"loggraph/internal/model"
)

var (
	reStrikethrough = regexp.MustCompile(`~~(.+?)~~`)
	reBlocked       = regexp.MustCompile(`\[BLOCK\]`)
	reTag           = regexp.MustCompile(`#([^\s#@^~\[\].,;:!?，。；：！？]+)`)
	reMention       = regexp.MustCompile(`@([^\s#@^~\[\].,;:!?，。；：！？]+)`)
	reReference     = regexp.MustCompile(`\^([0-9a-fA-F-]{36})`)
)

// ParseResult holds everything extracted from a Block's content.
type ParseResult struct {
	Status     model.Status
	Tags       []string // #project / #standard
	Mentions   []string // @person
	References []string // ^block-id
}

// Parse analyses raw markdown content and returns structured data.
func Parse(content string) ParseResult {
	r := ParseResult{Status: model.StatusActive}

	if reBlocked.MatchString(content) {
		r.Status = model.StatusBlocked
	} else if reStrikethrough.MatchString(content) {
		r.Status = model.StatusCompleted
	}

	for _, m := range reTag.FindAllStringSubmatch(content, -1) {
		r.Tags = append(r.Tags, strings.TrimSpace(m[1]))
	}
	for _, m := range reMention.FindAllStringSubmatch(content, -1) {
		r.Mentions = append(r.Mentions, strings.TrimSpace(m[1]))
	}
	for _, m := range reReference.FindAllStringSubmatch(content, -1) {
		r.References = append(r.References, strings.TrimSpace(m[1]))
	}

	return r
}
