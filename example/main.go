package main

import (
	"fmt"
	"log"
	"net/http"

	. "github.com/danielberigoi/van-go"
)

func pageIndex() string {
	return pageLayout(
		Div()(
			H1()("Index page"),
			P()("This is a paragraph."),
			A(Att{"href": "/test"})("Test page"),
		),
	)
}

func pageTest() string {
	return pageLayout(
		Div()(
			H1()("Test page"),
			P()("This is a paragraph."),
			A(Att{"href": "/"})("Home page"),
		),
	)
}

func pageLayout(content string) string {
	return Html(Att{"data-theme": "dark"})(
		componentHeader(),
		Body()(
			Main(Att{"class": "container"})(
				content,
				Ul()(
					Li()("Item 1"),
					Li()("Item 2"),
					Li()("Item 3"),
				),
			),
		),
	)
}

func componentHeader() string {
	return Head()(
		Title()("Hello Van Go!"),
		Link(Att{"rel": "stylesheet", "href": "https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css"})(),
	)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, pageIndex())
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, pageTest())
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
