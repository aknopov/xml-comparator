package xmlcomparator

import (
	"bytes"
	"encoding/xml"
	"hash/crc32"
	"sort"
	"strconv"
	"strings"
)

type DiffType int

const (
	DiffEqual DiffType = iota
	DiffName
	DiffSpace
	DiffContent
	DiffAttributes
	DiffChildren
	DiffChildrenOrder
)

// Abstract XML node presentation
type Node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr
	Content  string
	Children []Node
	Parent   *Node
	Idx      int
}

type XmlDiff struct {
	Ref       *Node
	Type      DiffType
	Idx       int
	DiffAttrs []xml.Attr
	DiffNodes []Node
}

// For parsing
type parseNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr  `xml:"-"`
	Content  []byte      `xml:",innerxml"`
	CharData string      `xml:",chardata"`
	Children []parseNode `xml:",any"`
	Parent   *parseNode  `xml:"-"`
	Hash     uint32      `xml:"-"`
}

var crc32c = crc32.MakeTable(crc32.Castagnoli)

// Walks breadth-first through the XML tree calling the function for iteslef and then for each child node
//   - f - function to call for each node; should return `false` to stop traversiong
func (node *parseNode) Walk(f func(*parseNode) bool) {
	if !f(node) {
		return
	}

	for i := range node.Children {
		node.Children[i].Walk(f)
	}
}

// Unmarshals XML data into a Node structure - "encoding/xml" package compatible
func (n *parseNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node parseNode

	return d.DecodeElement((*node)(n), &start)
}

// Unmarshals XML string into a Node structure
//   - xmlString - XML string to unmarshal
//
// Returns: root node of the XML tree and error if any
func UnmarshalXML(xmlString string) (*parseNode, error) {
	buf := bytes.NewBuffer([]byte(xmlString))
	dec := xml.NewDecoder(buf)

	var root parseNode
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	root.Walk(func(n *parseNode) bool {
		for i := range n.Children {
			n.Children[i].Parent = n
		}
		return true
	})

	root.hashCode()

	return &root, nil
}

// Creates a string representation of the XML path to the node.
//
// Path elements are node names separated by slashes.
//
// Child element might have its index, unless it is the only child - handy for dealing with arrays.
func (node *parseNode) Path() string {
	path := make([]string, 0)
	currNode := node

	for currNode.Parent != nil {
		siblings := currNode.Parent.Children
		nodeName := currNode.Name()
		if len(siblings) == 1 {
			path = append(path, "/"+nodeName)
		} else {
			for i := 0; i < len(siblings); i++ {
				if siblings[i].Hash == currNode.Hash {
					path = append(path, "/"+nodeName+"["+strconv.Itoa(i)+"]")
					break
				}
			}
		}
		currNode = currNode.Parent
	}
	path = append(path, "/"+currNode.Name())

	// Reverse the path
	size := len(path)
	for i := 0; i < size/2; i++ {
		path[i], path[size-i-1] = path[size-i-1], path[i]
	}

	return strings.Join(path, "")
}

// Converts XML node to a string that includes node name and attribites.
func (node *parseNode) String() string {
	attStr := ""
	for i := range node.Attrs {
		attStr += AttrName(node.Attrs[i]) + "=" + node.Attrs[i].Value
		if i < len(node.Attrs)-1 {
			attStr += ", "
		}
	}

	ret := node.Name() + "[" + attStr + "]"

	if len(node.Children) == 0 {
		ret += " = " + string(node.Content)
	}

	return ret
}

// Convenience shortcut functions

func (node *parseNode) Name() string {
	return node.XMLName.Local
}

func (node *parseNode) Space() string {
	return node.XMLName.Space
}

func AttrName(attr xml.Attr) string {
	return attr.Name.Local
}

func AttrSpace(attr xml.Attr) string {
	return attr.Name.Space
}

func AttrValue(attr xml.Attr) string {
	return attr.Value
}

func extractAttributes(node *parseNode) map[string]string {
	attrs := make(map[string]string, len(node.Attrs))
	for i := range node.Attrs {
		// Namesapce attributes are processed separately
		if !isNameSpaceAttr(node.Attrs[i]) {
			attrs[AttrName(node.Attrs[i])] = node.Attrs[i].Value
		}
	}
	return attrs
}

func isNameSpaceAttr(attr xml.Attr) bool {
	return AttrSpace(attr) == "xmlns" || AttrName(attr) == "xmlns"
}

func sortedClone[T comparable](slice []T, isLess func(T, T) bool) []T {
	ret := make([]T, len(slice))
	copy(ret, slice)
	sort.Slice(ret, func(i, j int) bool { return isLess(ret[i], ret[j]) })
	return ret
}

//------- Export conversions -------

// UC func (pNode *parseNode) toNode() Node {
// 	node := Node{
// 		XMLName:  pNode.XMLName,
// 		Attrs:    pNode.Attrs,
// 		Content:  string(pNode.Content),
// 		Children: make([]Node, len(pNode.Children)),
// 	}

// 	for i := range pNode.Children {
// 		node.Children[i] = pNode.Children[i].toNode()
// 	}

// 	return node
// }

//------- hash code generation -------

// Recursive function
func (node *parseNode) hashCode() uint32 {
	if node.Hash != 0 {
		return node.Hash
	}

	node.Hash = crc32.Checksum([]byte(node.Name()), crc32c)
	node.Hash = crc32.Update(node.Hash, crc32c, []byte(strings.TrimSpace(node.CharData)))

	for i := range node.Attrs {
		if !isNameSpaceAttr(node.Attrs[i]) {
			node.Hash = crc32.Update(node.Hash, crc32c, []byte(AttrName(node.Attrs[i])))
			node.Hash = crc32.Update(node.Hash, crc32c, []byte(AttrValue(node.Attrs[i])))
		}
	}

	// Cheap and cheerful
	for i := range node.Children {
		node.Hash = 31*node.Hash + node.Children[i].hashCode()
	}

	return node.Hash
}
