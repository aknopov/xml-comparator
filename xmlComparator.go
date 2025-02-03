package xmlcomparator

import (
	"encoding/xml"
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/aknopov/handymaps/bimap"
)

const (
	eps = 1.e-6
)

var numberPattern = regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?$`)

var hashComparator = func(x, y uint32) bool { return x < y }
var attrComparator = func(x, y xml.Attr) bool { return attrName(&x) < attrName(&y) }

// Compares two XML strings.
//   - sample1 - first XML string
//   - sample2 - second XML string
//   - stopOnFirst - stop comparison on the first difference
func CompareXmlStrings(sample1 string, sample2 string, stopOnFirst bool) []string {
	return CompareXmlStringsEx(sample1, sample2, stopOnFirst, []string{})
}

// Compares two XML strings.
//   - sample1 - first XML string
//   - sample2 - second XML string
//   - stopOnFirst - stop comparison on the first difference
//   - ignoredDiscrepancies - list of regular expressions for ignored discrepancies
func CompareXmlStringsEx(sample1 string, sample2 string, stopOnFirst bool, ignoredDiscrepancies []string) []string {
	root1, err := parseXML(sample1)
	if root1 == nil || err != nil {
		return []string{"Can't parse the first sample: " + err.Error()}
	}

	root2, err := parseXML(sample2)
	if root2 == nil || err != nil {
		return []string{"Can't parse the second sample: " + err.Error()}
	}

	diffRecorder := CreateDiffRecorder(ignoredDiscrepancies)

	nodesDifferent(root1, root2, diffRecorder, stopOnFirst)

	return diffRecorder.Messages
}

func nodesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *DiffRecorder, stopOnFirst bool) {
	switch {
	case nodeNamesDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return
	case nodeSpacesDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return
	case nodesTextDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return
	case attributesDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return
	case childrenDifferent(node1, node2, diffRecorder, stopOnFirst):
		return
	}
}

func nodeNamesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *DiffRecorder) bool {
	name1 := nodeName(node1)
	name2 := nodeName(node2)
	if name1 == name2 {
		return false
	}

	diffRecorder.AddMessage("Node names differ: '%s' vs '%s', path='%s'", name1, name2, node1.Path())
	return true
}

func nodeSpacesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *DiffRecorder) bool {
	space1 := nodeSpace(node1)
	space2 := nodeSpace(node2)
	if space1 == space2 || space1 == "" || space2 == "" {
		return false
	}

	if diffRecorder.AreNamespacesNew(space1, space2) {
		diffRecorder.AddMessage("Node namespaces differ: '%s' vs '%s', path='%s'", space1, space2, node1.Path())
	}
	return true
}
func nodesTextDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *DiffRecorder) bool {
	ownText1 := strings.TrimSpace(node1.CharData)

	ownText2 := strings.TrimSpace(node2.CharData)
	if ownText1 == ownText2 || areEqualNumbers(ownText1, ownText2) {
		return false
	}

	diffRecorder.AddMessage("Nodes text differ: '%s' vs '%s', path='%s'", ownText1, ownText2, node1.Path())
	return true
}

func areEqualNumbers(text1, text2 string) bool {
	if numberPattern.MatchString(text1) && numberPattern.MatchString(text2) {
		val1, _ := strconv.ParseFloat(text1, 32)
		val2, _ := strconv.ParseFloat(text2, 32)
		return math.Abs(val2-val1) <= eps*(math.Abs(val2)+math.Abs(val1)+eps)
	}
	return false
}

func attributesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *DiffRecorder) bool {
	attrs1 := node1.extractAttributes()
	attrs2 := node2.extractAttributes()
	if slices.Equal(attrs1, attrs2) || slices.Equal(sorted(attrs1, attrComparator), sorted(attrs2, attrComparator)) {
		return false
	}

	diffs := CompareSequences(attrs1, attrs2, func(a, b xml.Attr) bool { return a == b })

	matchingdMap := createMatchingNodesMap(diffs, attrName)

	unmatchedDiffs := make([]Diff[xml.Attr], 0, len(diffs)/2)
	for i := 0; i < len(diffs); i++ {
		if !matchingdMap.ContainsValue(i) && !matchingdMap.ContainsKey(i) {
			unmatchedDiffs = append(unmatchedDiffs, diffs[i])
		}
	}

	// Log first mismatched attributes
	var sDiffs string
	if len(unmatchedDiffs) > 0 {
		sDiffs = fmt.Sprintf("counts %d vs %d: %s", len(attrs1), len(attrs2), extractNames(unmatchedDiffs, attrName))
	}
	it := matchingdMap.Iterator()
	for it.HasNext() {
		i, j := it.Next()
		attr1 := &diffs[i].e
		attr2 := &diffs[j].e
		if len(sDiffs) != 0 {
			sDiffs += ", "
		}
		sDiffs += fmt.Sprintf("'%s=%s' vs '%s=%s'", attrName(attr1), attr1.Value, attrName(attr2), attr2.Value)
	}

	diffRecorder.AddMessage("Attributes differ: %s, path='%s'", sDiffs, node1.Path())

	return true
}

func (node *parseNode) extractAttributes() []xml.Attr {
	attrs := make([]xml.Attr, 0, len(node.Attrs))
	for i := range node.Attrs {
		// Namesapce attributes are processed separately
		if !isNameSpaceAttr(&node.Attrs[i]) {
			attrs = append(attrs, node.Attrs[i])
		}
	}
	return attrs
}

func childrenDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *DiffRecorder, stopOnFirst bool) bool {
	// Simple case - identical children by hash
	hashes1 := extractChildHashes(node1)
	hashes2 := extractChildHashes(node2)
	if slices.Equal(hashes1, hashes2) {
		return false
	}

	// Simple case - permutation of children
	if len(hashes1) == len(hashes2) {
		sortedHashes1 := sorted(hashes1, hashComparator)
		sortedHashes2 := sorted(hashes2, hashComparator)
		if slices.Equal(sortedHashes1, sortedHashes2) {
			diffRecorder.AddMessage("Children order differ for %d nodes, path='%s'", len(hashes1), node1.Path())
			// TODO Implement comparison and output of sorted children
			return true
		}
	}

	diffs := CompareSequences(node1.Children, node2.Children, func(a, b parseNode) bool { return a.Hash == b.Hash })
	compareMatchingChildren(node1, node2, diffs, diffRecorder, stopOnFirst)

	return true
}

func extractChildHashes(node *parseNode) []uint32 {
	hashes := make([]uint32, len(node.Children))
	for i := range node.Children {
		hashes[i] = node.Children[i].Hash
	}
	return hashes
}

func compareMatchingChildren(node1 *parseNode, node2 *parseNode, diffs []Diff[parseNode], diffRecorder *DiffRecorder, stopOnFirst bool) {
	matchingdMap := createMatchingNodesMap(diffs, nodeName)

	unmatchedDiffs := make([]Diff[parseNode], 0, len(diffs)/2)
	for i := 0; i < len(diffs); i++ {
		if !matchingdMap.ContainsValue(i) && !matchingdMap.ContainsKey(i) {
			unmatchedDiffs = append(unmatchedDiffs, diffs[i])
		}
	}

	// Log first message for this node
	if len(unmatchedDiffs) > 0 {
		diffRecorder.AddMessage("Children differ: counts %d vs %d: %s, path='%s'", len(node1.Children), len(node2.Children),
			extractNames(unmatchedDiffs, nodeName), node1.Path())
	}

	// Recursion!
	iterateMatchingNodes(matchingdMap, diffs, diffRecorder, stopOnFirst)
}

// Matches nodes in diff list there were modified and can be further compared.
// Matching diffs should have complementary edit operation (add/delete) and the same element name.
func createMatchingNodesMap[T any](diffs []Diff[T], namer func(*T) string) *bimap.BiMap[int, int] {
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

func iterateMatchingNodes(matchingMap *bimap.BiMap[int, int], diffs []Diff[parseNode], diffRecorder *DiffRecorder, stopOnFirst bool) {
	it := matchingMap.Iterator()
	for it.HasNext() {
		i, j := it.Next()
		nodesDifferent(&diffs[i].e, &diffs[j].e, diffRecorder, stopOnFirst)
	}
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

func sorted[T comparable](slice []T, isLess func(T, T) bool) []T {
	ret := make([]T, len(slice))
	copy(ret, slice)
	sort.Slice(ret, func(i, j int) bool { return isLess(ret[i], ret[j]) })
	return ret
}
