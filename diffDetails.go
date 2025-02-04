package xmlcomparator

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/aknopov/handymaps/bimap"
)

type DiffType int

const (
	DiffName DiffType = iota + 1
	DiffSpace
	DiffContent
	DiffAttributes
	DiffChildren
	DiffChildrenOrder
)

type XmlDiff interface {
	DescribeDiff() string
	GetType() DiffType
}

type TextualDiff struct {
	Type    DiffType
	Text1   string
	Text2   string
	XmlPath string
}

type AttributeDiff struct {
	Diffs   []Diff[xml.Attr]
	Len1    int
	Len2    int
	XmlPath string
}

type OrderDiff struct {
	Len     int
	XmlPath string
}

type ChildrenDiff struct {
	Diffs   []Diff[parseNode]
	Len1    int
	Len2    int
	XmlPath string
}

// ------------

func createTextDiff(diffType DiffType, text1 string, text2 string, xmlPath string) *TextualDiff {
	return &TextualDiff{Type: diffType, Text1: text1, Text2: text2, XmlPath: xmlPath}
}

func (diff TextualDiff) DescribeDiff() string {
	switch diff.Type {
	case DiffName:
		return fmt.Sprintf("Node names differ: '%s' vs '%s', path='%s'", diff.Text1, diff.Text2, diff.XmlPath)
	case DiffSpace:
		return fmt.Sprintf("Node namespaces differ: '%s' vs '%s', path='%s'", diff.Text1, diff.Text2, diff.XmlPath)
	case DiffContent:
		return fmt.Sprintf("Node texts differ: '%s' vs '%s', path='%s'", diff.Text1, diff.Text2, diff.XmlPath)
	default:
		panic("Unexpected textual diff type")
	}
}

func (diff TextualDiff) GetType() DiffType {
	return diff.Type
}

// ------------

func createAttributeDiff(diffs []Diff[xml.Attr], len1 int, len2 int, xmlPath string) *AttributeDiff {
	return &AttributeDiff{Diffs: diffs, Len1: len1, Len2: len2, XmlPath: xmlPath}
}

func (diff AttributeDiff) DescribeDiff() string {
	matchingdMap := createMatchingElementsMap(diff.Diffs, attrName)

	unmatchedDiffs := make([]Diff[xml.Attr], 0, len(diff.Diffs)/2)
	for i := 0; i < len(diff.Diffs); i++ {
		if !matchingdMap.ContainsValue(i) && !matchingdMap.ContainsKey(i) {
			unmatchedDiffs = append(unmatchedDiffs, diff.Diffs[i])
		}
	}

	// Log first mismatched attributes...
	var sDiffs string
	if len(unmatchedDiffs) > 0 {
		sDiffs = fmt.Sprintf("counts %d vs %d: %s", diff.Len1, diff.Len2, extractNames(unmatchedDiffs, attrName))
	}
	// ... then matching with different content
	it := matchingdMap.Iterator()
	for it.HasNext() {
		i, j := it.Next()
		attr1 := &diff.Diffs[i].e
		attr2 := &diff.Diffs[j].e
		if len(sDiffs) != 0 {
			sDiffs += ", "
		}
		sDiffs += fmt.Sprintf("'%s=%s' vs '%s=%s'", attrName(attr1), attr1.Value, attrName(attr2), attr2.Value)
	}

	return fmt.Sprintf("Attributes differ: %s, path='%s'", sDiffs, diff.XmlPath)
}

func (diff AttributeDiff) GetType() DiffType {
	return DiffAttributes
}

// ------------

func createOrderDiff(len int, xmlPath string) *OrderDiff {
	return &OrderDiff{Len: len, XmlPath: xmlPath}
}

func (diff OrderDiff) DescribeDiff() string {
	return fmt.Sprintf("Children order differ for %d nodes, path='%s'", diff.Len, diff.XmlPath)
}

func (diff OrderDiff) GetType() DiffType {
	return DiffChildrenOrder
}

// ------------

func createChildrenDiff(diffs []Diff[parseNode], len1 int, len2 int, xmlPath string) *ChildrenDiff {
	return &ChildrenDiff{Diffs: diffs, Len1: len1, Len2: len2, XmlPath: xmlPath}
}

func (diff ChildrenDiff) DescribeDiff() string {
	// return fmt.Sprintf("Children differ: counts %d vs %d, path='%s'", diff.Len1, diff.Len2, diff.XmlPath)
	matchingdMap := createMatchingElementsMap(diff.Diffs, nodeName)

	unmatchedDiffs := make([]Diff[parseNode], 0, len(diff.Diffs)/2)
	for i := 0; i < len(diff.Diffs); i++ {
		if !matchingdMap.ContainsValue(i) && !matchingdMap.ContainsKey(i) {
			unmatchedDiffs = append(unmatchedDiffs, diff.Diffs[i])
		}
	}

	// Log first message for this node
	if len(unmatchedDiffs) > 0 {
		return fmt.Sprintf("Children differ: counts %d vs %d: %s, path='%s'", diff.Len1, diff.Len2,
			extractNames(unmatchedDiffs, nodeName), diff.XmlPath)
	}
	return ""
}

func (diff ChildrenDiff) GetType() DiffType {
	return DiffChildren
}

// ------------

// Matches nodes in diff list there were modified and can be further compared.
// Matching diffs should have complementary edit operation (add/delete) and the same element name.
func createMatchingElementsMap[T any](diffs []Diff[T], namer func(*T) string) *bimap.BiMap[int, int] {
	modifiedMap := bimap.NewBiMapEx[int, int](len(diffs) / 2)

	for i := 0; i < len(diffs); i++ {
		if modifiedMap.ContainsValue(i) {
			continue
		}

		complementDiff := diffAdd
		if diffs[i].t == diffAdd {
			complementDiff = diffDelete
		}

		for j := i + 1; j < len(diffs); j++ {
			if modifiedMap.ContainsValue(j) {
				continue
			}

			if diffs[j].t == complementDiff && namer(&diffs[i].e) == namer(&diffs[j].e) {
				modifiedMap.Put(i, j)
				break
			}
		}
	}

	return modifiedMap
}

func extractNames[T any](mismatchedDiffs []Diff[T], namer func(*T) string) string {
	names := make([]string, 0, len(mismatchedDiffs))

	// First names from the first sample (deleted ones)
	names = append(names, extractNamesByType(mismatchedDiffs, diffDelete, "+", namer)...)
	// Then names from the second sample (added ones)
	names = append(names, extractNamesByType(mismatchedDiffs, diffAdd, "-", namer)...)

	return strings.Join(names, ", ")
}

// Extracts names with run-length "compression"
func extractNamesByType[T any](mismatchedDiffs []Diff[T], diffType editType, sign string, namer func(*T) string) []string {
	names := make([]string, 0)
	var startIdx, dataIdx int
	prevName := ""

	for i := range mismatchedDiffs {
		if mismatchedDiffs[i].t == diffType {
			if prevName == "" {
				dataIdx = mismatchedDiffs[i].aIdx
				prevName = namer(&mismatchedDiffs[i].e)
				startIdx = i
			}
		} else {
			if prevName != "" {
				names = append(names, fmt.Sprintf("%s[%d]:%s%d", prevName, dataIdx, sign, i-startIdx))
			}
			prevName = ""
		}
	}
	if prevName != "" {
		names = append(names, fmt.Sprintf("%s[%d]:%s%d", prevName, dataIdx, sign, len(mismatchedDiffs)-startIdx))
	}

	return names
}
