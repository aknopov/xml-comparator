package xmlcomparator

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/aknopov/handymaps/ordered"
)

const (
	eps = 1.e-6
)

var numberPattern = regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?$`)

type keyValue struct {
	key   string
	value string
}

type nodeId struct {
	name string // name of the node
	hash int    // hash of the node
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

func nodesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool) bool {
	switch {
	case nodeNamesDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return true
	case nodeSpacesDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return true
	case nodesTextDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return true
	case attributesDifferent(node1, node2, diffRecorder) && stopOnFirst:
		return true
	case childrenDifferent(node1, node2, diffRecorder, stopOnFirst):
		return true
	}

	return false
}

func nodeNamesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {
	name1 := node1.XMLName.Local
	name2 := node2.XMLName.Local
	if name1 == name2 {
		return false
	}

	diffRecorder.AddMessage(fmt.Sprintf("Node names differ: '%s' vs '%s', path='%s'", name1, name2, node1.Path()))
	return true
}

func nodeSpacesDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {
	space1 := node1.XMLName.Space
	space2 := node2.XMLName.Space
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

func extractAttributes(node *Node) map[string]string {
	attrs := make(map[string]string, len(node.Attrs))
	for _, attr := range node.Attrs {
		// Namesapce attributes are processed separately
		if attr.Name.Space != "xmlns" && attr.Name.Local != "xmlns" {
			attrs[attr.Name.Local] = attr.Value
		}
	}
	return attrs
}

func childrenDifferent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool) bool {
	childIds1 := childIds(node1)
	childIds2 := childIds(node2)

	if childrenNamesAreEqual(childIds1, childIds2) {
		return childrenDifferentByContent(node1, node2, diffRecorder, stopOnFirst)
	}

	if childrenOrderDifferent(childIds1, childIds2) {
		diffRecorder.AddMessage(fmt.Sprintf("Children order differ for %d nodes, path='%s'", len(childIds1), node1.Path()))
		// no chance to recover from this
		return true
	}

	// Positive values in the map belong only to the 1-st node, negative - only to the 2-nd
	diffMap := ordered.NewOrderedMapEx[nodeId, int](len(childIds1))
	for _, id := range childIds1 {
		diffMap.Compute(id, func(_ nodeId, v int) int { return v + 1 })
	}
	for _, id := range childIds2 {
		diffMap.Compute(id, func(_ nodeId, v int) int { return v - 1 })
	}

	// Leave entries with non-zero values
	diffNames := make([]string, 0, diffMap.Len()/2)
	it := diffMap.Iterator()
	for it.HasNext() {
		k, v := it.Next()
		if v != 0 {
			var childIdx int
			if v > 0 {
				childIdx = findIndex(childIds1, k)
			} else {
				childIdx = findIndex(childIds2, k)
			}
			diffNames = append(diffNames, fmt.Sprintf("%s[%d]:%+d", k.name, childIdx, v))
		}
	}

	if len(diffNames) > 0 {
		// Sort diff names placing first `node1` children for consistent output
		sort.Slice(diffNames, func(i, j int) bool { return isNameLess(diffNames[i], diffNames[j]) })
		diffRecorder.AddMessage(
			fmt.Sprintf("Children differ: %d vs %d (diffs: [%s]), path='%s'", len(childIds1), len(childIds2),
				strings.Join(diffNames, ", "), node1.Path()))
	}

	compareMatchingChildren(node1, node2, diffRecorder, stopOnFirst, *diffMap)

	return true
}

func childrenNamesAreEqual(childIds1, childIds2 []nodeId) bool {
	if len(childIds1) != len(childIds2) {
		return false
	}
	for i := 0; i < len(childIds1); i++ {
		if childIds1[i].name != childIds2[i].name {
			return false
		}
	}
	return true
}

func isNameLess(s1, s2 string) bool {
	switch {
	case strings.Contains(s1, ":+") && strings.Contains(s2, ":-"):
		return true
	case strings.Contains(s1, ":-") && strings.Contains(s2, ":+"):
		return false
	}
	return s1 < s2
}

// Attempts to match children by their names, already knowing that nodes are not equal.
// This is "cheap and cheerful" substitution for a full-blown diff algorithm (Longest Common Subsequence).
func compareMatchingChildren(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool, diffMap ordered.OrderedMap[nodeId, int]) {
	i1 := 0
	i2 := 0
	for i1 < len(node1.Children) && i2 < len(node2.Children) {
		child1 := &node1.Children[i1]
		child2 := &node2.Children[i2]
		id1 := nodeId{child1.XMLName.Local, child1.Hash}
		id2 := nodeId{child2.XMLName.Local, child2.Hash}
		if id1 == id2 {
			// Recursion!
			if nodesDifferent(child1, child2, diffRecorder, stopOnFirst) && stopOnFirst {
				return
			}
			i1++
			i2++
		} else {
			if c, _ := diffMap.Get(id1); c >= 0 {
				i1++
			} else if c, _ := diffMap.Get(id2); c <= 0 {
				i2++
			}
		}
	}
}

func childrenDifferentByContent(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool) bool {
	if len(node1.Children) != len(node2.Children) {
		panic("Children lists have different lengths")
	}

	for i := 0; i < len(node1.Children); i++ {
		// Recursion!
		if nodesDifferent(&node1.Children[i], &node2.Children[i], diffRecorder, stopOnFirst) && stopOnFirst {
			return true
		}
	}

	return false
}

// Checking children order is based on childer names as well as their hashcodes (that reflect attributes, char data, etc)

func isIdLess(nodeId1, nodeId2 nodeId) bool {
	if nodeId1.name < nodeId2.name {
		return true
	} else if nodeId1.name > nodeId2.name {
		return false
	}
	return nodeId1.hash < nodeId2.hash
}

func childIds(node *Node) []nodeId {
	ret := make([]nodeId, 0, len(node.Children))
	for _, child := range node.Children {
		ret = append(ret, nodeId{child.XMLName.Local, hashCode(&child)})
	}
	return ret
}

func childrenOrderDifferent(childIds1 []nodeId, childIds2 []nodeId) bool {
	if len(childIds1) < 2 || len(childIds1) != len(childIds2) || slicesEqual(childIds1, childIds2) {
		return false
	}

	childIds1Sorted := slices.Clone(childIds1)
	childIds2Sorted := slices.Clone(childIds2)
	sort.Slice(childIds1Sorted, func(i, j int) bool { return isIdLess(childIds1[i], childIds1[j]) })
	sort.Slice(childIds1Sorted, func(i, j int) bool { return isIdLess(childIds2[i], childIds2[j]) })
	return slicesEqual(childIds1Sorted, childIds2Sorted)
}
