package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaces(t *testing.T) {
	assert := assert.New(t)

	root, err := parseXML(soapString)
	assert.Nil(err)

	assert.Equal("Envelope", nodeName(root))
	assert.Equal("http://www.w3.org/2001/12/soap-envelope", nodeSpace(root))
	assert.Equal("Body", nodeName(&root.Children[0]))
	assert.Equal("http://www.w3.org/2001/12/soap-envelope", nodeSpace(&root.Children[0]))
	assert.Equal("GetQuotation", nodeName(&root.Children[0].Children[0]))
	assert.Equal("http://www.xyz.org/quotations", nodeSpace(&root.Children[0].Children[0]))
}

func TestXmlPathString(t *testing.T) {
	assert := assert.New(t)

	root, _ := parseXML(xmlString2)

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
