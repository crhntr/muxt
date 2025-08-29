package dom

import (
	"iter"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/crhntr/dom/spec"
)

type Document struct {
	node *html.Node
}

func (d *Document) Head() spec.Element { return d.QuerySelector("head") }
func (d *Document) Body() spec.Element { return d.QuerySelector("body") }

func (d *Document) String() string                  { return outerHTML(d.node) }
func (d *Document) NodeType() spec.NodeType         { return nodeType(d.node.Type) }
func (d *Document) CloneNode(deep bool) spec.Node   { return NewNode(cloneNode(d.node, deep)) }
func (d *Document) IsSameNode(other spec.Node) bool { return isSameNode(d.node, other) }
func (d *Document) GetElementsByTagName(name string) spec.ElementCollection {
	return getElementsByTagName(d.node, name)
}

func (d *Document) GetElementsByClassName(name string) spec.ElementCollection {
	return getElementsByClassName(d.node, name)
}

func (d *Document) QuerySelector(query string) spec.Element {
	return querySelector(d.node, query, false)
}

func (d *Document) QuerySelectorAll(query string) spec.NodeList[spec.Element] {
	return querySelectorAll(d.node, query, false)
}

func (d *Document) QuerySelectorSequence(query string) iter.Seq[spec.Element] {
	m := cascadia.MustCompile(query)
	return func(yield func(spec.Element) bool) {
		querySelectorSequence(d.node, m, yield)
	}
}

func (d *Document) Contains(other spec.Node) bool { return contains(d.node, other) }

func (d *Document) CompareDocumentPosition(other spec.Node) spec.DocumentPosition {
	return compareDocumentPosition(d.node, other)
}

// TextContent returns an empty string.
// The spec says it should return null
// https://developer.mozilla.org/en-US/docs/Web/API/Node/textContent
func (d *Document) TextContent() string { return "" }

// Document

func (*Document) CreateElement(localName string) spec.Element {
	localName = strings.ToLower(localName)
	return &Element{
		node: &html.Node{
			DataAtom: atom.Lookup([]byte(localName)),
			Type:     html.ElementNode,
			Data:     localName,
		},
	}
}

func (*Document) CreateElementIs(localName, is string) spec.Element {
	localName = strings.ToLower(localName)
	return &Element{
		node: &html.Node{
			DataAtom: atom.Lookup([]byte(localName)),
			Type:     html.ElementNode,
			Data:     localName,
			Attr:     []html.Attribute{{Key: "is", Val: is}},
		},
	}
}

func (*Document) CreateTextNode(text string) spec.Text {
	return &Text{
		node: &html.Node{
			Type: html.TextNode,
			Data: text,
		},
	}
}
