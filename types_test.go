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

	assert.Equal("Envelope", root.XMLName.Local)
	assert.Equal("http://www.w3.org/2001/12/soap-envelope", root.XMLName.Space)
	assert.Equal("Body", root.Children[0].XMLName.Local)
	assert.Equal("http://www.w3.org/2001/12/soap-envelope", root.Children[0].XMLName.Space)
	assert.Equal("GetQuotation", root.Children[0].Children[0].XMLName.Local)
	assert.Equal("http://www.xyz.org/quotations", root.Children[0].Children[0].XMLName.Space)
}

func TestWalking(t *testing.T) {
	assert := assert.New(t)

	root, _ := UnmarshalXML(xmlString2)

	root.Walk(func(n *Node) bool {
		assert.True(n.XMLName.Local == "root" || n.Parent != nil)
		return true
	})
}
