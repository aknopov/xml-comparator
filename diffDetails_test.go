package xmlcomparator

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreators(t *testing.T) {
	assertT := assert.New(t)

	textDiff := createTextDiff(DiffName, "a", "b", "/")
	assertT.IsType(&textualDiff{}, textDiff)

	attribDiff := createAttributeDiff([]diffT[xml.Attr]{}, 0, 0, "/")
	assertT.IsType(&attributeDiff{}, attribDiff)

	ordrDiff := createOrderDiff(0, "/")
	assertT.IsType(&orderDiff{}, ordrDiff)

	childDiff := createChildrenDiff([]diffT[parseNode]{}, 0, 0, "/")
	assertT.IsType(&childrenDiff{}, childDiff)
}

func TestDescribeDiff(t *testing.T) {
	assertT := assert.New(t)

	parseError := parserError{"some error"}
	assertT.Equal("some error", parseError.DescribeDiff())

	textDiff := createTextDiff(DiffName, "a", "b", "/")
	assertT.Equal("Node names differ: 'a' vs 'b', path='/'", textDiff.DescribeDiff())

	diffs1 := []diffT[xml.Attr]{{e: xml.Attr{Name: xml.Name{Space: "spc", Local: "name"}, Value: "val"}, t: diffSame}}
	attribDiff := createAttributeDiff(diffs1, 0, 0, "/")
	assertT.Equal("Attributes differ: counts 0 vs 0: , path='/'", attribDiff.DescribeDiff())

	ordrDiff := createOrderDiff(1, "/")
	assertT.Equal("Children order differ for 1 nodes, path='/'", ordrDiff.DescribeDiff())

	diffs2 := []diffT[parseNode]{{e: parseNode{XMLName: xml.Name{Space: "spc", Local: "name"}}, t: diffSame}}
	childDiff := createChildrenDiff(diffs2, 0, 0, "/")
	assertT.Equal("Children differ: counts 0 vs 0: , path='/'", childDiff.DescribeDiff())
}

func TestGetType(t *testing.T) {
	assertT := assert.New(t)

	tests := []struct {
		diff  XmlDiff
		dType DiffType
	}{
		{&parserError{"some error"}, ParseError},
		{createTextDiff(DiffName, "a", "b", "/"), DiffName},
		{createTextDiff(DiffSpace, "a", "b", "/"), DiffSpace},
		{createTextDiff(DiffContent, "a", "b", "/"), DiffContent},
		{createAttributeDiff(make([]diffT[xml.Attr], 0), 0, 0, "/"), DiffAttributes},
		{createOrderDiff(0, "/"), DiffChildrenOrder},
		{createChildrenDiff(make([]diffT[parseNode], 0), 0, 0, "/"), DiffChildren},
	}

	for _, tt := range tests {
		assertT.Equal(tt.dType, tt.diff.GetType())
	}
}

func TestInvalidDescribeDiff(t *testing.T) {
	assertT := assert.New(t)

	invalidDiff := createTextDiff(DiffChildren, "a", "b", "/")
	assertT.Panics(func() { invalidDiff.DescribeDiff() })
}
