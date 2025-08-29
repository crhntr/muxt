package dom

import (
	"bytes"
	"io"
	"slices"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"

	"github.com/crhntr/dom/spec"
)

func outerHTML(nodes ...*html.Node) string {
	var buf bytes.Buffer
	for _, node := range nodes {
		if err := html.Render(&buf, node); err != nil {
			return ""
		}
	}
	return buf.String()
}

func nodeType(nodeType html.NodeType) spec.NodeType {
	switch nodeType {
	case html.TextNode:
		return spec.NodeTypeText
	case html.DocumentNode:
		return spec.NodeTypeDocument
	case html.ElementNode:
		return spec.NodeTypeElement
	case html.CommentNode:
		return spec.NodeTypeComment
	case html.DoctypeNode:
		return spec.NodeTypeDocumentType
	default:
		fallthrough
	case html.ErrorNode, html.RawNode:
		return spec.NodeTypeUnknown
	}
}

func NewNode(node *html.Node) spec.Node {
	if node == nil {
		return nil
	}
	switch node.Type {
	case html.ElementNode:
		return &Element{node: node}
	case html.TextNode:
		return &Text{node: node}
	case html.DocumentNode:
		return &Document{node: node}
	default:
		panic("not supported")
	}
}

func htmlNodeToDomChildNode(node *html.Node) spec.ChildNode {
	if node == nil {
		return nil
	}
	switch node.Type {
	case html.ElementNode:
		return &Element{node: node}
	case html.TextNode:
		return &Text{node: node}
	default:
		panic("not supported")
	}
}

func htmlNodeToDomElement(node *html.Node) spec.Element {
	if node == nil {
		return nil
	}
	return &Element{node: node}
}

func domNodeToHTMLNode(node spec.Node) *html.Node {
	switch ot := node.(type) {
	case *Element:
		return ot.node
	case *Text:
		return ot.node
	case *Document:
		return ot.node
	default:
		panic("not implemented")
	}
}

func walkNodes(start *html.Node, fn func(node *html.Node) (done bool)) bool {
	if fn(start) {
		return true
	}

	c := start.FirstChild
	for c != nil {
		if walkNodes(c, fn) {
			return true
		}
		c = c.NextSibling
	}

	return false
}

type firstChildIterator html.Node

func (node *firstChildIterator) Length() int {
	c := (*html.Node)(node)
	result := 0
	for c != nil {
		result++
		c = c.NextSibling
	}
	return result
}

func (node *firstChildIterator) Item(index int) spec.Node {
	c := (*html.Node)(node)
	offset := 0
	for c != nil {
		if offset == index {
			return NewNode(c)
		}
		offset++
		c = c.NextSibling
	}
	return nil
}

func isConnected(node *html.Node) bool {
	p := node.Parent
	for p != nil {
		if p.Type == html.DocumentNode {
			return true
		}
		p = p.Parent
	}
	return false
}

func ownerDocument(node *html.Node) spec.Document {
	n := ownerDocumentNode(node)
	if n == nil {
		return nil
	}
	return &Document{node: n}
}

func ownerDocumentNode(node *html.Node) *html.Node {
	p := node.Parent
	for p != nil {
		if p.Type == html.DocumentNode {
			return p
		}
		p = p.Parent
	}
	return nil
}
func parentNode(node *html.Node) spec.Node       { return NewNode(node.Parent) }
func parentElement(node *html.Node) spec.Element { return htmlNodeToDomElement(node.Parent) }
func hasChildNodes(node *html.Node) bool         { return node.FirstChild != nil }
func childNodes(node *html.Node) spec.NodeList[spec.Node] {
	return (*firstChildIterator)(node.FirstChild)
}
func firstChild(node *html.Node) spec.ChildNode      { return htmlNodeToDomChildNode(node.FirstChild) }
func lastChild(node *html.Node) spec.ChildNode       { return htmlNodeToDomChildNode(node.LastChild) }
func previousSibling(node *html.Node) spec.ChildNode { return htmlNodeToDomChildNode(node.PrevSibling) }
func nextSibling(node *html.Node) spec.ChildNode     { return htmlNodeToDomChildNode(node.NextSibling) }

