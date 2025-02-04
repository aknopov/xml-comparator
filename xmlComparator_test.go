package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var emptyList = []string{}

func TestInvalidStrings(t *testing.T) {
	assertT := assert.New(t)

	xmlSample := "<a><b/><c/></a>"

	assertT.Equal([]string{"Can't parse the first sample: EOF"}, CompareXmlStrings("", "", false))
	assertT.Equal([]string{"Can't parse the first sample: EOF"}, CompareXmlStrings("not an XML", xmlSample, false))
	assertT.Equal([]string{"Can't parse the second sample: EOF"}, CompareXmlStrings(xmlSample, "", false))
	assertT.Equal([]string{"Can't parse the second sample: EOF"}, CompareXmlStrings(xmlSample, "not an XML", false))
}

func TestDifferentNames(t *testing.T) {
	assertT := assert.New(t)

	assertT.Equal([]string{"Node names differ: 'note' vs 'root', path='/note'"}, CompareXmlStrings(xmlString1, xmlString2, true))
}

func TestDifferentNameSpaces(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<X:a xmlns:X="space1"><b/><c/></X:a>`
	xmlSample2 := `<a xmlns="space2"><b/><c/></a>`
	assertT.Equal([]string{"Node namespaces differ: 'space1' vs 'space2', path='/a'"}, CompareXmlStrings(xmlSample1, xmlSample2, true))
}

func TestIgnoringNameSpacePrefixes(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<X:a xmlns:X="space1"><b/><c/></X:a>`
	xmlSample2 := `<a xmlns="space1"><b/><c/></a>`
	assertT.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentAttributes(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy"/>`
	xmlSample3 := `<a attr1="12" attr2="ab"/>`
	assertT.Equal([]string{"Attributes differ: counts 2 vs 1: attr1[0]:+1, path='/a'"},
		CompareXmlStrings(xmlSample1, xmlSample2, false))
	assertT.Equal([]string{"Attributes differ: 'attr2=xy' vs 'attr2=ab', path='/a'"},
		CompareXmlStrings(xmlSample1, xmlSample3, true))

	xmlSample4 := `<X:a xmlns:X="space1"><b foo=""/><c/></X:a>`
	xmlSample5 := `<a xmlns="space2"><b foo="bar"/><c/></a>`
	diffs := CompareXmlStrings(xmlSample4, xmlSample5, false)
	assertT.Equal(2, len(diffs))
	assertT.Equal("Node namespaces differ: 'space1' vs 'space2', path='/a'", diffs[0])
	assertT.Equal("Attributes differ: 'foo=' vs 'foo=bar', path='/a/b[0]'", diffs[1])
}

func TestEqualWithDifferentAttributesOrder(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<a attr1="12" attr2="xy"/>`
	xmlSample2 := `<a attr2="xy" attr1="12"/>`
	assertT.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentCharData(t *testing.T) {
	assertT := assert.New(t)

	assertT.Equal([]string{"Node texts differ: '' vs 'Some text ...\n    \n\tmixed with elements', path='/note'"},
		CompareXmlStrings(xmlString1, xmlMixed, true))
}

func TestIgnoringWhitespace(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<a><b/></a>`
	xmlSample2 := `<a>
	    <b/></a>`
	assertT.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestIgnoreComments(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := "<a><!-- test --></a>"
	xmlSample2 := "<a></a>"
	assertT.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestStoppingOnTheFirstError(t *testing.T) {
	assertT := assert.New(t)

	diffs1 := CompareXmlStrings(xmlString1, xmlMixed, true)
	assertT.Equal(1, len(diffs1))
	assertT.Equal("Node texts differ: '' vs 'Some text ...\n    \n\tmixed with elements', path='/note'", diffs1[0])

	diffs2 := CompareXmlStrings(xmlString1, xmlMixed, false)
	assertT.Equal(3, len(diffs2))
	assertT.Equal(diffs1[0], diffs2[0])
	assertT.Equal("Node texts differ: 'Tove' vs 'Jani', path='/note/to[0]'", diffs2[1])
	assertT.Equal("Node texts differ: 'Jani' vs 'Tove', path='/note/from[1]'", diffs2[2])
}

func TestIgnoreList(t *testing.T) {
	assertT := assert.New(t)

	diffs := CompareXmlStringsEx(xmlString1, xmlMixed, false, []string{`Node texts differ: '.+' vs '.+'`})
	assertT.Equal(1, len(diffs))

	xmlString5 := `<a>Node Content</a>`
	xmlString6 := `<a>Another Content</a>`
	diffs = CompareXmlStringsEx(xmlString5, xmlString6, false, []string{`Node texts differ: '.+' vs '.+', path='/a'`})
	assertT.Equal(emptyList, diffs)
}

func TestCDataComparison(t *testing.T) {
	assertT := assert.New(t)

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
	assertT.Equal(emptyList, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentElementsOrder(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<a><b/><c/></a>`
	xmlSample2 := `<a><c/><b/></a>`
	assertT.Equal([]string{"Children order differ for 2 nodes, path='/a'"}, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentElementsOrderByAttributes(t *testing.T) {
	assertT := assert.New(t)

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
	assertT.Equal([]string{"Children order differ for 4 nodes, path='/items'"}, CompareXmlStrings(xmlSample1, xmlSample2, false))
}

func TestDifferentChildren(t *testing.T) {
	assertT := assert.New(t)

	xmlSample1 := `<a><b><c/><c/><d/></b></a>`
	xmlSample2 := `<a><b><d/><e/><e/><e/></b></a>`
	assertT.Equal([]string{"Children differ: counts 3 vs 4: c[0]:+2, e[1]:-3, path='/a/b'"}, CompareXmlStrings(xmlSample1, xmlSample2, false))
	assertT.Equal([]string{"Children differ: counts 4 vs 3: e[1]:+3, c[0]:-2, path='/a/b'"}, CompareXmlStrings(xmlSample2, xmlSample1, false))
}

func TestDifferentChildren2(t *testing.T) {
	assertT := assert.New(t)

	// Edits: DELETE 'c', MODIFY 'd', Add 'e'
	xmlSample1 := `<a><b><c/><d>1</d></b></a>`
	xmlSample2 := `<a><b><d>2</d><e/></b></a>`
	assertT.Equal([]string{"Children differ: counts 2 vs 2: c[0]:+1, e[1]:-1, path='/a/b'", "Node texts differ: '1' vs '2', path='/a/b/d[1]'"},
		CompareXmlStrings(xmlSample1, xmlSample2, false))
	assertT.Equal([]string{"Children differ: counts 2 vs 2: e[1]:+1, c[0]:-1, path='/a/b'", "Node texts differ: '2' vs '1', path='/a/b/d[0]'"},
		CompareXmlStrings(xmlSample2, xmlSample1, false))
}

func TestComputeDifferences(t *testing.T) {
	assertT := assert.New(t)

	recorder := ComputeDifferences(xmlString1, xmlMixed, false, []string{})
	assertT.Equal(3, len(recorder.Diffs))
	diff1 := recorder.Diffs[0]
	diff2 := recorder.Diffs[1]
	diff3 := recorder.Diffs[2]

	assertT.Equal(DiffContent, diff1.GetType())
	assertT.Equal("Node texts differ: '' vs 'Some text ...\n    \n\tmixed with elements', path='/note'", diff1.DescribeDiff())
	assertT.Equal(DiffContent, diff2.GetType())
	assertT.Equal("Node texts differ: 'Tove' vs 'Jani', path='/note/to[0]'", diff2.DescribeDiff())
	assertT.Equal(DiffContent, diff3.GetType())
	assertT.Equal("Node texts differ: 'Jani' vs 'Tove', path='/note/from[1]'", diff3.DescribeDiff())
}

func TestAreEqualNumbers(t *testing.T) {
	assertT := assert.New(t)

	assertT.True(areEqualNumbers("0.2", "0.20"))
	assertT.True(areEqualNumbers("2", "1.9999997"))
	assertT.False(areEqualNumbers("1.2", "1,2"))
	assertT.False(areEqualNumbers("2", "abc"))
}
