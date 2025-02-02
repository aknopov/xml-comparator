package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsing(t *testing.T) {
	assert := assert.New(t)

	root, err := UnmarshalXML(xmlString1)
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

	root, err := UnmarshalXML("bogus")
	assert.Nil(root)
	assert.NotNil(err)
}

func TestNamespaces(t *testing.T) {
	assert := assert.New(t)

	root, err := UnmarshalXML(soapString)
	assert.Nil(err)

	assert.Equal("Envelope", root.Name())
	assert.Equal("http://www.w3.org/2001/12/soap-envelope", root.Space())
	assert.Equal("Body", root.Children[0].Name())
	assert.Equal("http://www.w3.org/2001/12/soap-envelope", root.Children[0].Space())
	assert.Equal("GetQuotation", root.Children[0].Children[0].Name())
	assert.Equal("http://www.xyz.org/quotations", root.Children[0].Children[0].Space())
}

func TestWalking(t *testing.T) {
	assert := assert.New(t)

	root, _ := UnmarshalXML(xmlString2)

	root.Walk(func(n *parseNode) bool {
		assert.True(n.Name() == "root" || n.Parent != nil)
		assert.NotZero(n.Hash)
		return true
	})
}

func TestXmlPathString(t *testing.T) {
	assert := assert.New(t)

	root, _ := UnmarshalXML(xmlString2)

	assert.Equal("/root", root.Path())
	assert.Equal("/root/animal[0]", root.Children[0].Path())
	assert.Equal("/root/animal[0]/p[0]", root.Children[0].Children[0].Path())
	assert.Equal("/root/animal[0]/dog[1]", root.Children[0].Children[1].Path())
	assert.Equal("/root/animal[0]/dog[1]/p", root.Children[0].Children[1].Children[0].Path())
	assert.Equal("/root/birds[1]", root.Children[1].Path())
	assert.Equal("/root/birds[1]/p[0]", root.Children[1].Children[0].Path())
	assert.Equal("/root/birds[1]/p[1]", root.Children[1].Children[1].Path())
	assert.Equal("/root/animal[2]", root.Children[2].Path())
	assert.Equal("/root/animal[2]/p", root.Children[2].Children[0].Path())
}

func TestHashCodeGeneration(t *testing.T) {
	assert := assert.New(t)

	root1, _ := UnmarshalXML(`<a><b/><c/></a>`)
	root2, _ := UnmarshalXML(`<a><c/><b/></a>`)
	assert.NotEqual(root1.Hash, root2.Hash)

	root3, _ := UnmarshalXML(`<a><b>Text</b><c/></a>`)
	assert.NotEqual(root1.Hash, root3.Hash)

	root4, _ := UnmarshalXML(`<a><b foo="bar"/><c/></a>`)
	assert.NotEqual(root1.Hash, root4.Hash)
}
