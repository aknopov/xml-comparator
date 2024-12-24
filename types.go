package xmlcomparator

import (
	"bytes"
	"encoding/xml"
)

// Abstract XML node presentation
type Node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:"-"`
	Content  []byte     `xml:",innerxml"`
	CharData string     `xml:",chardata"`
	Children []Node     `xml:",any"`
	Parent   *Node      `xml:"-"`
}

// Path to an element in XML from the root
type XmlPath struct {
	Node Node
}

// Walks depth-first through the XML tree calling the function for iteslef and then for each child node
//   - f - function to call for each node; should return `false` to stop traversiong
func (node *Node) Walk(f func(*Node) bool) {

	if !f(node) {
		return
	}

	for i := range node.Children {
		node.Children[i].Walk(f)
	}
}

// Unmarshals XML data into a Node structure - "encoding/xml" package compatible
func (n *Node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node Node

	return d.DecodeElement((*node)(n), &start)
}

// Unmarshals XML string into a Node structure
//   - xmlString - XML string to unmarshal
//
// Returns: root node of the XML tree and error if any
func UnmarshalXML(xmlString string) (*Node, error) {
	buf := bytes.NewBuffer([]byte(xmlString))
	dec := xml.NewDecoder(buf)

	var root Node
	if err := dec.Decode(&root); err != nil {
		return nil, err
	}

	root.Walk(func(n *Node) bool {
		for i := range n.Children {
			n.Children[i].Parent = n
		}
		return true
	})

	return &root, nil
}

// Converts XML node to a string that includes node name and attribites.
func (node *Node) Stringify() string {
	attStr := ""
	for i, a := range node.Attrs {
		attStr += a.Name.Local + "=" + a.Value
		if i < len(node.Attrs)-1 {
			attStr += ", "
		}
	}

	ret := node.XMLName.Local + "[" + attStr + "]"

	if len(node.Children) == 0 {
		ret += " = " + string(node.Content)
	}

	return ret
}

// Compares two slices of comparable elements (insteaed)
//   - a - first slice
//   - b - second slice
//
// Returns: `true` if slices are identical, `false` otherwise
func SlicesEqual[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// Gets a value from a map by key or returns a default value if the key is not found
//   - aMap - map to get value from
//   - k - key to get value for
//   - def - default value to return if key is not found
func GetOrDefault[K comparable, V any](aMap map[K]V, k K, def V) V {
	if v, ok := aMap[k]; ok {
		return v
	}
	return def
}
