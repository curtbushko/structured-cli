package git

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Blame represents the structured output of 'git blame --porcelain'.
type Blame struct {
	// File is the path of the blamed file.
	File string `json:"file"`

	// Lines contains the blame information for each line.
	Lines []BlameLine `json:"lines"`
}

// BlameLine represents blame information for a single line.
type BlameLine struct {
	// LineNumber is the line number in the current file.
	LineNumber int `json:"lineNumber"`

	// Hash is the commit hash that last modified this line.
	Hash string `json:"hash"`

	// Author is the name of the author who last modified this line.
	Author string `json:"author"`

	// Date is the date when this line was last modified.
	Date string `json:"date,omitempty"`

	// Content is the content of the line.
	Content string `json:"content"`
}

// BlameParser parses the output of 'git blame --porcelain'.
type BlameParser struct {
	schema domain.Schema
	// Regex to match header line: <hash> <orig-line> <final-line> <num-lines>
	headerRe *regexp.Regexp
}

// NewBlameParser creates a new BlameParser with the git-blame schema.
func NewBlameParser() *BlameParser {
	return &BlameParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-blame.json",
			"Git Blame Output",
			"object",
			map[string]domain.PropertySchema{
				"file":  {Type: "string", Description: "File path"},
				"lines": {Type: "array", Description: "Blame information per line"},
			},
			[]string{"file", "lines"},
		),
		// Match: <40-char hash> <orig-line> <final-line> [<num-lines>]
		headerRe: regexp.MustCompile(`^([a-f0-9]{40})\s+(\d+)\s+(\d+)`),
	}
}

// Parse reads git blame porcelain output and returns structured data.
func (p *BlameParser) Parse(r io.Reader) (domain.ParseResult, error) {
	blame := &Blame{
		Lines: []BlameLine{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder
	var currentLine *BlameLine

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		// Check for header line (starts with hash)
		if matches := p.headerRe.FindStringSubmatch(line); matches != nil {
			// Save previous line if exists
			if currentLine != nil {
				blame.Lines = append(blame.Lines, *currentLine)
			}

			lineNum, _ := strconv.Atoi(matches[3])
			currentLine = &BlameLine{
				Hash:       matches[1],
				LineNumber: lineNum,
			}
			continue
		}

		// Parse metadata lines
		if currentLine != nil {
			switch {
			case strings.HasPrefix(line, "author "):
				currentLine.Author = strings.TrimPrefix(line, "author ")

			case strings.HasPrefix(line, "author-time "):
				timestamp := strings.TrimPrefix(line, "author-time ")
				currentLine.Date = timestamp

			case strings.HasPrefix(line, "filename "):
				if blame.File == "" {
					blame.File = strings.TrimPrefix(line, "filename ")
				}

			case strings.HasPrefix(line, "\t"):
				// Content line (starts with tab)
				currentLine.Content = strings.TrimPrefix(line, "\t")
			}
		}
	}

	// Add the last line
	if currentLine != nil {
		blame.Lines = append(blame.Lines, *currentLine)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(blame, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for git blame output.
func (p *BlameParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *BlameParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "blame"
}
