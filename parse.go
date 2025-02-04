package xmlcomparator

import (
	"bytes"
	"encoding/xml"
	"hash/crc32"
	"strings"
)

var crc32c = crc32.MakeTable(crc32.Castagnoli)

type parseNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr  `xml:"-"`
	Content  []byte      `xml:",innerxml"`
	CharData string      `xml:",chardata"`
	Children []parseNode `xml:",any"`
	Parent   *parseNode  `xml:"-"`
	Hash     uint32      `xml:"-"`
}

// Unmarshals XML data into a Node structure - `Decoder` requirement to parse attributes.
func (n *parseNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node parseNode

	return d.DecodeElement((*node)(n), &start)
}

// Unmarshals XML string into a Node structure
//   - xmlString - XML string to unmarshal
//
// Returns: root node of the XML tree and error if any
func parseXML(xmlString string) (*parseNode, error) {
	buf := bytes.NewBuffer([]byte(xmlString))
	dec := xml.NewDecoder(buf)

	var root parseNode
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	root.walk(func(n *parseNode) bool {
		for i := range n.Children {
			n.Children[i].Parent = n
		}
		return true
	})

	root.hashCode()

	return &root, nil
}

// Walks depth-first through the XML tree calling the function for iteslef and then for each child node
//   - f - function to call for each node; should return `false` to stop traversiong
func (node *parseNode) walk(f func(*parseNode) bool) {
	if !f(node) {
		return
	}

	for i := range node.Children {
		node.Children[i].walk(f)
	}
}

//------- hash code generation -------

// Recursive function
func (node *parseNode) hashCode() uint32 {
	if node.Hash != 0 {
		return node.Hash
	}

	node.Hash = crc32.Checksum([]byte(nodeName(node)), crc32c)
	node.Hash = crc32.Update(node.Hash, crc32c, []byte(strings.TrimSpace(node.CharData)))

	for i := range node.Attrs {
		attrPtr := &node.Attrs[i]
		if !isNameSpaceAttr(attrPtr) {
			node.Hash = crc32.Update(node.Hash, crc32c, []byte(attrName(attrPtr)))
			node.Hash = crc32.Update(node.Hash, crc32c, []byte(attrValue(attrPtr)))
		}
	}

	// Cheap and cheerful
	for i := range node.Children {
		node.Hash = 31*node.Hash + node.Children[i].hashCode()
	}

	return node.Hash
}
