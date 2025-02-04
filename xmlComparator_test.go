package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyList = []string{}

func TestInvalidStrings(t *testing.T) {
	assert := assert.New(t)

	xmlSample := "<a><b/><c/></a>"

	assert.Equal([]string{"Can't parse the first sample: EOF"}, CompareXmlStrings("", "", false))
	assert.Equal([]string{"Can't parse the first sample: EOF"}, CompareXmlStrings("not an XML", xmlSample, false))
	assert.Equal([]string{"Can't parse the second sample: EOF"}, CompareXmlStrings(xmlSample, "", false))
	assert.Equal([]string{"Can't parse the second sample: EOF"}, CompareXmlStrings(xmlSample, "not an XML", false))
}

func TestDifferentNames(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]string{"Node names differ: 'note' vs 'root', path='/note'"}, CompareXmlStrings(xmlString1, xmlString2, true))
}

func TestDifferentNameSpaces(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<X:a xmlns:X="space1"><b/><c/></X:a>`
	xmlSample2 := `<a xmlns="space2"><b/><c/></a>`
	assert.Equal([]string{"Node namespaces differ: 'space1' vs 'space2', path='/a'"}, CompareXmlStrings(xmlSample1, xmlSample2, true))
}

func TestIgnoringNameSpacePrefixes(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<X:a xmlns:X="space1"><b/><c/></X:a>`
	xmlSample2 := `<a xmlns="space1"><b/><c/></a>`
	assert.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentAttributes(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy"/>`
	xmlSample3 := `<a attr1="12" attr2="ab"/>`
	assert.Equal([]string{"Attributes differ: counts 2 vs 1: attr1[0]:+1, path='/a'"},
		CompareXmlStrings(xmlSample1, xmlSample2, false))
	assert.Equal([]string{"Attributes differ: 'attr2=xy' vs 'attr2=ab', path='/a'"},
		CompareXmlStrings(xmlSample1, xmlSample3, true))

	xmlSample4 := `<X:a xmlns:X="space1"><b foo=""/><c/></X:a>`
	xmlSample5 := `<a xmlns="space2"><b foo="bar"/><c/></a>`
	diffs := CompareXmlStrings(xmlSample4, xmlSample5, false)
	assert.Equal(2, len(diffs))
	assert.Equal("Node namespaces differ: 'space1' vs 'space2', path='/a'", diffs[0])
	assert.Equal("Attributes differ: 'foo=' vs 'foo=bar', path='/a/b[0]'", diffs[1])
}

func TestEqualWithDifferentAttributesOrder(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy" attr1="12"/>`
	assert.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentCharData(t *testing.T) {
	assert := assert.New(t)

	assert.Equal([]string{"Node texts differ: '' vs 'Some text ...\n    \n\tmixed with elements', path='/note'"},
		CompareXmlStrings(xmlString1, xmlMixed, true))
}

func TestIgnoringWhitespace(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a><b/></a>`
	xmlSample2 := `<a>
	    <b/></a>`
	assert.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestIgnoreComments(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := "<a><!-- test --></a>"
	xmlSample2 := "<a></a>"
	assert.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestStoppingOnTheFirstError(t *testing.T) {
	assert := assert.New(t)

	diffs1 := CompareXmlStrings(xmlString1, xmlMixed, true)
	assert.Equal(1, len(diffs1))
	assert.Equal("Node texts differ: '' vs 'Some text ...\n    \n\tmixed with elements', path='/note'", diffs1[0])

	diffs2 := CompareXmlStrings(xmlString1, xmlMixed, false)
	assert.Equal(3, len(diffs2))
	assert.Equal(diffs1[0], diffs2[0])
	assert.Equal("Node texts differ: 'Tove' vs 'Jani', path='/note/to[0]'", diffs2[1])
	assert.Equal("Node texts differ: 'Jani' vs 'Tove', path='/note/from[1]'", diffs2[2])
}

func TestIgnoreList(t *testing.T) {
	assert := assert.New(t)

	diffs := CompareXmlStringsEx(xmlString1, xmlMixed, false, []string{`Node texts differ: '.+' vs '.+'`})
	assert.Equal(1, len(diffs))

	xmlString5 := `<a>Node Content</a>`
	xmlString6 := `<a>Another Content</a>`
	diffs = CompareXmlStringsEx(xmlString5, xmlString6, false, []string{`Node texts differ: '.+' vs '.+', path='/a'`})
	assert.Equal(emptyList, diffs)
}

func TestCDataComparison(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<!DOCTYPE a>
<a xmlns:xyz="https://www.xmldiff.com/xyz">
	<b>text</b>
	<c>
		<d/>
		<xyz:e/>
	</c>
</a>`
	xmlSample2 := `
<a xmlns:vwy="https://www.xmldiff.com/xyz">
	<b><![CDATA[text]]></b>
	<c>
		<d/>
		<vwy:e/>
	</c>
</a>`
	assert.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentElementsOrder(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a><b/><c/></a>`
	xmlSample2 := `<a><c/><b/></a>`
	assert.Equal([]string{"Children order differ for 2 nodes, path='/a'"}, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentElementsOrderByAttributes(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `
<items version="2.1">
  <item uid="ca_1">
    <name>name 1</name>
  </item>
  <item uid="ca_2">
    <name>name 2</name>
  </item>
  <item uid="ca_2">
    <name>name 2</name>
  </item>
  <item uid="ca_4">
    <name>name 4</name>
  </item>
</items>
`
	xmlSample2 := `
<items version="2.1">
  <item uid="ca_1">
    <name>name 1</name>
  </item>
  <item uid="ca_2">
    <name>name 2</name>
  </item>
  <item uid="ca_4">
    <name>name 4</name>
  </item>
  <item uid="ca_2">
    <name>name 2</name>
  </item>
</items>
`
	assert.Equal([]string{"Children order differ for 4 nodes, path='/items'"}, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentChildren(t *testing.T) {
	assert := assert.New(t)

	xmlSample1 := `<a><b><c/><c/><d/></b></a>`
	xmlSample2 := `<a><b><d/><e/><e/><e/></b></a>`
	assert.Equal([]string{"Children differ: counts 3 vs 4: c[0]:+2, e[1]:-3, path='/a/b'"}, CompareXmlStrings(xmlSample1, xmlSample2, false))
	assert.Equal([]string{"Children differ: counts 4 vs 3: e[1]:+3, c[0]:-2, path='/a/b'"}, CompareXmlStrings(xmlSample2, xmlSample1, false))
}

func TestDifferentChildren2(t *testing.T) {
	assert := assert.New(t)

	// Edits: DELETE 'c', MODIFY 'd', Add 'e'
	xmlSample1 := `<a><b><c/><d>1</d></b></a>`
	xmlSample2 := `<a><b><d>2</d><e/></b></a>`
	assert.Equal([]string{"Children differ: counts 2 vs 2: c[0]:+1, e[1]:-1, path='/a/b'", "Node texts differ: '1' vs '2', path='/a/b/d[1]'"},
		CompareXmlStrings(xmlSample1, xmlSample2, false))
	assert.Equal([]string{"Children differ: counts 2 vs 2: e[1]:+1, c[0]:-1, path='/a/b'", "Node texts differ: '2' vs '1', path='/a/b/d[0]'"},
		CompareXmlStrings(xmlSample2, xmlSample1, false))
}

func TestAreEqualNumbers(t *testing.T) {
	assert := assert.New(t)

	assert.True(areEqualNumbers("0.2", "0.20"))
	assert.True(areEqualNumbers("2", "1.9999997"))
	assert.False(areEqualNumbers("1.2", "1,2"))
	assert.False(areEqualNumbers("2", "abc"))
}
