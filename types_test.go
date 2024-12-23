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

var xmlString2 =`
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

func TestWalking(t *testing.T) {
	assert := assert.New(t)

	root, _ := UnmarshalXML(xmlString2)

	root.Walk(func(n *Node) bool {
		assert.True(n.XMLName.Local == "root" || n.Parent != nil)
		return true
	})
}