package van_go

import (
	"fmt"
	"slices"
	"strings"
)

type Att map[string]interface{}
type TFN func(...interface{}) string

var singleTagNames = []string{"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr"}

type Tag struct {
	name  string
	attrs []Att
}

func (t *Tag) render(children ...interface{}) string {
	var _attrs strings.Builder

	for _, attr := range t.attrs {
		for key, value := range attr {
			_attrs.WriteString(fmt.Sprintf(" %s=\"%v\"", key, value))
		}
	}

	if slices.Contains(singleTagNames, t.name) {
		return fmt.Sprintf("<%s%s/>", t.name, _attrs.String())
	}

	return fmt.Sprintf("<%s%s>%s</%s>", t.name, _attrs.String(), fmt.Sprint(children...), t.name)
}

func NewTag(name string, attrs ...Att) *Tag {
	return &Tag{name: name, attrs: attrs}
}

func A(attrs ...Att) TFN      { return NewTag("a", attrs...).render }
func Body(attrs ...Att) TFN   { return NewTag("body", attrs...).render }
func Button(attrs ...Att) TFN { return NewTag("button", attrs...).render }

func Div(attrs ...Att) TFN    { return NewTag("div", attrs...).render }
func Form(attrs ...Att) TFN   { return NewTag("form", attrs...).render }
func H1(attrs ...Att) TFN     { return NewTag("h1", attrs...).render }
func H2(attrs ...Att) TFN     { return NewTag("h2", attrs...).render }
func H3(attrs ...Att) TFN     { return NewTag("h3", attrs...).render }
func H4(attrs ...Att) TFN     { return NewTag("h4", attrs...).render }
func H5(attrs ...Att) TFN     { return NewTag("h5", attrs...).render }
func H6(attrs ...Att) TFN     { return NewTag("h6", attrs...).render }
func Head(attrs ...Att) TFN   { return NewTag("head", attrs...).render }
func Html(attrs ...Att) TFN   { return NewTag("html", attrs...).render }
func IFrame(attrs ...Att) TFN { return NewTag("iframe", attrs...).render }
func Img(attrs ...Att) TFN    { return NewTag("img", attrs...).render }
func Input(attrs ...Att) TFN  { return NewTag("input", attrs...).render }
func Label(attrs ...Att) TFN  { return NewTag("label", attrs...).render }
func Li(attrs ...Att) TFN     { return NewTag("li", attrs...).render }
func Link(attrs ...Att) TFN   { return NewTag("link", attrs...).render }
func Main(attrs ...Att) TFN   { return NewTag("main", attrs...).render }
func Option(attrs ...Att) TFN { return NewTag("option", attrs...).render }
func P(attrs ...Att) TFN      { return NewTag("p", attrs...).render }
func Script(attrs ...Att) TFN { return NewTag("script", attrs...).render }
func Select(attrs ...Att) TFN { return NewTag("select", attrs...).render }
func Span(attrs ...Att) TFN   { return NewTag("span", attrs...).render }
func Style(attrs ...Att) TFN  { return NewTag("style", attrs...).render }
func Table(attrs ...Att) TFN  { return NewTag("table", attrs...).render }
func Tbody(attrs ...Att) TFN  { return NewTag("tbody", attrs...).render }
func Td(attrs ...Att) TFN     { return NewTag("td", attrs...).render }
func Tfoot(attrs ...Att) TFN  { return NewTag("tfoot", attrs...).render }
func Th(attrs ...Att) TFN     { return NewTag("th", attrs...).render }
func Thead(attrs ...Att) TFN  { return NewTag("thead", attrs...).render }
func Title(attrs ...Att) TFN  { return NewTag("title", attrs...).render }
func Tr(attrs ...Att) TFN     { return NewTag("tr", attrs...).render }
func Ul(attrs ...Att) TFN     { return NewTag("ul", attrs...).render }

func Custom(name string, attrs ...Att) TFN { return NewTag(name, attrs...).render }
