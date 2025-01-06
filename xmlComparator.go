package xmlcomparator

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/aknopov/handymaps/bimap"
)

const (
	eps = 1.e-6
)

var numberPattern = regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?$`)

type keyValue struct {
	key   string
	value string
}

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
	root1, err := UnmarshalXML(sample1)
	if root1 == nil || err != nil {
		return []string{"Can't parse first sample: " + err.Error()}
	}

	root2, err := UnmarshalXML(sample2)
	if root2 == nil || err != nil {
		return []string{"Can't parse second sample: " + err.Error()}
	}

	diffRecorder := CreateDiffRecorder(ignoredDiscrepancies)

	nodesDifferent(root1, root2, diffRecorder, stopOnFirst)

	return diffRecorder.Messages
}

func nodesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool) {
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

func nodeNamesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {
	name1 := node1.Name()
	name2 := node2.Name()
	if name1 == name2 {
		return false
	}

	diffRecorder.AddMessage(fmt.Sprintf("Node names differ: '%s' vs '%s', path='%s'", name1, name2, node1.Path()))
	return true
}

func nodeSpacesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {
	space1 := node1.Space()
	space2 := node2.Space()
	if space1 == space2 || space1 == "" || space2 == "" {
		return false
	}

	if diffRecorder.AreNamespacesNew(space1, space2) {
		diffRecorder.AddMessage(fmt.Sprintf("Node namespaces differ: '%s' vs '%s', path='%s'", space1, space2, node1.Path()))
	}
	return true
}
func nodesTextDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {
	ownText1 := strings.TrimSpace(node1.CharData)

	ownText2 := strings.TrimSpace(node2.CharData)
	if ownText1 == ownText2 || areEqualNumbers(ownText1, ownText2) {
		return false
	}

	diffRecorder.AddMessage(fmt.Sprintf("Nodes text differ: '%s' vs '%s', path='%s'", ownText1, ownText2, node1.Path()))
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

func attributesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {
	if len(node1.Attrs) != len(node2.Attrs) {
		diffRecorder.AddMessage(fmt.Sprintf("Attributes count differ: %d vs %d, path='%s'", len(node1.Attrs), len(node2.Attrs), node1.Path()))
		return false
	}

	attrMap1 := extractAttributes(node1)
	attrMap2 := extractAttributes(node2)

	unique1 := make([]keyValue, 0)
	unique2 := make([]keyValue, 0)

	for k, v1 := range attrMap1 {
		v2, ok := attrMap2[k]
		if !ok {
			unique1 = append(unique1, keyValue{k, v1})
		}
		if v1 != v2 {
			unique1 = append(unique1, keyValue{k, v1})
			unique2 = append(unique2, keyValue{k, v2})
		}
	}

	if len(unique1) == 0 && len(unique2) == 0 {
		return false
	}

	diffRecorder.AddMessage(fmt.Sprintf("Attributes differ: '%v' vs '%v', path='%s'", keyValueToString(unique1),
		keyValueToString(unique2), node1.Path()))
	return true
}

func keyValueToString(pairs []keyValue) string {
	ret := "["
	for i, p := range pairs {
		ret += p.key + "=" + p.value
		if i < len(pairs)-1 {
			ret += ", "
		}
	}
	return ret + "]"
}

func childrenDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool) bool {
	// Simple case - identical children by hash
	hashes1 := extractChildHashes(node1)
	hashes2 := extractChildHashes(node2)
	if slices.Equal(hashes1, hashes2) {
		return false
	}

	// Simple case - permutation of children
	if len(hashes1) == len(hashes2) {
		sortedHashes1 := sortedClone(hashes1, func(x, y uint32) bool { return x < y })
		sortedHashes2 := sortedClone(hashes2, func(x, y uint32) bool { return x < y })
		if slices.Equal(sortedHashes1, sortedHashes2) {
			diffRecorder.AddMessage(fmt.Sprintf("Children order differ for %d nodes, path='%s'", len(hashes1), node1.Path()))
			// TODO Implement comparison and output of sorted children
			return true
		}
	}

	diffs := CompareSequences(node1.Children, node2.Children, func(a, b Node) bool { return a.Hash == b.Hash })
	compareMatchingChildren(node1, node2, diffs, diffRecorder, stopOnFirst)

	return true
}

func extractChildHashes(node *Node) []uint32 {
	hashes := make([]uint32, len(node.Children))
	for i := range node.Children {
		hashes[i] = node.Children[i].Hash
	}
	return hashes
}

func compareMatchingChildren(node1 *Node, node2 *Node, diffs []Diff[Node], diffRecorder *DiffRecorder, stopOnFirst bool) {
	matchingMap := createMatchingMap(diffs)

	mismatchedDiffs := make([]Diff[Node], 0, len(diffs)/2)
	for i := 0; i < len(diffs); i++ {
		if !matchingMap.ContainsValue(i) && !matchingMap.ContainsKey(i) {
			mismatchedDiffs = append(mismatchedDiffs, diffs[i])
		}
	}

	// Log first message for this node
	if len(mismatchedDiffs) > 0 {
		diffRecorder.AddMessage(
			fmt.Sprintf("Children differ: counts %d vs %d (diffs: %s), path='%s'", len(node1.Children), len(node2.Children),
				extractNames(mismatchedDiffs), node1.Path()))
	}

	// Recursion!
	iterateMatchingDiffs(matchingMap, diffs, diffRecorder, stopOnFirst)
}

// Matches nodes in diff list by their names and modification operation.
func createMatchingMap(diffs []Diff[Node]) *bimap.BiMap[int, int] {
	matchingMap := bimap.NewBiMapEx[int, int](len(diffs) / 2)

	for i := 0; i < len(diffs); i++ {
		if matchingMap.ContainsValue(i) {
			continue
		}

		complementDiff := DiffAdd
		if diffs[i].t == DiffAdd {
			complementDiff = DiffDelete
		}

		for j := i + 1; j < len(diffs); j++ {
			if diffs[j].t == complementDiff && diffs[i].aIdx == diffs[j].bIdx && diffs[i].e.Name() == diffs[j].e.Name() {
				matchingMap.Put(i, j)
				break
			}
		}
	}

	return matchingMap
}

func iterateMatchingDiffs(matchingMap *bimap.BiMap[int, int], diffs []Diff[Node], diffRecorder *DiffRecorder, stopOnFirst bool) {
	it := matchingMap.Iterator()
	for it.HasNext() {
		i, j := it.Next()
		nodesDifferent(&diffs[i].e, &diffs[j].e, diffRecorder, stopOnFirst)
	}
}

// Extracts names with run-length "compression"
func extractNames(mismatchedDiffs []Diff[Node]) string {
	names := make([]string, 0, len(mismatchedDiffs))

	// First names from the first sample (deleted ones)
	names = append(names, extractNamesByType(mismatchedDiffs, DiffDelete, "+")...)
	// Then names from the second sample (added ones)
	names = append(names, extractNamesByType(mismatchedDiffs, DiffAdd, "-")...)

	return strings.Join(names, ", ")
}

func extractNamesByType(mismatchedDiffs []Diff[Node], diffType DiffType, sign string) []string {
	names := make([]string, 0)
	var startIdx, dataIdx int
	prevName := ""

	for i := range mismatchedDiffs {
		if mismatchedDiffs[i].t == diffType {
			if prevName == "" {
				dataIdx = getDataIdx(mismatchedDiffs[i], diffType)
				prevName = mismatchedDiffs[i].e.Name()
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

func getDataIdx(diff Diff[Node], diffType DiffType) int {
	if diffType == DiffDelete {
		return diff.aIdx
	}
	return diff.bIdx
}
