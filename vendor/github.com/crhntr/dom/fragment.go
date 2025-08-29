package dom

import (
	"bytes"
	"iter"
	"slices"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"

	"github.com/crhntr/dom/spec"
)

type DocumentFragment struct {
	nodes []*html.Node
}

func NewDocumentFragment(nodes []*html.Node) *DocumentFragment {
	return &DocumentFragment{nodes: nodes}
}

func (d *DocumentFragment) String() string { return outerHTML(d.nodes...) }

func (d *DocumentFragment) NodeType() spec.NodeType { return spec.NodeTypeDocumentFragment }

func (d *DocumentFragment) CloneNode(deep bool) spec.Node {
	if !deep {
		return &DocumentFragment{nodes: d.nodes}
	}
	df := &DocumentFragment{nodes: make([]*html.Node, len(d.nodes))}
	for _, e := range d.nodes {
		df.nodes = append(df.nodes, cloneNode(e, deep))
	}
	return df
}

func (d *DocumentFragment) IsSameNode(other spec.Node) bool {
	o, ok := other.(*DocumentFragment)
	if !ok {
		return false
	}
	return d == o
}

func (d *DocumentFragment) CompareDocumentPosition(other spec.Node) spec.DocumentPosition {
	return compareDocumentFragmentPosition(d.nodes, other)
}

func (d *DocumentFragment) TextContent() string {
	var buf bytes.Buffer
	for _, n := range d.nodes {
		recursiveTextContent(&buf, n)
	}
	return buf.String()
}

func (d *DocumentFragment) Children() spec.ElementCollection {
	elementChildren := make(elementList, 0, len(d.nodes))
	for _, n := range d.nodes {
		if n != nil && n.Type == html.ElementNode {
			elementChildren = append(elementChildren, n)
		}
	}
	return slices.Clip(elementChildren)
}

func (d *DocumentFragment) FirstElementChild() spec.Element {
	for _, n := range d.nodes {
		if n.Type == html.ElementNode {
			return &Element{node: n}
		}
	}
	return nil
}

func (d *DocumentFragment) LastElementChild() spec.Element {
	for i := range d.nodes {
		n := d.nodes[len(d.nodes)-1-i]
		if n.Type == html.ElementNode {
			return &Element{node: n}
		}
	}
	return nil
}

func (d *DocumentFragment) ChildElementCount() int {
	count := 0
	for _, n := range d.nodes {
		if n.Type == html.ElementNode {
			count++
		}
	}
	return count
}

func (d *DocumentFragment) Append(nodes ...spec.Node) {
	d.nodes = slices.Grow(d.nodes, len(nodes))
	for _, node := range nodes {
		d.nodes = append(d.nodes, domNodeToHTMLNode(node))
	}
}

func (d *DocumentFragment) Prepend(nodes ...spec.Node) {
	children := make([]*html.Node, 0, len(d.nodes)+len(nodes))
	for _, node := range nodes {
		children = append(children, domNodeToHTMLNode(node))
	}
	d.nodes = append(children, d.nodes...)
}

func (d *DocumentFragment) ReplaceChildren(nodes ...spec.Node) {
	list := make([]*html.Node, 0, len(nodes))
	for _, node := range nodes {
		list = append(list, domNodeToHTMLNode(node))
	}
	d.nodes = list
}

func (d *DocumentFragment) QuerySelector(query string) spec.Element {
	for _, n := range d.nodes {
		el := querySelector(n, query, true)
		if el != nil {
			return el
		}
	}
	return nil
}

func (d *DocumentFragment) QuerySelectorAll(query string) spec.NodeList[spec.Element] {
	var list nodeListHTMLElements
	for _, n := range d.nodes {
		list = append(list, querySelectorAll(n, query, true)...)
	}
	return slices.Clip(list)
}

func (d *DocumentFragment) QuerySelectorSequence(query string) iter.Seq[spec.Element] {
	m := cascadia.MustCompile(query)
	return func(yield func(spec.Element) bool) {
		for _, n := range d.nodes {
			if m.Match(n) {
				if !yield(&Element{node: n}) {
					return
				}
			}
			if !querySelectorSequence(n, m, yield) {
				return
			}
		}
	}
}
