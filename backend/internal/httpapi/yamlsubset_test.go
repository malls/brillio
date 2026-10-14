package httpapi_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

// yamlReader reads the strict YAML subset that backend/openapi.yaml is
// written in. The standard library has no YAML parser and docs/SPEC.md
// section 2 allows no other dependency, so the OpenAPI tests use this.
//
// The subset: block mappings, block sequences indented deeper than their
// key, "- key: value" items, plain and "double-quoted" scalars, double-quoted
// keys, | and |- literal blocks, full-line # comments, and JSON flow values
// ([...] or {...}), which may continue onto lines indented deeper than their
// key and are decoded by encoding/json. Scalars stay strings: there is no
// type resolution. Anything else is an error that names the line and the
// rule, so the reader never guesses: tabs, CR, trailing whitespace, document
// markers, anchors, aliases, tags, complex keys, single-quoted and folded
// scalars, duplicate keys, bad indentation, invalid JSON flow, inline
// comments, and plain scalars that start with an indicator or contain ": ".
type yamlReader struct {
	lines []string
	n     int // index of the current line
}

// parseYAMLSubset parses src into map[string]any, []any and string values,
// with each JSON flow value as encoding/json decodes it into an any. The
// root must be a mapping.
func parseYAMLSubset(src []byte) (map[string]any, error) {
	r := &yamlReader{lines: strings.Split(strings.TrimSuffix(string(src), "\n"), "\n")}
	for i, line := range r.lines {
		switch {
		case strings.ContainsRune(line, '\r'):
			return nil, lineErr(i, "CR line endings are not in the subset")
		case strings.ContainsRune(line, '\t'):
			return nil, lineErr(i, "tabs are not in the subset")
		case strings.TrimRight(line, " ") != line:
			return nil, lineErr(i, "trailing whitespace is not in the subset")
		case line == "---" || line == "..." || strings.HasPrefix(line, "--- "):
			return nil, lineErr(i, "document markers are not in the subset")
		}
	}
	if r.next() != 0 || isItem(r.lines[r.n]) {
		return nil, errors.New("the root must be a mapping at column 0")
	}
	return r.mapping(0)
}

// lineErr returns an error for the line at index i.
func lineErr(i int, format string, args ...any) error {
	return fmt.Errorf("line %d: %s", i+1, fmt.Sprintf(format, args...))
}

// indentOf returns the number of spaces that line starts with.
func indentOf(line string) int {
	return len(line) - len(strings.TrimLeft(line, " "))
}

// next skips blank lines and full-line comments and returns the indentation
// of the line it stops at, or -1 at the end of the document.
func (r *yamlReader) next() int {
	for ; r.n < len(r.lines); r.n++ {
		if s := strings.TrimLeft(r.lines[r.n], " "); s != "" && s[0] != '#' {
			return indentOf(r.lines[r.n])
		}
	}
	return -1
}

// isItem reports whether s, with its indentation removed, is a sequence item.
func isItem(s string) bool {
	return s == "-" || strings.HasPrefix(s, "- ")
}

// mapping reads "key: value" entries at column indent until a shallower line.
func (r *yamlReader) mapping(indent int) (map[string]any, error) {
	m := map[string]any{}
	for {
		switch ind := r.next(); {
		case ind < indent:
			return m, nil
		case ind > indent:
			return nil, lineErr(r.n, "indented deeper than the mapping it is in")
		}
		line := r.lines[r.n][indent:]
		if isItem(line) {
			return nil, lineErr(r.n, "sequence item where a mapping key was expected")
		}
		key, rest, err := splitKey(line)
		if err != nil {
			return nil, lineErr(r.n, "%v", err)
		}
		if _, dup := m[key]; dup {
			return nil, lineErr(r.n, "duplicate key %q", key)
		}
		v, err := r.value(rest, indent)
		if err != nil {
			return nil, err
		}
		m[key] = v
	}
}

