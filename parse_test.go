package xmlcomparator

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	assertT := assert.New(t)

	root, err := parseXML(xmlString1)
	assertT.Nil(err)

	assertT.Nil(root.Parent)
	for _, child := range root.Children {
		assertT.NotNil(child.Parent)
	}

	assertT.Equal("note[color=red]", root.String())
	assertT.Equal(5, len(root.Children))
	assertT.Equal("to[] = Tove", root.Children[0].String())
	assertT.Equal("body[] = Don't forget me this weekend!", root.Children[4].String())
}

func TestParsingFailure(t *testing.T) {
	assertT := assert.New(t)

	root, err := parseXML("bogus")
	assertT.Nil(root)
	assertT.NotNil(err)
}

func TestWalking(t *testing.T) {
	assertT := assert.New(t)

	root, _ := parseXML(xmlString2)

	root.walk(func(n *parseNode) bool {
		assertT.True(nodeName(n) == "root" || n.Parent != nil)
		assertT.NotZero(n.Hash)
		return true
	})
}

func TestHashCodeGeneration(t *testing.T) {
	assertT := assert.New(t)

	root1, _ := parseXML(`<a><b/><c/></a>`)
	root2, _ := parseXML(`<a><c/><b/></a>`)
	assertT.NotEqual(root1.Hash, root2.Hash)

	root3, _ := parseXML(`<a><b>Text</b><c/></a>`)
	assertT.NotEqual(root1.Hash, root3.Hash)

	root4, _ := parseXML(`<a><b foo="bar"/><c/></a>`)
	assertT.NotEqual(root1.Hash, root4.Hash)
}

func TestHashCodeCaching(t *testing.T) {
	assertT := assert.New(t)

	node := parseNode{XMLName: xml.Name{Space: "spc", Local: "name"}}
	assertT.Equal(uint32(0), node.Hash)
	hash := node.hashCode()
	assertT.Equal(hash, node.Hash)

	assertT.Equal(hash, node.hashCode())
}
