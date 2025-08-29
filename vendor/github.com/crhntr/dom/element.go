package dom

import (
	"bytes"
	"iter"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"

	"github.com/crhntr/dom/spec"
)

type Element struct {
	node *html.Node
}

func (e *Element) QuerySelectorSequence(query string) iter.Seq[spec.Element] {
	m := cascadia.MustCompile(query)
	return func(yield func(spec.Element) bool) {
		querySelectorSequence(e.node, m, yield)
	}
}

// NewNode

func (e *Element) NodeType() spec.NodeType         { return nodeType(e.node.Type) }
func (e *Element) IsConnected() bool               { return isConnected(e.node) }
func (e *Element) OwnerDocument() spec.Document    { return ownerDocument(e.node) }
func (e *Element) ParentNode() spec.Node           { return parentNode(e.node) }
func (e *Element) ParentElement() spec.Element     { return parentElement(e.node) }
func (e *Element) PreviousSibling() spec.ChildNode { return previousSibling(e.node) }
func (e *Element) NextSibling() spec.ChildNode     { return nextSibling(e.node) }
func (e *Element) TextContent() string             { return textContent(e.node) }
func (e *Element) CloneNode(deep bool) spec.Node   { return NewNode(cloneNode(e.node, deep)) }
func (e *Element) IsSameNode(other spec.Node) bool { return isSameNode(e.node, other) }
func (e *Element) Length() int {
	c := e.node.FirstChild
	result := 0
	for c != nil {
		result++
		c = c.NextSibling
	}
	return result
}

func (e *Element) CompareDocumentPosition(other spec.Node) spec.DocumentPosition {
	return compareDocumentPosition(e.node, other)
}

// ParentNode

func (e *Element) Children() spec.ElementCollection   { return children(e.node) }
func (e *Element) FirstElementChild() spec.Element    { return firstElementChild(e.node) }
func (e *Element) LastElementChild() spec.Element     { return lastElementChild(e.node) }
func (e *Element) ChildElementCount() int             { return childElementCount(e.node) }
func (e *Element) Prepend(nodes ...spec.Node)         { prependNodes(e.node, nodes) }
func (e *Element) Append(nodes ...spec.Node)          { appendNodes(e.node, nodes...) }
func (e *Element) ReplaceChildren(nodes ...spec.Node) { replaceChildren(e.node, nodes) }
func (e *Element) GetElementsByTagName(name string) spec.ElementCollection {
	return getElementsByTagName(e.node, name)
}

func (e *Element) GetElementsByClassName(name string) spec.ElementCollection {
	return getElementsByClassName(e.node, name)
}

func (e *Element) QuerySelector(query string) spec.Element {
	return querySelector(e.node, query, false)
}

func (e *Element) QuerySelectorAll(query string) spec.NodeList[spec.Element] {
	return querySelectorAll(e.node, query, false)
}
func (e *Element) Closest(selector string) spec.Element { return closest(e.node, selector) }
func (e *Element) Matches(selector string) bool         { return matches(e.node, selector) }

func (e *Element) HasChildNodes() bool                  { return hasChildNodes(e.node) }
func (e *Element) ChildNodes() spec.NodeList[spec.Node] { return childNodes(e.node) }
func (e *Element) FirstChild() spec.ChildNode           { return firstChild(e.node) }
func (e *Element) LastChild() spec.ChildNode            { return lastChild(e.node) }
func (e *Element) Contains(other spec.Node) bool        { return contains(e.node, other) }
func (e *Element) InsertBefore(node, child spec.ChildNode) spec.ChildNode {
	return insertBefore(e.node, node, child)
}
func (e *Element) AppendChild(node spec.ChildNode) spec.ChildNode { return appendChild(e.node, node) }
func (e *Element) ReplaceChild(node, child spec.ChildNode) spec.ChildNode {
	return replaceChild(e.node, node, child)
}
func (e *Element) RemoveChild(node spec.ChildNode) spec.ChildNode { return removeChild(e.node, node) }

// Element

func (e *Element) TagName() string                 { return strings.ToUpper(e.node.Data) }
func (e *Element) ID() string                      { return getAttribute(e.node, "id") }
func (e *Element) ClassName() string               { return getAttribute(e.node, "class") }
func (e *Element) GetAttribute(name string) string { return getAttribute(e.node, name) }

func (e *Element) SetAttribute(name, value string) {
	name = strings.ToLower(name)
	for index, att := range e.node.Attr {
		if att.Key == name {
			e.node.Attr[index].Val = value
		}
	}
	e.node.Attr = append(e.node.Attr, html.Attribute{
		Key: name, Val: value,
	})
}

func (e *Element) RemoveAttribute(name string) {
	name = strings.ToLower(name)
	filtered := e.node.Attr[:0]
	for _, att := range e.node.Attr {
		if att.Key == name {
			continue
		}
		filtered = append(filtered, att)
	}
	e.node.Attr = filtered
}

func (e *Element) ToggleAttribute(name string) bool {
	name = strings.ToLower(name)
	if e.HasAttribute(name) {
		e.RemoveAttribute(name)
		return false
	}
	e.SetAttribute(name, "")
	return true
}

func (e *Element) HasAttribute(name string) bool {
	name = strings.ToLower(name)
	for _, att := range e.node.Attr {
		if att.Key == name {
			return true
		}
	}
	return false
}

func (e *Element) SetInnerHTML(s string) {
	nodes, err := html.ParseFragment(strings.NewReader(s), &html.Node{Type: html.ElementNode})
	if err != nil {
		panic(err)
	}
	clearChildren(e.node)
	for _, n := range nodes {
		e.node.AppendChild(n)
	}
}

func (e *Element) InnerHTML() string {
	var buf bytes.Buffer
	c := e.node.FirstChild
	for c != nil {
		err := html.Render(&buf, c)
		if err != nil {
			panic(err)
		}
		c = c.NextSibling
	}
	return buf.String()
}

func (e *Element) SetOuterHTML(s string) {
	nodes, err := html.ParseFragment(strings.NewReader(s), &html.Node{Type: html.ElementNode})
	if err != nil {
		panic(err)
	}
	if len(nodes) == 0 {
		return
	}
	if e.node.Parent == nil {
		panic("browser: SetOuterHTML called on an unattached node")
	}
	for _, node := range nodes {
		e.node.Parent.InsertBefore(node, e.node)
	}
	e.node.Parent.RemoveChild(e.node)
}

func (e *Element) OuterHTML() string { return outerHTML(e.node) }
func (e *Element) String() string    { return e.OuterHTML() }

type siblingElements struct {
	firstChild *html.Node
}

func (list siblingElements) Length() int {
	result := 0
	for c := list.firstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		result++
	}
	return result
}

func (list siblingElements) Item(index int) spec.Element {
	childIndex := 0
	for c := list.firstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if childIndex == index {
			return &Element{node: c}
		}
		childIndex++
	}
	return nil
}

func (list siblingElements) NamedItem(name string) spec.Element {
	for c := list.firstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		if isNamed(c, name) {
			return &Element{node: c}
		}
	}
	return nil
}

type elementList []*html.Node

func (list elementList) Length() int { return len(list) }

func (list elementList) Item(index int) spec.Element {
	if index < 0 || index >= len(list) {
		return nil
	}
	return &Element{node: list[index]}
}

func (list elementList) NamedItem(name string) spec.Element {
	for _, el := range list {
		if isNamed(el, name) {
			return &Element{node: el}
		}
	}
	return nil
}
