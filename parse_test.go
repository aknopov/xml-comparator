package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	assert := assert.New(t)

	root, err := parseXML(xmlString1)
	assert.Nil(err)

	assert.Nil(root.Parent)
	for _, child := range root.Children {
		assert.NotNil(child.Parent)
	}

	assert.Equal("note[color=red]", root.String())
	assert.Equal(5, len(root.Children))
	assert.Equal("to[] = Tove", root.Children[0].String())
	assert.Equal("body[] = Don't forget me this weekend!", root.Children[4].String())
}

func TestParsingFailure(t *testing.T) {
	assert := assert.New(t)

	root, err := parseXML("bogus")
	assert.Nil(root)
	assert.NotNil(err)
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
	assert := assert.New(t)

	root1, _ := parseXML(`<a><b/><c/></a>`)
	root2, _ := parseXML(`<a><c/><b/></a>`)
	assert.NotEqual(root1.Hash, root2.Hash)

	root3, _ := parseXML(`<a><b>Text</b><c/></a>`)
	assert.NotEqual(root1.Hash, root3.Hash)

	root4, _ := parseXML(`<a><b foo="bar"/><c/></a>`)
	assert.NotEqual(root1.Hash, root4.Hash)
}