// sequence reads "- " items at column indent until a shallower line.
func (r *yamlReader) sequence(indent int) ([]any, error) {
	var s []any
	for {
		switch ind := r.next(); {
		case ind < indent:
			return s, nil
		case ind > indent:
			return nil, lineErr(r.n, "indented deeper than the sequence it is in")
		}
		line := r.lines[r.n][indent:]
		if !isItem(line) {
			return nil, lineErr(r.n, "mapping key where a sequence item was expected")
		}
		rest := strings.TrimLeft(line[1:], " ")
		var v any
		var err error
		if isEntry(rest) {
			// A "- key: value" item: re-read the line as a mapping that starts
			// at the key's column.
			col := indent + len(line) - len(rest)
			r.lines[r.n] = strings.Repeat(" ", col) + rest
			v, err = r.mapping(col)
		} else {
			v, err = r.value(rest, indent)
		}
		if err != nil {
			return nil, err
		}
		s = append(s, v)
	}
}

// isEntry reports whether s, the text after a "- ", is a "key: value" entry.
// Keys in such items are plain.
func isEntry(s string) bool {
	return s != "" && !strings.ContainsRune(`[{"|>`, rune(s[0])) && strings.Contains(s+" ", ": ")
}

// splitKey splits a "key: value" line into the key and the value text, which
// is "" when the value is a block on the following lines.
func splitKey(line string) (key, rest string, err error) {
	if line[0] == '"' {
		if key, rest, err = jsonString(line); err != nil {
			return "", "", fmt.Errorf("invalid double-quoted key: %w", err)
		}
		if rest != ":" && !strings.HasPrefix(rest, ": ") {
			return "", "", fmt.Errorf("double-quoted key %q is not followed by \": \"", key)
		}
	} else {
		i := strings.Index(line+" ", ": ")
		if i < 0 {
			i = len(line)
		}
		key, rest = line[:i], line[i:]
		if rule := plainRule(key); rule != "" {
			return "", "", fmt.Errorf("plain key %q %s", key, rule)
		}
		if rest == "" {
			return "", "", fmt.Errorf("%q is not a \"key: value\" entry", line)
		}
	}
	return key, strings.TrimLeft(rest[1:], " "), nil
}

// value reads the value whose text, rest, follows a key or a "- " at column
// indent on the current line, and moves past every line the value spans.
func (r *yamlReader) value(rest string, indent int) (any, error) {
	at := r.n
	r.n++
	switch {
	case rest == "":
		ind := r.next()
		if ind <= indent {
			return nil, lineErr(at, "no value: a nested block must be indented deeper than its key")
		}
		if isItem(r.lines[r.n][ind:]) {
			return r.sequence(ind)
		}
		return r.mapping(ind)
	case rest == "|" || rest == "|-":
		return r.literal(at, indent, rest == "|-")
	case rest[0] == '[' || rest[0] == '{':
		return r.flow(at, rest, indent)
	case rest[0] == '"':
		s, tail, err := jsonString(rest)
		if err != nil {
			return nil, lineErr(at, "invalid double-quoted scalar (it must be a one-line JSON string): %v", err)
		}
		if tail != "" {
			return nil, lineErr(at, "text %q after a quoted scalar (inline comments are not in the subset)", tail)
		}
		return s, nil
	}
	if rule := plainRule(rest); rule != "" {
		return nil, lineErr(at, "plain scalar %q %s", rest, rule)
	}
	return rest, nil
}

// jsonString decodes the JSON string that starts s and returns the text
// after it.
func jsonString(s string) (string, string, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	var v string
	if err := dec.Decode(&v); err != nil {
		return "", "", err
	}
	return v, s[dec.InputOffset():], nil
}

// plainConstructs names the full-YAML construct that a plain scalar or key
// starting with each of these characters would begin.
var plainConstructs = map[byte]string{
	'&': "an anchor", '*': "an alias", '!': "a tag", '>': "a folded scalar",
	'|': "a block scalar other than | and |-", '?': "a complex key", '%': "a directive",
	'#': "a comment", '[': "a flow collection", '{': "a flow collection", '\'': "a single-quoted scalar",
}

