package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestDifferentCharData(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]string{"Nodes text differ: '' vs 'Some text ...\n    \n\tmixed with elements'"}, CompareXmlString(xmlString1, xmlMixed, true))
}

func TestDifferentAttributes(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy"/>`
	xmlSample3 := `<a attr1="12" attr2="ab"/>`
	xmlSample4 := `<a attr2="xy" attr1="12"/>`
	assert.Equal([]string{"Attributes count differ: 2 vs 1"}, CompareXmlString(xmlSample1, xmlSample2, false))
	assert.Equal([]string{"Attributes differ: '[{attr2 xy}]' vs '[{attr2 ab}]'"}, CompareXmlString(xmlSample1, xmlSample3, false))
	assert.Equal([]string{}, CompareXmlString(xmlSample1, xmlSample4, false)) // order doesn't matter
}

func TestStoppingOnFirstError(t *testing.T) {
	// UC
}

func TestIgnoreList(t *testing.T) {
	//	assert := assert.New(t)

	// UC	assert.Equal([]string{}, CompareXmlStringEx(xmlString1, xmlString2, false, []string{"color"}))
}

func TestAreFieldsTheSameNumbers(t *testing.T) {
	assert := assert.New(t)

	assert.True(areFieldsTheSameNumbers("0.2", "0.20"))
	assert.True(areFieldsTheSameNumbers("2", "1.9999997"))
	assert.False(areFieldsTheSameNumbers("2", "abc"))
}

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

testAreIdentical_withIgnoreWhitespacees
String testXml = String.format("<a>%n <b/>%n</a>");
String controlXml = "<a><b/></a>";
assertThat(testXml).and(controlXml).ignoreWhitespace().areIdentical();

testAreIdentical_withIgnoreComments_1_0
String testXml = "<a><!-- test --></a>";
String controlXml = "<a></a>";
assertThat(testXml).and(controlXml).ignoreCommentsUsingXSLTVersion("1.0").areIdentical();

testAreIdentical_withNormalizeWhitespace
String testXml = String.format("<a>%n  <b>%n  Test%n  Node%n  </b>%n</a>");
String controlXml = "<a><b>Test Node</b></a>";
assertThat(testXml).and(controlXml).normalizeWhitespace().areIdentical();

estAreIdentical_withDifferentAttributesOrder
String testXml = "<Element attr2=\"xy\" attr1=\"12\"/>";
String controlXml = "<Element attr1=\"12\" attr2=\"xy\"/>";
assertThat(testXml).and(controlXml).areIdentical();
*/
