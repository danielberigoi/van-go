![Self-Portrait with Ice Cream](https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTtHIZV8z8Devr_vFmzk8EoIkX8g3C4pgoeLg&s)

*Self-Portrait with Ice Cream* — Vincent van Cone, 2024

## HTML, best served cold.

[VanJS](https://github.com/vanjs-org/van) but in golang. HTML is frozen on the server in Go, then melted in the browser into something interactive. No Go dependencies, no JS build step.

## Installation
```sh
go get github.com/danielberigoi/van-go
```
```go
import "github.com/danielberigoi/van-go" // package cone
```

## Freezing HTML
```go
import . "github.com/danielberigoi/van-go"

http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    Html()(
        Head()(Title()("Hello Van Go!")),
        Body(Att{"class": "container"})(
            H1()("Hello, ", r.URL.Query().Get("name")),
        ),
    ).Render(w)
})
```

- Text and attribute values are HTML-escaped. Use `Raw(...)` for trusted HTML.
- Attributes are rendered in sorted order, so output is deterministic.
- `true` renders a bare boolean attribute, and `false` or `nil` leaves it out.
- `Html` adds `<!DOCTYPE html>`. Void elements render as `<br>`.
- Nodes stream to an `io.Writer` with `Render(w)` and also implement `fmt.Stringer`.
- All HTML elements have a function, and `Custom("my-el")` covers anything else.
- `<script>` and `<style>` content is written unescaped. A closing tag inside it is rejected.

## Scoops: the parts that melt
```go
func counter(start int) Node {
    return Scoop(func(s *Scope) Node {
        count := s.State("count", start)
        return Div()(
            Button(On("click", "count.val--"))("-"),
            Span()(count),
            Button(On("click", "count.val++"))("+"),
            P()("Doubled: ", Derive("count.val * 2", start*2)),
        )
    })
}

// once per page, at the end of <body>
Body()(counter(0), counter(10), Runtime())
```

The server sends the full HTML with the initial values, so the page works and is indexable before any JS runs. `Runtime()` loads VanJS 1.6.1 (5KB) from jsDelivr, pinned and checked with a Subresource Integrity hash, plus a small inline runtime. Together they turn each scoop's states into reactive states and bind them to the DOM.

To avoid the CDN (offline, intranet, strict CSP), serve your own copy of [VanJS](https://github.com/vanjs-org/van) and use `RuntimeWith("/static/van-1.6.1.js")` instead.

| Go | Browser |
|---|---|
| `s.State("count", 0)` | `count`, a reactive state; read and write it with `count.val` |
| `Span()(count)` | text that follows `count.val` |
| `Att{"value": name}` | an attribute or property that follows `name.val` |
| `Derive("count.val * 2", 0)` | a computed value; the second argument is its server-side initial value |
| `On("input", "name.val = event.target.value")` | an event handler; the DOM event is `event` |

For anything else, name the scoop and register a JS hook:

```go
NamedScoop("clock", func(s *Scope) Node { ... })
```
```html
<script type="module">
  cone.register("clock", (root, { time }) => {
    setInterval(() => time.val = new Date().toLocaleTimeString(), 1000)
  })
</script>
```

After inserting server-rendered HTML (for example with htmx), call `cone.hydrateAll(node)`.

### Limitations
- Expressions are compiled with `new Function`, so a Content Security Policy needs `'unsafe-eval'`.
- A state or `Derive` used as a child is wrapped in a `<span>`.
- `Derive` needs its initial value computed in Go, because the server does not run JS. When the browser's first value differs from what the server rendered, the runtime logs a `cone:` warning naming the expression and element.
- There is no list rendering yet.

See [`example/`](example) for a runnable demo (`go run ./example`).

Van Gogh sold one painting in his lifetime. Star this repo and beat him.
