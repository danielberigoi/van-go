package cone

import (
	"errors"
	"strings"
	"testing"
)

func render(t *testing.T, n Node) string {
	t.Helper()
	var sb strings.Builder
	if err := n.Render(&sb); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return sb.String()
}

func renderErr(t *testing.T, n Node) error {
	t.Helper()
	err := n.Render(&strings.Builder{})
	if err == nil {
		t.Fatal("expected an error")
	}
	return err
}

func TestRender(t *testing.T) {
	tests := []struct {
		name string
		node Node
		want string
	}{
		{"text is escaped", P()("<script>alert(1)</script>"),
			"<p>&lt;script&gt;alert(1)&lt;/script&gt;</p>"},
		{"attr values are escaped", A(Att{"href": `x" onclick="evil()`})("hi"),
			`<a href="x&#34; onclick=&#34;evil()">hi</a>`},
		{"attrs are sorted", Div(Att{"d": 4, "a": 1, "c": 3, "b": 2})(),
			`<div a="1" b="2" c="3" d="4"></div>`},
		{"later attrs win", Div(Att{"id": "x"}, Att{"id": "y"})(), `<div id="y"></div>`},
		{"children have no separators", Li()(1, 2, "a", "b", 3.5), "<li>12ab3.5</li>"},
		{"bool attrs", Input(Att{"disabled": true, "checked": false, "x": nil})(),
			"<input disabled>"},
		{"void element", Br()(), "<br>"},
		{"html adds doctype", Html()(Body()()), "<!DOCTYPE html><html><body></body></html>"},
		{"nested and slices", Ul()([]Node{Li()("a"), Li()("b")}, []any{nil, "c"}),
			"<ul><li>a</li><li>b</li>c</ul>"},
		{"raw", Div()(Raw("<b>ok</b>")), "<div><b>ok</b></div>"},
		{"group", Group(P()("a"), "b"), "<p>a</p>b"},
		{"custom", Custom("my-el", Att{"x": "1"})("hi"), `<my-el x="1">hi</my-el>`},
		{"script is not escaped", Script()("if (a && b < c) {}"), "<script>if (a && b < c) {}</script>"},
		{"style is not escaped", Style()("a > b {}"), "<style>a > b {}</style>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := render(t, tt.node); got != tt.want {
				t.Errorf("\n got: %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestStringer(t *testing.T) {
	if got := Div()("x").(interface{ String() string }).String(); got != "<div>x</div>" {
		t.Errorf("got %q", got)
	}
}

func TestRenderErrors(t *testing.T) {
	renderErr(t, Custom("bad name")())
	renderErr(t, Div(Att{`onclick="x"`: 1})())
	renderErr(t, Br()("child"))
	renderErr(t, Script()("</script><script>evil()"))
	renderErr(t, Style()("</STYLE>"))
	renderErr(t, Div()(P()(Br()("child"))))
}

func TestScoop(t *testing.T) {
	n := Scoop(func(s *Scope) Node {
		count := s.State("count", 1)
		name := s.State("name", "</script>")
		return Div()(
			Button(On("click", "count.val++"))("+"),
			Span()(count),
			P(Att{"class": Derive("count.val > 1 ? 'big' : ''", "")})(
				Derive("count.val * 2", count.Initial().(int)*2),
			),
			Input(Att{"value": name}, On("input", "name.val = event.target.value"))(),
		)
	})
	want := `<div data-cone-scoop="" style="display:contents">` +
		`<script type="application/json" data-cone-state>{"count":1,"name":"\u003c/script\u003e"}</script>` +
		`<div>` +
		`<button data-cone-on-click="count.val++">+</button>` +
		`<span><span data-cone-text="count.val">1</span></span>` +
		`<p class="" data-cone-attr-class="count.val &gt; 1 ? &#39;big&#39; : &#39;&#39;">` +
		`<span data-cone-text="count.val * 2">2</span></p>` +
		`<input data-cone-on-input="name.val = event.target.value" value="&lt;/script&gt;" data-cone-attr-value="name.val">` +
		`</div></div>`
	if got := render(t, n); got != want {
		t.Errorf("\n got: %s\nwant: %s", got, want)
	}
}

func TestScoopBoolBinding(t *testing.T) {
	n := Scoop(func(s *Scope) Node {
		busy := s.State("busy", false)
		return Button(Att{"disabled": busy})("go")
	})
	if got := render(t, n); !strings.Contains(got, `<button data-cone-attr-disabled="busy.val">`) {
		t.Errorf("false initial value should omit the attribute, got %s", got)
	}
}

func TestStateNames(t *testing.T) {
	for _, name := range []string{"", "1x", "a-b", "class", "event", "delete"} {
		func() {
			defer func() {
				r := recover()
				if err, ok := r.(error); !ok || !errors.Is(err, errBadName) {
					t.Errorf("State(%q): want errBadName panic, got %v", name, r)
				}
			}()
			Scoop(func(s *Scope) Node { s.State(name, 0); return nil })
		}()
	}
	defer func() {
		if recover() == nil {
			t.Error("duplicate state name should panic")
		}
	}()
	Scoop(func(s *Scope) Node { s.State("a", 0); s.State("a", 1); return nil })
}

func TestScoopBadJSON(t *testing.T) {
	renderErr(t, Scoop(func(s *Scope) Node { s.State("f", func() {}); return nil }))
}

func TestRuntime(t *testing.T) {
	got := render(t, Runtime())
	want := `<link crossorigin="anonymous" href="https://cdn.jsdelivr.net/npm/vanjs-core@1.6.1/src/van.js" integrity="sha384-IT8EZ3jBP7WwN5DPEUb9anFeHH3eXUbPj7CUKtwQjavUXhsy7ubIhkk2YFxpZeXb" rel="modulepreload">` +
		`<script type="module">import reactive from "https://cdn.jsdelivr.net/npm/vanjs-core@1.6.1/src/van.js";`
	if !strings.HasPrefix(got, want) || !strings.Contains(got, "window.cone") {
		t.Errorf("unexpected runtime: %.300s", got)
	}
}

func TestRuntimeWith(t *testing.T) {
	got := render(t, RuntimeWith(`/static/van.js"</script>`))
	want := `<script type="module">import reactive from "/static/van.js\"\u003c/script\u003e";`
	if !strings.HasPrefix(got, want) {
		t.Errorf("unexpected runtime: %.200s", got)
	}
}