func textContent(node *html.Node) string {
	var buf bytes.Buffer
	recursiveTextContent(&buf, node)
	return buf.String()
}

func recursiveTextContent(sw io.StringWriter, n *html.Node) {
	if n.Type == html.TextNode {
		_, err := sw.WriteString(n.Data)
		if err != nil {
			panic(err)
		}
	}
	c := n.FirstChild
	for c != nil {
		recursiveTextContent(sw, c)
		c = c.NextSibling
	}
}

func cloneNode(node *html.Node, deep bool) *html.Node {
	result := &html.Node{
		Type:      node.Type,
		Namespace: node.Namespace,
		Data:      node.Data,
		DataAtom:  node.DataAtom,
	}
	if deep {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			result.AppendChild(cloneNode(c, true))
		}
	}
	if node.Attr != nil {
		result.Attr = make([]html.Attribute, len(node.Attr))
		for i, at := range node.Attr {
			result.Attr[i].Key = at.Key
			result.Attr[i].Val = at.Val
			result.Attr[i].Namespace = at.Namespace
		}
	}
	return result
}

func isSameNode(node *html.Node, other spec.Node) bool {
	if node == nil || other == nil {
		return false
	}
	n := domNodeToHTMLNode(other)
	return n != nil && node == n
}

func contains(node *html.Node, other spec.Node) bool {
	o := domNodeToHTMLNode(other)
	if o == nil {
		return false
	}

	found := false
	walkNodes(node, func(n *html.Node) bool {
		found = n == o
		return found
	})

	return found
}

func insertBefore(parent *html.Node, node, child spec.ChildNode) spec.ChildNode {
	n := domNodeToHTMLNode(node)
	c := domNodeToHTMLNode(child)
	if n.Parent != nil {
		n.Parent.RemoveChild(n)
	}
	parent.InsertBefore(n, c)
	return htmlNodeToDomChildNode(n)
}

func appendChild(parent *html.Node, node spec.ChildNode) spec.ChildNode {
	n := domNodeToHTMLNode(node)
	if n.Parent != nil {
		n.Parent.RemoveChild(n)
	}
	parent.AppendChild(n)
	return htmlNodeToDomChildNode(n)
}

func replaceChild(parent *html.Node, node, child spec.ChildNode) spec.ChildNode {
	n := domNodeToHTMLNode(node)
	c := domNodeToHTMLNode(child)
	if c.Parent != parent {
		panic("browser: ReplaceChild called for an attached child node")
	}
	if c.PrevSibling != nil {
		c.PrevSibling.NextSibling = n
	}
	if c.NextSibling != nil {
		c.NextSibling.PrevSibling = n
	}
	if parent.FirstChild == c {
		parent.FirstChild = n
	}
	if parent.LastChild == c {
		parent.LastChild = n
	}
	n.PrevSibling = c.PrevSibling
	n.NextSibling = c.NextSibling
	n.Parent = c.Parent

	c.PrevSibling = nil
	c.NextSibling = nil
	c.Parent = nil

	return htmlNodeToDomChildNode(c)
}

func removeChild(parent *html.Node, node spec.ChildNode) spec.ChildNode {
	n := domNodeToHTMLNode(node)
	parent.RemoveChild(n)
	return htmlNodeToDomChildNode(n)
}

func children(parent *html.Node) spec.ElementCollection {
	return siblingElements{firstChild: parent.FirstChild}
}

func firstElementChild(node *html.Node) spec.Element {
	child := node.FirstChild
	for child != nil {
		if child.Type == html.ElementNode {
			return &Element{node: child}
		}
		child = child.NextSibling
	}
	return nil
}

