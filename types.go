package xmlcomparator

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// Creates a string representation of the XML path to the node.
//
// path elements are node names separated by slashes.
//
// Child element might have its index, unless it is the only child - handy for dealing with arrays.
func (node *parseNode) path() string {
	path := make([]string, 0)
	currNode := node

	for currNode.Parent != nil {
		siblings := currNode.Parent.Children
		nodeName := nodeName(currNode)
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
	path = append(path, "/"+nodeName(currNode))

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
		attStr += attrName(&node.Attrs[i]) + "=" + node.Attrs[i].Value
		if i < len(node.Attrs)-1 {
			attStr += ", "
		}
	}

	ret := nodeName(node) + "[" + attStr + "]"

	if len(node.Children) == 0 {
		ret += " = " + string(node.Content)
	}

	return ret
}

// Convenience shortcut functions

func nodeName(node *parseNode) string {
	return node.XMLName.Local
}
func nodeSpace(node *parseNode) string {
	return node.XMLName.Space
}

func attrName(attr *xml.Attr) string {
	return attr.Name.Local
}

func attrSpace(attr *xml.Attr) string {
	return attr.Name.Space
}

func attrValue(attr *xml.Attr) string {
	return attr.Value
}

func isNameSpaceAttr(attr *xml.Attr) bool {
	return attrSpace(attr) == "xmlns" || attrName(attr) == "xmlns"
}