// plainRule returns the subset rule that the plain scalar or key s breaks,
// or "" if it breaks none.
func plainRule(s string) string {
	switch {
	case s == "":
		return "is empty"
	case plainConstructs[s[0]] != "":
		return "starts " + plainConstructs[s[0]] + ", which is not in the subset"
	case strings.ContainsRune(",]}@`", rune(s[0])),
		strings.ContainsRune("-:", rune(s[0])) && (len(s) == 1 || s[1] == ' '):
		return fmt.Sprintf("starts with the indicator %q", s[:1])
	case strings.Contains(s, ": ") || strings.HasSuffix(s, ":"):
		return `contains ": " (quote it or use a | block)`
	case strings.Contains(s, " #"):
		return "contains an inline comment, which is not in the subset"
	}
	return ""
}

// literal reads the | (clip) or |- (strip) block scalar whose header is on
// line at. Its indentation is that of its first non-empty line, which must be
// deeper than indent, and it ends at the first non-empty line indented less.
func (r *yamlReader) literal(at, indent int, strip bool) (string, error) {
	var lines []string
	blockIndent := -1
	for ; r.n < len(r.lines); r.n++ {
		line := r.lines[r.n]
		if line == "" {
			lines = append(lines, "")
			continue
		}
		ind := indentOf(line)
		if blockIndent < 0 && ind > indent {
			blockIndent = ind
		}
		if ind < blockIndent || blockIndent < 0 {
			break
		}
		lines = append(lines, line[blockIndent:])
	}
	// Trailing blank lines are not content under either chomping style.
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return "", lineErr(at, "empty literal block")
	}
	s := strings.Join(lines, "\n")
	if !strip {
		s += "\n"
	}
	return s, nil
}

// flow reads the JSON flow value that starts with rest on line at and may
// continue onto lines indented deeper than indent. Nothing may follow it on
// its last line.
func (r *yamlReader) flow(at int, rest string, indent int) (any, error) {
	text := rest
	for last := at; ; last = r.n - 1 {
		dec := json.NewDecoder(strings.NewReader(text))
		var v any
		switch err := dec.Decode(&v); {
		case err == nil:
			if tail := text[dec.InputOffset():]; tail != "" {
				return nil, lineErr(last, "text %q after a JSON flow value (inline comments are not in the subset)", tail)
			}
			return v, nil
		case err != io.ErrUnexpectedEOF:
			return nil, lineErr(last, "invalid JSON flow value: %v", err)
		case r.n == len(r.lines):
			return nil, lineErr(at, "unterminated JSON flow value")
		}
		line := r.lines[r.n]
		if line != "" && indentOf(line) <= indent {
			return nil, lineErr(r.n, "a JSON flow value must continue indented deeper than its key")
		}
		text += "\n" + line
		r.n++
	}
}

