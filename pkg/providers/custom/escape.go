package custom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

// jsonEscapeFunc is the template function added to sprig actions that write
// inside a JSON string.
const jsonEscapeFunc = "notifyJSONEscape"

// isJSONBody reports whether the request body is JSON: a JSON Content-Type
// header, or a format that starts like a JSON object or array.
func isJSONBody(headers map[string]string, format string) bool {
	for k, v := range headers {
		if strings.EqualFold(k, "Content-Type") && strings.Contains(strings.ToLower(v), "json") {
			return true
		}
	}
	format = strings.TrimSpace(format)
	return strings.HasPrefix(format, "{") || strings.HasPrefix(format, "[")
}

// escapeJSONString returns v escaped for use inside a JSON string, without
// the quotes.
func escapeJSONString(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		s = fmt.Sprint(v)
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	out := strings.TrimSuffix(buf.String(), "\n")
	return out[1 : len(out)-1]
}

// inJSONString scans s and reports whether its end is inside a JSON string.
// inString is the state at the start of s.
func inJSONString(s string, inString bool) bool {
	escaped := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case !inString:
			if c == '"' {
				inString = true
			}
		case escaped:
			escaped = false
		case c == '\\':
			escaped = true
		case c == '"':
			inString = false
		}
	}
	return inString
}

// replaceDataJSON replaces each {{data}} in format with msg. An occurrence
// inside a JSON string gets msg escaped, so the input line cannot close the
// string and add fields. An occurrence outside a string (for example a
// format of just {{data}} that forwards JSON lines) stays raw.
func replaceDataJSON(format, msg string) string {
	const placeholder = "{{data}}"
	parts := strings.Split(format, placeholder)
	var b strings.Builder
	inString := false
	for i, part := range parts {
		b.WriteString(part)
		inString = inJSONString(part, inString)
		if i == len(parts)-1 {
			break
		}
		if inString {
			b.WriteString(escapeJSONString(msg))
		} else {
			b.WriteString(msg)
			inString = inJSONString(msg, inString)
		}
	}
	return b.String()
}

// escapeTemplateJSONStrings makes every action of tmpl that writes inside a
// JSON string pass its value through jsonEscapeFunc, as html/template does
// for HTML contexts. tmpl must have jsonEscapeFunc in its function map.
func escapeTemplateJSONStrings(tmpl *template.Template) {
	if tmpl.Tree == nil {
		return
	}
	escapeJSONList(tmpl.Root, false)
}

func escapeJSONList(list *parse.ListNode, inString bool) bool {
	if list == nil {
		return inString
	}
	for _, node := range list.Nodes {
		switch n := node.(type) {
		case *parse.TextNode:
			inString = inJSONString(string(n.Text), inString)
		case *parse.ActionNode:
			if inString && len(n.Pipe.Decl) == 0 {
				n.Pipe.Cmds = append(n.Pipe.Cmds, &parse.CommandNode{
					NodeType: parse.NodeCommand,
					Args:     []parse.Node{parse.NewIdentifier(jsonEscapeFunc).SetTree(nil).SetPos(n.Pos)},
				})
			}
		case *parse.IfNode:
			inString = escapeJSONBranch(&n.BranchNode, inString)
		case *parse.RangeNode:
			inString = escapeJSONBranch(&n.BranchNode, inString)
		case *parse.WithNode:
			inString = escapeJSONBranch(&n.BranchNode, inString)
		}
	}
	return inString
}

func escapeJSONBranch(n *parse.BranchNode, inString bool) bool {
	after := escapeJSONList(n.List, inString)
	escapeJSONList(n.ElseList, inString)
	return after
}
