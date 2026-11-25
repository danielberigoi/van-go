package cone

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
)

func Scoop(build func(s *Scope) Node) Node { return NamedScoop("", build) }

func NamedScoop(name string, build func(s *Scope) Node) Node {
	s := &Scope{states: map[string]*State{}}
	content := build(s)
	return &scoop{name: name, scope: s, content: content}
}

type Scope struct {
	states map[string]*State
}

func (s *Scope) State(name string, initial any) *State {
	if !jsIdentRe.MatchString(name) || jsReserved[name] {
		panic(fmt.Errorf("%w: %q is not a usable JavaScript identifier", errBadName, name))
	}
	if _, dup := s.states[name]; dup {
		panic(fmt.Errorf("%w: %q is declared twice in the same scoop", errBadName, name))
	}
	st := &State{name: name, val: initial}
	s.states[name] = st
	return st
}

type State struct {
	name string
	val  any
}

func (s *State) Name() string { return s.name }

func (s *State) Initial() any { return s.val }

func (s *State) js() string               { return s.name + ".val" }
func (s *State) initial() any             { return s.val }
func (s *State) Render(w io.Writer) error { return renderBinding(w, s) }
func (s *State) render(w *stickyWriter)   { writeBinding(w, s) }
func (s *State) String() string           { return renderString(s) }

func Derive(js string, initial any) *Derived { return &Derived{expr: js, val: initial} }

type Derived struct {
	expr string
	val  any
}

func (d *Derived) js() string               { return d.expr }
func (d *Derived) initial() any             { return d.val }
func (d *Derived) Render(w io.Writer) error { return renderBinding(w, d) }
func (d *Derived) render(w *stickyWriter)   { writeBinding(w, d) }
func (d *Derived) String() string           { return renderString(d) }

type binding interface {
	js() string
	initial() any
}

func writeBinding(w *stickyWriter, b binding) {
	w.str(`<span data-cone-text="` + html.EscapeString(b.js()) + `">`)
	writeChild(w, b.initial())
	w.str("</span>")
}

func renderBinding(w io.Writer, b binding) error {
	sw := &stickyWriter{w: w}
	writeBinding(sw, b)
	return sw.err
}

func On(event, js string) Att { return Att{"data-cone-on-" + event: js} }

type scoop struct {
	name    string
	scope   *Scope
	content Node
}

func (sc *scoop) Render(w io.Writer) error {
	sw := &stickyWriter{w: w}
	sc.render(sw)
	return sw.err
}

func (sc *scoop) String() string { return renderString(sc) }

func (sc *scoop) render(w *stickyWriter) {
	vals := make(map[string]any, len(sc.scope.states))
	for name, st := range sc.scope.states {
		vals[name] = st.val
	}
	// json.Marshal escapes <, > and &, which keeps the state from closing the <script>.
	data, err := json.Marshal(vals)
	if err != nil {
		w.fail(fmt.Errorf("cone: scoop state is not JSON-encodable: %w", err))
		return
	}
	w.str(`<div data-cone-scoop="` + html.EscapeString(sc.name) + `" style="display:contents">`)
	w.str(`<script type="application/json" data-cone-state>` + string(data) + `</script>`)
	writeChild(w, sc.content)
	w.str("</div>")
}

var errBadName = errors.New("cone: invalid state name")

var jsIdentRe = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`)

var jsReserved = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(`
		await break case catch class const continue debugger default delete do
		else enum export extends false finally for function if implements import
		in instanceof interface let new null package private protected public
		return static super switch this throw true try typeof var void while
		with yield arguments eval undefined NaN Infinity
		event`) {
		m[w] = true
	}
	return m
}()

//go:embed js/runtime.js
var runtimeJS string

const (
	cdnURL       = "https://cdn.jsdelivr.net/npm/vanjs-core@1.6.1/src/van.js"
	cdnIntegrity = "sha384-IT8EZ3jBP7WwN5DPEUb9anFeHH3eXUbPj7CUKtwQjavUXhsy7ubIhkk2YFxpZeXb"
)

func Runtime() Node { return runtime(cdnURL, cdnIntegrity) }

func RuntimeWith(src string) Node { return runtime(src, "") }

func runtime(src, integrity string) Node {
	quoted, _ := json.Marshal(src)
	var preload Node
	if integrity != "" {
		// A static import cannot carry an integrity hash; the browser reuses this verified preload for it.
		preload = Link(Att{"rel": "modulepreload", "href": src, "integrity": integrity, "crossorigin": "anonymous"})()
	}
	return Group(preload, Script(Att{"type": "module"})("import reactive from "+string(quoted)+";\n"+runtimeJS))
}
