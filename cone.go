package cone

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"sort"
	"strings"
)

type Att map[string]any

type TFN func(children ...any) Node

type Node interface {
	Render(w io.Writer) error
}

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

var rawTextElements = map[string]bool{"script": true, "style": true}

var (
	tagNameRe  = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]*$`)
	attrNameRe = regexp.MustCompile(`^[^\s"'<>/=\x00-\x1f\x7f]+$`)
)

type Element struct {
	name     string
	attrs    []Att
	children []any
}

func NewTag(name string, attrs ...Att) TFN {
	return func(children ...any) Node {
		return &Element{name: name, attrs: attrs, children: children}
	}
}

func Custom(name string, attrs ...Att) TFN { return NewTag(name, attrs...) }

func (e *Element) Render(w io.Writer) error {
	sw := &stickyWriter{w: w}
	e.render(sw)
	return sw.err
}

func (e *Element) String() string { return renderString(e) }

func (e *Element) render(w *stickyWriter) {
	if !tagNameRe.MatchString(e.name) {
		w.fail(fmt.Errorf("cone: invalid tag name %q", e.name))
		return
	}
	if e.name == "html" {
		w.str("<!DOCTYPE html>")
	}
	w.str("<" + e.name)
	writeAttrs(w, e.attrs)
	w.str(">")

	if voidElements[e.name] {
		if len(e.children) > 0 {
			w.fail(fmt.Errorf("cone: <%s> is a void element and cannot have children", e.name))
		}
		return
	}

	if rawTextElements[e.name] {
		writeRawText(w, e.name, e.children)
	} else {
		for _, c := range e.children {
			writeChild(w, c)
		}
	}
	w.str("</" + e.name + ">")
}

func writeAttrs(w *stickyWriter, attrs []Att) {
	merged := map[string]any{}
	for _, a := range attrs {
		for k, v := range a {
			merged[k] = v
		}
	}
	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if !attrNameRe.MatchString(k) {
			w.fail(fmt.Errorf("cone: invalid attribute name %q", k))
			return
		}
		v := merged[k]
		if b, ok := v.(binding); ok {
			writeAttrValue(w, k, b.initial())
			w.str(` data-cone-attr-` + strings.ToLower(k) + `="` + html.EscapeString(b.js()) + `"`)
			continue
		}
		writeAttrValue(w, k, v)
	}
}

func writeAttrValue(w *stickyWriter, name string, v any) {
	switch v := v.(type) {
	case nil:
	case bool:
		if v {
			w.str(" " + name)
		}
	default:
		w.str(" " + name + `="` + html.EscapeString(fmt.Sprint(v)) + `"`)
	}
}

func writeChild(w *stickyWriter, c any) {
	switch c := c.(type) {
	case nil:
	case Node:
		if r, ok := c.(interface{ render(*stickyWriter) }); ok {
			r.render(w)
		} else if w.err == nil {
			w.fail(c.Render(w.w))
		}
	case string:
		w.str(html.EscapeString(c))
	case []Node:
		for _, n := range c {
			writeChild(w, n)
		}
	case []any:
		for _, n := range c {
			writeChild(w, n)
		}
	case []string:
		for _, n := range c {
			writeChild(w, n)
		}
	default:
		w.str(html.EscapeString(fmt.Sprint(c)))
	}
}

func writeRawText(w *stickyWriter, tag string, children []any) {
	var sb strings.Builder
	for _, c := range children {
		switch c := c.(type) {
		case nil:
		case string:
			sb.WriteString(c)
		case rawNode:
			sb.WriteString(string(c))
		default:
			sb.WriteString(fmt.Sprint(c))
		}
	}
	text := sb.String()
	// Unescaped on purpose so JS and CSS work; a closing tag would let content break out.
	if strings.Contains(strings.ToLower(text), "</"+tag) {
		w.fail(fmt.Errorf("cone: <%s> content must not contain %q", tag, "</"+tag))
		return
	}
	w.str(text)
}

func Raw(s string) Node { return rawNode(s) }

type rawNode string

func (r rawNode) Render(w io.Writer) error {
	_, err := io.WriteString(w, string(r))
	return err
}
func (r rawNode) String() string { return string(r) }

func Group(children ...any) Node { return group(children) }

type group []any

func (g group) Render(w io.Writer) error {
	sw := &stickyWriter{w: w}
	g.render(sw)
	return sw.err
}
func (g group) render(w *stickyWriter) {
	for _, c := range g {
		writeChild(w, c)
	}
}
func (g group) String() string { return renderString(g) }

type stickyWriter struct {
	w   io.Writer
	err error
}

func (s *stickyWriter) str(v string) {
	if s.err == nil {
		_, s.err = io.WriteString(s.w, v)
	}
}

func (s *stickyWriter) fail(err error) {
	if s.err == nil && err != nil {
		s.err = err
	}
}

func renderString(n Node) string {
	var sb strings.Builder
	if err := n.Render(&sb); err != nil {
		return "<!-- " + strings.ReplaceAll(err.Error(), "--", "- -") + " -->"
	}
	return sb.String()
}
