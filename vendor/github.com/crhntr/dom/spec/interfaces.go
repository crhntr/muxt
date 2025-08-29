package spec

import "iter"

// Node is based on a subset of the methods and types in
// https://dom.spec.whatwg.org/#interface-node as of 2021-12-10
//
// MinimalNode contains a subset of methods used by root nodes like Document.
//
// Some methods (listed below) were removed from Node and added to ParentNode because they
// only make sense in the context of a non-leaf node. By removing them from node it reduces
// API surface area of the (leaf) Text node.
// - HasChildNodes
// - ChildNodes
// - FirstChild
// - LastChild
// - Contains
// - InsertBefore
// - AppendChild
// - ReplaceChild
// - RemoveChild
//
// The following methods were removed because they do not apply across all relevant node types.
// - NodeValue (only applies to Text and Attr. The former already has Data and the latter is ignored)
// - IsEqualNode (has different comparisons for different node types and this makes implementation difficult)
//
// The following methods have been added in addition to those documented in the whatwg document.
// - Length
type Node interface {
	NodeType() NodeType
	CloneNode(deep bool) Node
	IsSameNode(other Node) bool
	TextContent() string
	CompareDocumentPosition(other Node) DocumentPosition
}

type ChildNode interface {
	Node

	IsConnected() bool
	OwnerDocument() Document
	ParentNode() Node
	ParentElement() Element
	PreviousSibling() ChildNode
	NextSibling() ChildNode

	// Length should be based on https://dom.spec.whatwg.org/#concept-node-length
	Length() int

	// LookupPrefix(namespace string)
	// LookupNamespaceURI(prefix string)
	// IsDefaultNamespace(namespace string) bool
}

// Normalizer may be implemented by a Node and should follow https://dom.spec.whatwg.org/#dom-node-normalize
type Normalizer interface {
	Normalize()
}

type DocumentPosition int

// DocumentPosition is based on const values in
// https://dom.spec.whatwg.org/#interface-node (reviewed on 2021-12-10)
const (
	DocumentPositionDisconnected DocumentPosition = 1 << iota
	DocumentPositionPreceding
	DocumentPositionFollowing
	DocumentPositionContains
	DocumentPositionContainedBy
	DocumentPositionImplementationSpecific
)

// NodeType is based on const values in
// https://dom.spec.whatwg.org/#interface-node (reviewed on 2021-12-10)
type NodeType int

const (
	NodeTypeUnknown NodeType = iota
	NodeTypeElement
	NodeTypeAttribute
	NodeTypeText
	NodeTypeCdataSection
	NodeTypeEntityReference
	NodeTypeEntity
	NodeTypeProcessingInstruction
	NodeTypeComment
	NodeTypeDocument
	NodeTypeDocumentType
	NodeTypeDocumentFragment
	NodeTypeNotation
)

func (nt NodeType) String() string {
	switch nt {
	case NodeTypeElement:
		return "element"
	case NodeTypeAttribute:
		return "Attribute"
	case NodeTypeText:
		return "Text"
	case NodeTypeCdataSection:
		return "CdataSection"
	case NodeTypeEntityReference:
		return "EntityReference"
	case NodeTypeEntity:
		return "Entity"
	case NodeTypeProcessingInstruction:
		return "ProcessingInstruction"
	case NodeTypeComment:
		return "Comment"
	case NodeTypeDocument:
		return "Document"
	case NodeTypeDocumentType:
		return "DocumentType"
	case NodeTypeDocumentFragment:
		return "DocumentFragment"
	case NodeTypeNotation:
		return "Notation"

	default:
		fallthrough
	case NodeTypeUnknown:
		return "Unknown"
	}
}

type NodeList[T Node] interface {
	Length() int
	Item(int) T
}

type Text interface {
	ChildNode

	Data() string
	SetData(string)

	// Split(n int) Text // CONSIDER: maybe implement this
	// WholeText() string // CONSIDER: maybe implement this
}

type Document interface {
	Node

	ElementQueries

	CreateElement(localName string) Element
	CreateElementIs(localName, is string) Element

	// CreateDocumentFragment() node

	CreateTextNode(text string) Text

	Head() Element
	Body() Element
}

// ParentNode is based on https://dom.spec.whatwg.org/#interface-parentnode. It also includes some fields and
// methods from Node that only make sense for non-leaf nodes such as Element, DocumentFragment, and Document.
type ParentNode interface {
	Node

	Children() ElementCollection
	FirstElementChild() Element
	LastElementChild() Element
	ChildElementCount() int

	Prepend(nodes ...Node)
	Append(nodes ...Node)
	ReplaceChildren(nodes ...Node)

	ElementQueries

	// the following methods are from node; however, they only make sense for parent nodes

	HasChildNodes() bool
	ChildNodes() NodeList[Node]
	FirstChild() ChildNode
	LastChild() ChildNode
	InsertBefore(node, child ChildNode) ChildNode
	AppendChild(node ChildNode) ChildNode
	ReplaceChild(node, child ChildNode) ChildNode
	RemoveChild(node ChildNode) ChildNode
}

type ElementQueries interface {
	Contains(other Node) bool

	GetElementsByTagName(name string) ElementCollection
	GetElementsByClassName(name string) ElementCollection

	QuerySelector(query string) Element
	QuerySelectorAll(query string) NodeList[Element]

	QuerySelectorIterator
}

// Element is based on
//
// InnerText methods are ignored due to rendering complexity; however, implementations may add them
// based on InnerTextSetter.
type Element interface {
	Node
	ChildNode
	ParentNode

	TagName() string
	ID() string
	ClassName() string

	GetAttribute(name string) string
	SetAttribute(name, value string)
	RemoveAttribute(name string)
	ToggleAttribute(name string) bool
	HasAttribute(name string) bool

	Closest(selector string) Element
	Matches(selector string) bool

	SetInnerHTML(s string)
	InnerHTML() string
	SetOuterHTML(s string)
	OuterHTML() string
}

type InnerTextSetter interface {
	SetInnerText(s string)
	InnerText() string
}

type ElementCollection interface {
	// Length returns the number of elements in the collection.
	Length() int

	// Item returns the element with index from the collection. The elements are sorted in tree order.
	Item(index int) Element

	// NamedItem returns the first element with ID or name from the collection.
	NamedItem(name string) Element
}

type DocumentFragment interface {
	Node

	Children() ElementCollection
	FirstElementChild() Element
	LastElementChild() Element
	ChildElementCount() int

	Append(nodes ...Node)
	Prepend(nodes ...Node)
	ReplaceChildren(nodes ...Node)

	QuerySelector(query string) Element
	QuerySelectorAll(query string) NodeList[Element]
	QuerySelectorIterator
}

type Comment interface {
	Node

	Data() string
	SetData() string
}

type QuerySelectorIterator interface {
	QuerySelectorSequence(query string) iter.Seq[Element]
}
