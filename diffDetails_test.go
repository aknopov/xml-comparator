package xmlcomparator

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreators(t *testing.T) {
	assertT := assert.New(t)

	textDiff := createTextDiff(DiffName, "a", "b", "/")
	assertT.IsType(&TextualDiff{}, textDiff)

	attribDiff := createAttributeDiff([]Diff[xml.Attr]{}, 0, 0, "/")
	assertT.IsType(&AttributeDiff{}, attribDiff)

	orderDiff := createOrderDiff(0, "/")
	assertT.IsType(&OrderDiff{}, orderDiff)

	childrenDiff := createChildrenDiff([]Diff[parseNode]{}, 0, 0, "/")
	assertT.IsType(&ChildrenDiff{}, childrenDiff)
}

func TestDescribeDiff(t *testing.T) {
	assertT := assert.New(t)

	textDiff := createTextDiff(DiffName, "a", "b", "/")
	assertT.Equal("Node names differ: 'a' vs 'b', path='/'", textDiff.DescribeDiff())

	diffs1 := []Diff[xml.Attr]{{e: xml.Attr{Name: xml.Name{Space: "spc", Local: "name"}, Value: "val"}, t: diffSame}}
	attribDiff := createAttributeDiff(diffs1, 0, 0, "/")
	assertT.Equal("Attributes differ: counts 0 vs 0: , path='/'", attribDiff.DescribeDiff())

	orderDiff := createOrderDiff(1, "/")
	assertT.Equal("Children order differ for 1 nodes, path='/'", orderDiff.DescribeDiff())

	diffs2 := []Diff[parseNode]{{e: parseNode{XMLName: xml.Name{Space: "spc", Local: "name"}}, t: diffSame}}
	childrenDiff := createChildrenDiff(diffs2, 0, 0, "/")
	assertT.Equal("Children differ: counts 0 vs 0: , path='/'", childrenDiff.DescribeDiff())
}

func TestGetType(t *testing.T) {
	assertT := assert.New(t)

	tests := []struct {
		diff  XmlDiff
		dType DiffType
	}{
		{createTextDiff(DiffName, "a", "b", "/"), DiffName},
		{createTextDiff(DiffSpace, "a", "b", "/"), DiffSpace},
		{createTextDiff(DiffContent, "a", "b", "/"), DiffContent},
		{createAttributeDiff(make([]Diff[xml.Attr], 0), 0, 0, "/"), DiffAttributes},
		{createOrderDiff(0, "/"), DiffChildrenOrder},
		{createChildrenDiff(make([]Diff[parseNode], 0), 0, 0, "/"), DiffChildren},
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