func lastElementChild(node *html.Node) spec.Element {
	child := node.LastChild
	for child != nil {
		if child.Type == html.ElementNode {
			return &Element{node: child}
		}
		child = child.PrevSibling
	}
	return nil
}

func childElementCount(node *html.Node) int {
	var (
		result = 0
		child  = node.FirstChild
	)
	for child != nil {
		if child.Type == html.ElementNode {
			result++
		}
		child = child.NextSibling
	}
	return result
}

func prependNodes(node *html.Node, nodes []spec.Node) {
	for i := range nodes {
		dn := nodes[len(nodes)-1-i]
		if fragment, ok := dn.(*DocumentFragment); ok {
			for _, n := range fragment.nodes {
				prependHTMLNode(node, n)
			}
			continue
		}
		n := domNodeToHTMLNode(dn)
		prependHTMLNode(node, n)
	}
}

func prependHTMLNode(node *html.Node, n *html.Node) {
	fc := node.FirstChild
	if fc != nil {
		fc.PrevSibling = n
		n.NextSibling = fc
	}
	n.Parent = node
	node.FirstChild = n

	if node.LastChild == nil {
		node.LastChild = n
	}
}

func appendNodes(parent *html.Node, nodes ...spec.Node) {
	for _, node := range nodes {
		if fragment, ok := node.(*DocumentFragment); ok {
			for _, n := range fragment.nodes {
				parent.AppendChild(n)
			}
			continue
		}
		n := domNodeToHTMLNode(node)
		parent.AppendChild(n)
	}
}

func replaceChildren(parent *html.Node, nodes []spec.Node) {
	clearChildren(parent)
	for _, node := range nodes {
		n := domNodeToHTMLNode(node)
		parent.AppendChild(n)
	}
}

func clearChildren(node *html.Node) {
	if fc := node.FirstChild; fc != nil {
		fc.Parent = nil
	}
	if lc := node.LastChild; lc != nil {
		lc.Parent = nil
	}
	node.FirstChild = nil
	node.LastChild = nil
}

func getElementsByTagName(node *html.Node, name string) elementList {
	name = strings.ToUpper(name)
	var list elementList
	walkNodes(node, func(n *html.Node) bool {
		if strings.ToUpper(n.Data) == name {
			list = append(list, n)
		}
		return false
	})
	return list
}

func getElementsByClassName(node *html.Node, name string) elementList {
	var list elementList
	walkNodes(node, func(n *html.Node) bool {
		if hasClasses(getAttribute(n, "class"), name) {
			list = append(list, n)
		}
		return false
	})
	return list
}

func hasClasses(elementClassesStr, classesStr string) bool {
	elementClasses := strings.Fields(elementClassesStr)
	classes := strings.Fields(classesStr)

	set := make(map[string]struct{}, len(classesStr))
	for _, c := range classes {
		set[c] = struct{}{}
	}

	for _, c := range elementClasses {
		delete(set, c)
	}

	return len(set) == 0
}

func querySelector(node *html.Node, query string, includeParent bool) spec.Element {
	q := cascadia.MustCompile(query)
	if includeParent && q.Match(node) {
		return &Element{node: node}
	}
	if result := cascadia.Query(node, q); result != nil {
		return &Element{node: result}
	}
	return nil
}

func querySelectorAll(node *html.Node, query string, includeParent bool) nodeListHTMLElements {
	m := cascadia.MustCompile(query)
	var results []*html.Node
	if includeParent && m.Match(node) {
		results = slices.Insert(results, 0, node)
	}
	querySelectorSequence(node, m, func(element spec.Element) bool {
		el := element.(*Element)
		results = append(results, el.node)
		return true
	})
	return results
}

var _ spec.NodeList[spec.Element] = nodeListHTMLElements(nil)

type nodeListHTMLElements []*html.Node

func (n nodeListHTMLElements) Length() int { return len(n) }

