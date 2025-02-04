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
	ParseError
)

type XmlDiff interface {
	DescribeDiff() string
	GetType() DiffType
}

type parserError struct {
	text string
}

type textualDiff struct {
	diffType DiffType
	text1    string
	text2    string
	xmlPath  string
}

type attributeDiff struct {
	diffs   []diffT[xml.Attr]
	len1    int
	len2    int
	xmlPath string
}

type orderDiff struct {
	len     int
	xmlPath string
}

type childrenDiff struct {
	diffs   []diffT[parseNode]
	len1    int
	len2    int
	xmlPath string
}

// ------------

func (err parserError) DescribeDiff() string {
	return err.text
}

func (err parserError) GetType() DiffType {
	return ParseError
}

// ------------

func createTextDiff(diffType DiffType, text1 string, text2 string, xmlPath string) *textualDiff {
	return &textualDiff{diffType: diffType, text1: text1, text2: text2, xmlPath: xmlPath}
}

func (diff textualDiff) DescribeDiff() string {
	switch diff.diffType {
	case DiffName:
		return fmt.Sprintf("Node names differ: '%s' vs '%s', path='%s'", diff.text1, diff.text2, diff.xmlPath)
	case DiffSpace:
		return fmt.Sprintf("Node namespaces differ: '%s' vs '%s', path='%s'", diff.text1, diff.text2, diff.xmlPath)
	case DiffContent:
		return fmt.Sprintf("Node texts differ: '%s' vs '%s', path='%s'", diff.text1, diff.text2, diff.xmlPath)
	default:
		panic("Unexpected textual diff type")
	}
}

func (diff textualDiff) GetType() DiffType {
	return diff.diffType
}

// ------------

func createAttributeDiff(diffs []diffT[xml.Attr], len1 int, len2 int, xmlPath string) *attributeDiff {
	return &attributeDiff{diffs: diffs, len1: len1, len2: len2, xmlPath: xmlPath}
}

func (diff attributeDiff) DescribeDiff() string {
	matchingdMap := createMatchingElementsMap(diff.diffs, attrName)

	unmatchedDiffs := make([]diffT[xml.Attr], 0, len(diff.diffs)/2)
	for i := 0; i < len(diff.diffs); i++ {
		if !matchingdMap.ContainsValue(i) && !matchingdMap.ContainsKey(i) {
			unmatchedDiffs = append(unmatchedDiffs, diff.diffs[i])
		}
	}

	// Log first mismatched attributes...
	var sDiffs string
	if len(unmatchedDiffs) > 0 {
		sDiffs = fmt.Sprintf("counts %d vs %d: %s", diff.len1, diff.len2, extractNames(unmatchedDiffs, attrName))
	}
	// ... then matching with different content
	it := matchingdMap.Iterator()
	for it.HasNext() {
		i, j := it.Next()
		attr1 := &diff.diffs[i].e
		attr2 := &diff.diffs[j].e
		if len(sDiffs) != 0 {
			sDiffs += ", "
		}
		sDiffs += fmt.Sprintf("'%s=%s' vs '%s=%s'", attrName(attr1), attr1.Value, attrName(attr2), attr2.Value)
	}

	return fmt.Sprintf("Attributes differ: %s, path='%s'", sDiffs, diff.xmlPath)
}

func (diff attributeDiff) GetType() DiffType {
	return DiffAttributes
}

// ------------

func createOrderDiff(len int, xmlPath string) *orderDiff {
	return &orderDiff{len: len, xmlPath: xmlPath}
}

func (diff orderDiff) DescribeDiff() string {
	return fmt.Sprintf("Children order differ for %d nodes, path='%s'", diff.len, diff.xmlPath)
}

func (diff orderDiff) GetType() DiffType {
	return DiffChildrenOrder
}

// ------------

func createChildrenDiff(diffs []diffT[parseNode], len1 int, len2 int, xmlPath string) *childrenDiff {
	return &childrenDiff{diffs: diffs, len1: len1, len2: len2, xmlPath: xmlPath}
}

func (diff childrenDiff) DescribeDiff() string {
	// return fmt.Sprintf("Children differ: counts %d vs %d, path='%s'", diff.Len1, diff.Len2, diff.XmlPath)
	matchingdMap := createMatchingElementsMap(diff.diffs, nodeName)

	unmatchedDiffs := make([]diffT[parseNode], 0, len(diff.diffs)/2)
	for i := 0; i < len(diff.diffs); i++ {
		if !matchingdMap.ContainsValue(i) && !matchingdMap.ContainsKey(i) {
			unmatchedDiffs = append(unmatchedDiffs, diff.diffs[i])
		}
	}

	// Log first message for this node
	if len(unmatchedDiffs) > 0 {
		return fmt.Sprintf("Children differ: counts %d vs %d: %s, path='%s'", diff.len1, diff.len2,
			extractNames(unmatchedDiffs, nodeName), diff.xmlPath)
	}
	return ""
}

func (diff childrenDiff) GetType() DiffType {
	return DiffChildren
}

// ------------

// Matches nodes in diff list there were modified and can be further compared.
// Matching diffs should have complementary edit operation (add/delete) and the same element name.
func createMatchingElementsMap[T any](diffs []diffT[T], namer func(*T) string) *bimap.BiMap[int, int] {
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

func extractNames[T any](mismatchedDiffs []diffT[T], namer func(*T) string) string {
	names := make([]string, 0, len(mismatchedDiffs))

	// First names from the first sample (deleted ones)
	names = append(names, extractNamesByType(mismatchedDiffs, diffDelete, "+", namer)...)
	// Then names from the second sample (added ones)
	names = append(names, extractNamesByType(mismatchedDiffs, diffAdd, "-", namer)...)

	return strings.Join(names, ", ")
}

// Extracts names with run-length "compression"
func extractNamesByType[T any](mismatchedDiffs []diffT[T], diffType editType, sign string, namer func(*T) string) []string {
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
