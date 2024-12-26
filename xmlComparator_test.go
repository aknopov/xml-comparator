package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyList = []string{}

func TestInvalidStrings(t *testing.T) {
	assert := assert.New(t)

	xmlSample := "<a><b/><c/></a>"

	assert.Equal([]string{"Can't parse first sample: EOF"}, CompareXmlString("", "", false))
	assert.Equal([]string{"Can't parse first sample: EOF"}, CompareXmlString("not an XML", xmlSample, false))
	assert.Equal([]string{"Can't parse second sample: EOF"}, CompareXmlString(xmlSample, "", false))
	assert.Equal([]string{"Can't parse second sample: EOF"}, CompareXmlString(xmlSample, "not an XML", false))
}

func TestDifferentNames(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]string{"Node names differ: 'note' vs 'root'"}, CompareXmlString(xmlString1, xmlString2, true))
}

func TestDifferentNamespaces(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<X:a xmlns:X="space1"><b/><c/></X:a>`
	xmlSample2 := `<a xmlns="space2"><b/><c/></a>`
	assert.Equal([]string{"Node namespaces differ: 'space1' vs 'space2'"}, CompareXmlString(xmlSample1, xmlSample2, true))
}

func TestIgnoringNamesapcePrefixes(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<X:a xmlns:X="space1"><b/><c/></X:a>`
	xmlSample2 := `<a xmlns="space1"><b/><c/></a>`
	assert.Equal(emptyList, CompareXmlString(xmlSample1, xmlSample2, false))
}

func TestDifferentAttributes(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy"/>`
	xmlSample3 := `<a attr1="12" attr2="ab"/>`
	assert.Equal([]string{"Attributes count differ: 2 vs 1"}, CompareXmlString(xmlSample1, xmlSample2, false))
	assert.Equal([]string{"Attributes differ: '[attr2=xy]' vs '[attr2=ab]'"}, CompareXmlString(xmlSample1, xmlSample3, true))

	xmlSample4 := `<X:a xmlns:X="space1"><b foo=""/><c/></X:a>`
	xmlSample5 := `<a xmlns="space2"><b foo="bar"/><c/></a>`
	diffs := CompareXmlString(xmlSample4, xmlSample5, false)
	assert.Equal(3, len(diffs))
	assert.Equal("Node namespaces differ: 'space1' vs 'space2'", diffs[0])
	assert.Equal("Attributes differ: '[xmlns=space1]' vs '[xmlns=space2]'", diffs[1])
	assert.Equal("Attributes differ: '[foo=]' vs '[foo=bar]'", diffs[2])
}

func TestEqualWithDifferentAttributesOrder(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy" attr1="12"/>`
	assert.Equal(emptyList, CompareXmlString(xmlSample1, xmlSample2, false))
}

func TestDifferentCharData(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]string{"Nodes text differ: '' vs 'Some text ...\n    \n\tmixed with elements'"}, CompareXmlString(xmlString1, xmlMixed, true))
}

func TestIgnoringWhitespace(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a><b/></a>`
	xmlSample2 := `<a>
	    <b/></a>`
	assert.Equal(emptyList, CompareXmlString(xmlSample1, xmlSample2, false))
}

func TestIgnoreComments(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := "<a><!-- test --></a>"
	xmlSample2 := "<a></a>"
	assert.Equal(emptyList, CompareXmlString(xmlSample1, xmlSample2, false))
}

func TestStoppingOnFirstError(t *testing.T) {
	assert := assert.New(t)

	diffs1 := CompareXmlString(xmlString1, xmlMixed, true)
	assert.Equal(1, len(diffs1))
	assert.Equal("Nodes text differ: '' vs 'Some text ...\n    \n\tmixed with elements'", diffs1[0])

	diffs2 := CompareXmlString(xmlString1, xmlMixed, false)
	assert.Equal(3, len(diffs2))
	assert.Equal(diffs1[0], diffs2[0])
	assert.Equal("Nodes text differ: 'Tove' vs 'Jani'", diffs2[1])
	assert.Equal("Nodes text differ: 'Jani' vs 'Tove'", diffs2[2])
}

func TestIgnoreList(t *testing.T) {
	assert := assert.New(t)

	diffs := CompareXmlStringEx(xmlString1, xmlMixed, false, []string{`Nodes text differ: '\w+' vs '\w+'`})
	assert.Equal(1, len(diffs))
}

func TestCDataComparison(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<!DOCTYPE a><a xmlns:xyz="https://www.xmlunit.com/xyz"><b>text</b><c><d/><xyz:e/></c></a>`
	xmlSample2 := `<a xmlns:vwy="https://www.xmlunit.com/xyz"><b><![CDATA[text]]></b><c><d/><vwy:e/></c></a>`
	assert.Equal(emptyList, CompareXmlString(xmlSample1, xmlSample2, false))
}

func TestDifferentElementsOrder(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a><b/><c/></a>`
	xmlSample2 := `<a><c/><b/></a>`
	assert.Equal([]string{"Nodes order differ for 2 nodes"}, CompareXmlString(xmlSample1, xmlSample2, true))
}

func TestAreFieldsTheSameNumbers(t *testing.T) {
	assert := assert.New(t)

	assert.True(stringsAsNumbersEqual("0.2", "0.20"))
	assert.True(stringsAsNumbersEqual("2", "1.9999997"))
	assert.False(stringsAsNumbersEqual("2", "abc"))
}

// UC xmlunit -> xmldiff

/*
   String testXml = "<!DOCTYPE a>" +
           "<a xmlns:xyz=\"https://www.xmlunit.com/xyz\">" +
           "   <b>text</b>" +
           "   <c>" +
           "      <d/>" +
           "      <xyz:e/>" +
           "   </c>" +
           "</a>";
   String controlXml = "" +
           "<a xmlns:vwy=\"https://www.xmlunit.com/xyz\">" +
           "   <b><![CDATA[text]]></b>" +
           "   <c>" +
           "      <d/>" +
           "      <vwy:e/>" +
           "   </c>" +
           "</a>";
   assertThat(testXml).and(controlXml).areIdentical();


*/
