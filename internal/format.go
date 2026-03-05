package internal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// Format defines the interface used to deliver results to the end user.
type Formatter interface {
	// PrintMatch formats and prints the match to `writer`.
	PrintMatch(writer io.Writer, match matchInfo) error
}

// Formatters holds available formatters
var Formatters = map[string]Formatter{
	"text":   TextFormatter{},
	"ndjson": JSONFormatter{},
	"csv":    &CSVFormatter{},
}

// TextFormatter prints the result as human readable text.
type TextFormatter struct{}

func (f TextFormatter) PrintMatch(writer io.Writer, match matchInfo) error {
	var description string
	if match.MatchType == "name" {
		description = fmt.Sprintf("possible %s (name match)", match.DisplayName)
	} else {
		str := pluralize(match.LineCount, match.RowStr)
		if match.Confidence == "low" {
			str = str + ", low confidence"
		}
		if match.RowStr == "key" {
			description = fmt.Sprintf("found %s", match.DisplayName)
		} else {
			description = fmt.Sprintf("found %s (%s)", match.DisplayName, str)
		}
	}

	yellow := color.New(color.FgYellow).SprintFunc()
	_, err := fmt.Fprintf(writer, "%s %s\n", yellow(match.Identifier+":"), description)
	if err != nil {
		return err
	}

	values := match.Values
	if values != nil {
		// squish whitespace
		// TODO show whitespace
		for i, value := range values {
			values[i] = space.ReplaceAllString(value, " ")
		}

		if len(values) > 0 {
			_, err = fmt.Fprintln(writer, "    "+strings.Join(values, ", "))
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(writer, "")
		if err != nil {
			return err
		}
	}
	return nil
}

// JSONFormatter prints the result as a JSON object.
type JSONFormatter struct{}

type jsonEntry struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	MatchType  string `json:"match_type"`
	Confidence string `json:"confidence"`
}

type jsonEntryWithMatches struct {
	jsonEntry

	Matches      []string `json:"matches"`
	MatchesCount int      `json:"matches_count"`
}

func (f JSONFormatter) PrintMatch(writer io.Writer, match matchInfo) error {
	encoder := json.NewEncoder(writer)

	entry := jsonEntry{
		Identifier: match.Identifier,
		Name:       match.RuleName,
		MatchType:  match.MatchType,
		Confidence: match.Confidence,
	}

	values := match.Values
	if values != nil {
		return encoder.Encode(jsonEntryWithMatches{
			jsonEntry:    entry,
			Matches:      values,
			MatchesCount: len(values),
		})
	} else {
		return encoder.Encode(entry)
	}
}

// CSVFormatter prints the result as CSV rows.
type CSVFormatter struct {
	headerWritten bool
}

func (f *CSVFormatter) PrintMatch(writer io.Writer, match matchInfo) error {
	w := csv.NewWriter(writer)

	if !f.headerWritten {
		if err := w.Write([]string{"source", "data_type", "row_count", "sample_value"}); err != nil {
			return err
		}
		f.headerWritten = true
	}

	rowCount := strconv.Itoa(match.LineCount)

	if len(match.Values) > 0 {
		for _, val := range match.Values {
			if err := w.Write([]string{match.Identifier, match.DisplayName, rowCount, val}); err != nil {
				return err
			}
		}
	} else {
		if err := w.Write([]string{match.Identifier, match.DisplayName, rowCount, ""}); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}
