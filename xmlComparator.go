package xmlcomparator

import (
	"fmt"
	"strings"
)

// Compares two XML strings.
//   - sample1 - first XML string
//   - sample2 - second XML string
//   - stopOnFirst - stop comparison on the first difference
func CompareXmlString(sample1 string, sample2 string, stopOnFirst bool) []string {
	return CompareXmlStringEx(sample1, sample2, stopOnFirst, []string{})
}

// Compares two XML strings.
//   - sample1 - first XML string
//   - sample2 - second XML string
//   - ignoredDiscrepancies - list of regular expressions to ignore discrepancies
//   - stopOnFirst - stop comparison on the first difference
func CompareXmlStringEx(sample1 string, sample2 string, stopOnFirst bool, ignoredDiscrepancies []string) []string {
	root1, err := UnmarshalXML(sample1)
	if root1 == nil || err != nil {
		return []string{err.Error()}
	}

	root2, err := UnmarshalXML(sample2)
	if root2 == nil || err != nil {
		return []string{err.Error()}
	}

	diffRecorder := CreateDiffRecorder(ignoredDiscrepancies)

	nodesEqual(root1, root2, diffRecorder, stopOnFirst)

	return diffRecorder.Messages
}

func nodesEqual(node1 *Node, node2 *Node, diffRecorder *DiffRecorder, stopOnFirst bool) bool {

	orgDiffsCount := len(diffRecorder.Messages)

	if node1.XMLName.Local != node2.XMLName.Local && stopOnFirst {
		diffRecorder.AddMessage(fmt.Sprintf("Node names differ: '%s' vs '%s'", node1.XMLName.Local, node2.XMLName.Local))
		return false
	}
	if node1.XMLName.Space != node2.XMLName.Space && stopOnFirst {
		diffRecorder.AddMessage(fmt.Sprintf("Node names namespaces: '%s' vs '%s'", node1.XMLName.Space, node2.XMLName.Space))
		return false
	}
	if nodesTextDiffer(node1, node2, diffRecorder) && stopOnFirst {
		return false
	}

	return orgDiffsCount == len(diffRecorder.Messages)
}

func nodesTextDiffer(node1 *Node, node2 *Node, diffRecorder *DiffRecorder) bool {

	ownText1 := removeChildrenText(node1)
	ownText2 := removeChildrenText(node2)
	if ownText1 == ownText2 {
		return false
	}

	diffRecorder.AddMessage(fmt.Sprintf("Nodes text differ: '%s' vs '%s'", ownText1, ownText2))
	return true
}

func removeChildrenText(node2 *Node) string {
	orgText := string(node2.Content)
	for _, child := range node2.Children {
		orgText = strings.Replace(orgText, string(child.Content), "", 1)
	}
	return strings.TrimSpace(orgText)
}