func TestYAMLSubsetReader(t *testing.T) {
	const src = `# A full-line comment.
root:
  plain: a plain scalar, with a comma, key:value and C#
  "200": "quoted: value"
  nested:
    list:
      - one
      - name: compact
        in: query
      - ["flow", 2]
  clip: |
    First paragraph.

    Second paragraph.
  strip: |-
    # Block content, not a comment.
    No final newline.
  flow: {"a": [1, 2],
    "b": null
    }
  last: value
`
	want := map[string]any{
		"root": map[string]any{
			"plain": "a plain scalar, with a comma, key:value and C#",
			"200":   "quoted: value",
			"nested": map[string]any{
				"list": []any{"one", map[string]any{"name": "compact", "in": "query"}, []any{"flow", 2.0}},
			},
			"clip":  "First paragraph.\n\nSecond paragraph.\n",
			"strip": "# Block content, not a comment.\nNo final newline.",
			"flow":  map[string]any{"a": []any{1.0, 2.0}, "b": nil},
			"last":  "value",
		},
	}
	got, err := parseYAMLSubset([]byte(src))
	if err != nil {
		t.Fatalf("parseYAMLSubset() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseYAMLSubset() =\n%#v\nwant\n%#v", got, want)
	}

	rejects := []struct {
		name    string
		src     string
		wantErr string // a substring of the error, including the line number
	}{
		{"tab", "a:\n\tb: c\n", "line 2: tabs"},
		{"CR", "a: b\r\n", "line 1: CR line endings"},
		{"trailing whitespace", "a: b \n", "line 1: trailing whitespace"},
		{"document marker", "---\na: b\n", "line 1: document markers"},
		{"root sequence", "- a\n", "the root must be a mapping"},
		{"anchor", "a: &x b\n", "line 1: plain scalar \"&x b\" starts an anchor"},
		{"alias", "a: *x\n", "line 1: plain scalar \"*x\" starts an alias"},
		{"tag", "a: !!str b\n", "line 1: plain scalar \"!!str b\" starts a tag"},
		{"complex key", "? a\n: b\n", "line 1: plain key \"? a\" starts a complex key"},
		{"folded scalar", "a: >\n  b\n", "line 1: plain scalar \">\" starts a folded scalar"},
		{"keep chomping", "a: |+\n  b\n", "line 1: plain scalar \"|+\" starts a block scalar other than"},
		{"empty literal block", "a: |\nb: c\n", "line 1: empty literal block"},
		{"duplicate key", "a: b\na: c\n", "line 2: duplicate key \"a\""},
		{"deeper line after a scalar", "a: b\n  c\n", "line 2: indented deeper than the mapping"},
		{"bad dedent", "a:\n    b: c\n  d: e\n", "line 3: indented deeper than the mapping"},
		{"sequence at its key's column", "a:\n- b\n", "line 1: no value"},
		{"key with no value", "a:\nb: c\n", "line 1: no value"},
		{"item where a key belongs", "a:\n  b: c\n  - d\n", "line 3: sequence item where a mapping key"},
		{"key where an item belongs", "a:\n  - b\n  c: d\n", "line 3: mapping key where a sequence item"},
		{"not a key", "a b\n", "line 1: \"a b\" is not a \"key: value\" entry"},
		{"YAML flow that is not JSON", "a: {b: c}\n", "line 1: invalid JSON flow value"},
		{"unterminated JSON flow", "a: [1,\n  2\n", "line 1: unterminated JSON flow value"},
		{"closing bracket not deeper than its key", "x:\n  a: [1,\n  ]\n",
			"line 3: a JSON flow value must continue indented deeper"},
		{"inline comment after a plain scalar", "a: b # c\n", "line 1: plain scalar \"b # c\" contains an inline comment"},
		{"inline comment after JSON flow", "a: [1] # c\n", "line 1: text \" # c\" after a JSON flow value"},
		{"inline comment after a quoted scalar", "a: \"b\" # c\n", "line 1: text \" # c\" after a quoted scalar"},
		{"YAML-only escape", `a: "\x41"` + "\n", "line 1: invalid double-quoted scalar"},
		{"multi-line double-quoted scalar", "a: \"b\n  c\"\n", "line 1: invalid double-quoted scalar"},
		{"single-quoted scalar", "a: 'b'\n", "line 1: plain scalar \"'b'\" starts a single-quoted scalar"},
		{"plain scalar with a colon", "a: absent: no filter\n", "line 1: plain scalar \"absent: no filter\" contains \": \""},
		{"plain scalar starting with an indicator", "a: `b`\n", "line 1: plain scalar \"`b`\" starts with the indicator"},
		{"plain scalar starting a sequence", "a: - b\n", "line 1: plain scalar \"- b\" starts with the indicator"},
	}
	for _, tt := range rejects {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseYAMLSubset([]byte(tt.src))
			if err == nil {
				t.Fatalf("parseYAMLSubset(%q) = %v, want an error containing %q", tt.src, got, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("parseYAMLSubset(%q) error = %q, want it to contain %q", tt.src, err, tt.wantErr)
			}
		})
	}
}
