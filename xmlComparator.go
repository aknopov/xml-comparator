package xmlcomparator

import (
	"encoding/xml"
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
//
// Returns:
// A list of detected discrepancies as strings
func CompareXmlStrings(sample1 string, sample2 string, stopOnFirst bool) []string {
	return CompareXmlStringsEx(sample1, sample2, stopOnFirst, []string{})
}

// Compares two XML strings.
//   - sample1 - first XML string
//   - sample2 - second XML string
//   - stopOnFirst - stop comparison on the first difference
//   - ignoredDiscrepancies - list of regular expressions for ignored discrepancies
//
// Returns:
// A list of detected discrepancies as strings
func CompareXmlStringsEx(sample1 string, sample2 string, stopOnFirst bool, ignoredDiscrepancies []string) []string {
	return ComputeDifferences(sample1, sample2, stopOnFirst, ignoredDiscrepancies).GetMessages()
}

// Compares two XML strings.
//   - sample1 - first XML string
//   - sample2 - second XML string
//   - stopOnFirst - stop comparison on the first difference
//   - ignoredDiscrepancies - list of regular expressions for ignored discrepancies
//
// Returns:
// A list of detected discrepancies
func ComputeDifferences(sample1 string, sample2 string, stopOnFirst bool, ignoredDiscrepancies []string) DiffRecorder {
	diffRecorder := createDiffRecorder(ignoredDiscrepancies)

	root1, err := parseXML(sample1)
	if root1 == nil || err != nil {
		diffRecorder.addDiff(parserError{text: "Can't parse the first sample: " + err.Error()})
		return diffRecorder
	}

	root2, err := parseXML(sample2)
	if root2 == nil || err != nil {
		diffRecorder.addDiff(parserError{text: "Can't parse the second sample: " + err.Error()})
		return diffRecorder
	}

	nodesDifferent(root1, root2, diffRecorder, stopOnFirst)

	return diffRecorder
}

func nodesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *diffRecorder, stopOnFirst bool) {
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

func nodeNamesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *diffRecorder) bool {
	name1 := nodeName(node1)
	name2 := nodeName(node2)
	if name1 == name2 {
		return false
	}

	diffRecorder.addDiff(createTextDiff(DiffName, name1, name2, node1.path()))
	return true
}

func nodeSpacesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *diffRecorder) bool {
	space1 := nodeSpace(node1)
	space2 := nodeSpace(node2)
	if space1 == space2 || space1 == "" || space2 == "" {
		return false
	}

	if diffRecorder.areNamespacesNew(space1, space2) {
		diffRecorder.addDiff(createTextDiff(DiffSpace, space1, space2, node1.path()))
	}
	return true
}
func nodesTextDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *diffRecorder) bool {
	ownText1 := strings.TrimSpace(node1.CharData)

	ownText2 := strings.TrimSpace(node2.CharData)
	if ownText1 == ownText2 || areEqualNumbers(ownText1, ownText2) {
		return false
	}

	diffRecorder.addDiff(createTextDiff(DiffContent, ownText1, ownText2, node1.path()))
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

func attributesDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *diffRecorder) bool {
	attrs1 := node1.extractAttributes()
	attrs2 := node2.extractAttributes()
	if slices.Equal(attrs1, attrs2) || slices.Equal(sorted(attrs1, attrComparator), sorted(attrs2, attrComparator)) {
		return false
	}

	diffs := compareSequences(attrs1, attrs2, func(a, b xml.Attr) bool { return a == b })
	diffRecorder.addDiff(createAttributeDiff(diffs, len(attrs1), len(attrs2), node1.path()))

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

func childrenDifferent(node1 *parseNode, node2 *parseNode, diffRecorder *diffRecorder, stopOnFirst bool) bool {
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
			diffRecorder.addDiff(createOrderDiff(len(hashes1), node1.path()))
			// TODO Implement comparison and output of sorted children
			return true
		}
	}

	diffs := compareSequences(node1.Children, node2.Children, func(a, b parseNode) bool { return a.Hash == b.Hash })

	diffRecorder.addDiff(createChildrenDiff(diffs, len(node1.Children), len(node2.Children), node1.path()))

	matchingdMap := createMatchingElementsMap(diffs, nodeName)
	// Recursion!
	iterateMatchingNodes(matchingdMap, diffs, diffRecorder, stopOnFirst)

	return true
}

func extractChildHashes(node *parseNode) []uint32 {
	hashes := make([]uint32, len(node.Children))
	for i := range node.Children {
		hashes[i] = node.Children[i].Hash
	}
	return hashes
}

func iterateMatchingNodes(matchingMap *bimap.BiMap[int, int], diffs []diffT[parseNode], diffRecorder *diffRecorder, stopOnFirst bool) {
	it := matchingMap.Iterator()
	for it.HasNext() {
		i, j := it.Next()
		nodesDifferent(&diffs[i].e, &diffs[j].e, diffRecorder, stopOnFirst)
	}
}

func sorted[T comparable](slice []T, isLess func(T, T) bool) []T {
	ret := make([]T, len(slice))
	copy(ret, slice)
	sort.Slice(ret, func(i, j int) bool { return isLess(ret[i], ret[j]) })
	return ret
}
