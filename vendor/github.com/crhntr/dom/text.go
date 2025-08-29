package dom

import (
	"golang.org/x/net/html"

	"github.com/crhntr/dom/spec"
)

type Text struct {
	node *html.Node
}

func (t *Text) Data() string     { return t.node.Data }
func (t *Text) SetData(d string) { t.node.Data = d }

func (t *Text) NodeType() spec.NodeType         { return nodeType(t.node.Type) }
func (t *Text) IsConnected() bool               { return isConnected(t.node) }
func (t *Text) OwnerDocument() spec.Document    { return ownerDocument(t.node) }
func (t *Text) Length() int                     { return len(t.node.Data) }
func (t *Text) ParentNode() spec.Node           { return parentNode(t.node) }
func (t *Text) ParentElement() spec.Element     { return parentElement(t.node) }
func (t *Text) PreviousSibling() spec.ChildNode { return previousSibling(t.node) }
func (t *Text) NextSibling() spec.ChildNode     { return nextSibling(t.node) }
func (t *Text) TextContent() string             { return t.node.Data }
func (t *Text) CloneNode(_ bool) spec.Node {
	return &Text{
		node: &html.Node{
			Type: html.TextNode,
			Data: t.node.Data,
		},
	}
}

func (t *Text) CompareDocumentPosition(other spec.Node) spec.DocumentPosition {
	return compareDocumentPosition(t.node, other)
}

func (t *Text) IsSameNode(other spec.Node) bool { return isSameNode(t.node, other) }

func (t *Text) String() string { return t.node.Data }
