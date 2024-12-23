package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var xmlString1 = `
<note color="red">
    <to>Tove</to>
    <from>Jani</from>
    <date>2023-08-27T16:27:55+00:00</date>
    <heading>Reminder</heading>
    <body>Don't forget me this weekend!</body>
</note>`

var xmlString2 = `
<root type="vet_hospital">
    <animal>
        <p>This is dog</p>
        <dog>
           <p>tommy</p>
        </dog>
    </animal>
    <birds>
        <p class="bar">this is birds</p>
        <p>this is birds</p>
    </birds>
    <animal>
        <p>this is animals</p>
    </animal>
</root>`

var soapString = `
<?xml version = "1.0"?>
<SOAP-ENV:Envelope
   xmlns:SOAP-ENV = "http://www.w3.org/2001/12/soap-envelope"
   SOAP-ENV:encodingStyle = "http://www.w3.org/2001/12/soap-encoding">
   <SOAP-ENV:Body xmlns:m = "http://www.xyz.org/quotations">
      <m:GetQuotation>
         <m:QuotationsName>MiscroSoft</m:QuotationsName>
      </m:GetQuotation>
   </SOAP-ENV:Body>
</SOAP-ENV:Envelope>
`

func TestParsing(t *testing.T) {
	assert := assert.New(t)

	root, err := UnmarshalXML(xmlString1)
	assert.Nil(err)

	assert.Nil(root.Parent)
	for _, child := range root.Children {
		assert.NotNil(child.Parent)
	}

	assert.Equal("note[color=red]", root.Stringify())
	assert.Equal(5, len(root.Children))
	assert.Equal("to[] = Tove", root.Children[0].Stringify())
	assert.Equal("body[] = Don't forget me this weekend!", root.Children[4].Stringify())
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
