package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	. "github.com/danielberigoi/van-go"
)

func counter(label string, start int) Node {
	return Scoop(func(s *Scope) Node {
		count := s.State("count", start)
		return Article()(
			H3()(label),
			Button(On("click", "count.val--"))("−"),
			Strong(Att{"style": "margin: 0 1rem"})(count),
			Button(On("click", "count.val++"))("+"),
			P(Att{"class": Derive("count.val >= 10 ? 'pico-color-red-500' : ''", "")})(
				"Doubled: ", Derive("count.val * 2", start*2),
			),
		)
	})
}

func greeter(name string) Node {
	greeting := strings.TrimSpace(name)
	if greeting == "" {
		greeting = "stranger"
	}
	return Scoop(func(s *Scope) Node {
		who := s.State("who", name)
		return Article()(
			Label()("Your name",
				Input(Att{"value": who}, On("input", "who.val = event.target.value"))(),
			),
			P()("Hello, ", Derive("who.val.trim() || 'stranger'", greeting), "!"),
			Button(Att{"disabled": Derive("!who.val", name == "")}, On("click", "who.val = ''"))("Clear"),
		)
	})
}

func clock(now string) Node {
	return NamedScoop("clock", func(s *Scope) Node {
		time := s.State("time", now)
		return P()("It's ", Time()(time), " UTC")
	})
}

func page(title string, content ...any) Node {
	return Html(Att{"lang": "en", "data-theme": "dark"})(
		Head()(
			Meta(Att{"charset": "utf-8"})(),
			Meta(Att{"name": "viewport", "content": "width=device-width, initial-scale=1"})(),
			Title()(title),
			Link(Att{"rel": "stylesheet", "href": "https://cdn.jsdelivr.net/npm/@picocss/pico@2.0.6/css/pico.min.css"})(),
			Link(Att{"rel": "stylesheet", "href": "https://cdn.jsdelivr.net/npm/@picocss/pico@2.0.6/css/pico.colors.min.css"})(),
		),
		Body()(
			Main(Att{"class": "container"})(
				Nav()(Ul()(Li()(A(Att{"href": "/"})("Home")), Li()(A(Att{"href": "/static"})("Static page")))),
				content,
			),
			Runtime(),
			Script(Att{"type": "module"})(`
				cone.register("clock", (root, {time}) => {
					setInterval(() => time.val = new Date().toISOString().slice(11, 19), 1000)
				})
			`),
		),
	)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		err := page("van-go",
			H1()("Vincent van Cone"),
			P()("Frozen in Go, melted in your browser."),
			Div(Att{"class": "grid"})(counter("Apples", 0), counter("Pears", 8)),
			greeter(name),
			clock(time.Now().UTC().Format("15:04:05")),
		).Render(w)
		if err != nil {
			log.Print(err)
		}
	})

	http.HandleFunc("/static", func(w http.ResponseWriter, r *http.Request) {
		page("Static", H1()("No scoops here"), P()("Plain frozen HTML.")).Render(w)
	})

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