func (n nodeListHTMLElements) Item(i int) spec.Element {
	return &Element{
		node: n[i],
	}
}

func closest(node *html.Node, selector string) spec.Element {
	s := cascadia.MustCompile(selector)
	for p := node; p != nil; p = p.Parent {
		if s.Match(p) {
			return htmlNodeToDomElement(p)
		}
	}
	return nil
}

func matches(node *html.Node, selector string) bool {
	s := cascadia.MustCompile(selector)
	return s.Match(node)
}

func isNamed(node *html.Node, name string) bool {
	if node == nil || node.Type != html.ElementNode {
		return false
	}
	id := getAttribute(node, "id")
	nm := getAttribute(node, "name")
	return (id != "" && id == name) || (nm != "" && nm == name)
}

func getAttribute(node *html.Node, name string) string {
	name = strings.ToLower(name)
	for _, att := range node.Attr {
		if att.Key == name {
			return att.Val
		}
	}
	return ""
}

func querySelectorSequence(n *html.Node, m cascadia.Matcher, yield func(spec.Element) bool) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if m.Match(c) {
			if !yield(&Element{node: c}) {
				return false
			}
		}
		if !querySelectorSequence(c, m, yield) {
			return false
		}
	}
	return true
}

// compareDocumentPosition is based on https://dom.spec.whatwg.org/#dom-node-comparedocumentposition
// it does not support attributes
func compareDocumentPosition(this *html.Node, other spec.Node) spec.DocumentPosition {
	node1 := domNodeToHTMLNode(other)
	node2 := this
	if node1 == node2 {
		return 0
	}
	// attribute nodes not handled
	owner := ownerDocumentNode(node2)
	sameRoot := ownerDocumentNode(node1) == owner
	bothConnected := isConnected(node1) && isConnected(node2)
	if node1 == nil || node2 == nil || !bothConnected || !sameRoot {
		// random consistent value for preceding or following is not handled
		return spec.DocumentPositionDisconnected | spec.DocumentPositionImplementationSpecific
	}
	for ancestor := range node2.Ancestors() {
		if ancestor == node1 {
			return spec.DocumentPositionContains | spec.DocumentPositionPreceding
		}
	}
	for ancestor := range node1.Ancestors() {
		if ancestor == node2 {
			return spec.DocumentPositionContainedBy | spec.DocumentPositionFollowing
		}
	}
	for descendant := range owner.Descendants() {
		if descendant == node1 {
			return spec.DocumentPositionPreceding
		}
		if descendant == node2 {
			return spec.DocumentPositionFollowing
		}
	}
	return spec.DocumentPositionFollowing
}

// compareDocumentFragmentPosition was generated by an LLM. I did not test it.
func compareDocumentFragmentPosition(this []*html.Node, other spec.Node) spec.DocumentPosition {
	node1 := domNodeToHTMLNode(other)
	if node1 == nil {
		return spec.DocumentPositionDisconnected | spec.DocumentPositionImplementationSpecific
	}

	var closestPosition spec.DocumentPosition
	foundAny := false

	for _, node2 := range this {
		if !isConnected(node2) || ownerDocumentNode(node1) != ownerDocumentNode(node2) {
			continue
		}
		foundAny = true

		pos := compareDocumentPosition(node2, other)

		// Priority ordering: contains > preceding > following
		if pos&spec.DocumentPositionContains != 0 {
			return spec.DocumentPositionContains
		}
		if pos&spec.DocumentPositionPreceding != 0 {
			closestPosition = spec.DocumentPositionPreceding
		} else if pos&spec.DocumentPositionFollowing != 0 && closestPosition != spec.DocumentPositionPreceding {
			closestPosition = spec.DocumentPositionFollowing
		}
	}

	if !foundAny {
		return spec.DocumentPositionDisconnected | spec.DocumentPositionImplementationSpecific
	}

	return closestPosition
}
