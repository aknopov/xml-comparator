package xmlcomparator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaces(t *testing.T) {
	assertT := assert.New(t)

	root, err := parseXML(soapString)
	assertT.Nil(err)

	assertT.Equal("Envelope", nodeName(root))
	assertT.Equal("http://www.w3.org/2001/12/soap-envelope", nodeSpace(root))
	assertT.Equal("Body", nodeName(&root.Children[0]))
	assertT.Equal("http://www.w3.org/2001/12/soap-envelope", nodeSpace(&root.Children[0]))
	assertT.Equal("GetQuotation", nodeName(&root.Children[0].Children[0]))
	assertT.Equal("http://www.xyz.org/quotations", nodeSpace(&root.Children[0].Children[0]))
}

func TestXmlPathString(t *testing.T) {
	assertT := assert.New(t)

	root, _ := parseXML(xmlString2)

	assertT.Equal("/root", root.Path())
	assertT.Equal("/root/animal[0]", root.Children[0].Path())
	assertT.Equal("/root/animal[0]/p[0]", root.Children[0].Children[0].Path())
	assertT.Equal("/root/animal[0]/dog[1]", root.Children[0].Children[1].Path())
	assertT.Equal("/root/animal[0]/dog[1]/p", root.Children[0].Children[1].Children[0].Path())
	assertT.Equal("/root/birds[1]", root.Children[1].Path())
	assertT.Equal("/root/birds[1]/p[0]", root.Children[1].Children[0].Path())
	assertT.Equal("/root/birds[1]/p[1]", root.Children[1].Children[1].Path())
	assertT.Equal("/root/animal[2]", root.Children[2].Path())
	assertT.Equal("/root/animal[2]/p", root.Children[2].Children[0].Path())
}

func TestStringerInterface(t *testing.T) {
	assertT := assert.New(t)

	root, _ := parseXML(soapString)

	assertT.Equal("Envelope[SOAP-ENV=http://www.w3.org/2001/12/soap-envelope, encodingStyle=http://www.w3.org/2001/12/soap-encoding]",
		fmt.Sprint(root))
}
