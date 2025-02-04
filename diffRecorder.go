package xmlcomparator

import (
	"regexp"
)

type void struct{}

var empty void

type keyValue struct {
	key   string
	value string
}

// Discrepancy messages collected while walking the trees.
type DiffRecorder struct {
	ignoredDiscrepancies []*regexp.Regexp
	Diffs                []XmlDiff
	Messages             []string
	namespaces           map[keyValue]void
}

// Creates an instance of DiffRecorder.
func CreateDiffRecorder(ignoredDiscrepancies []string) *DiffRecorder {
	regexes := make([]*regexp.Regexp, len(ignoredDiscrepancies))
	for i := range ignoredDiscrepancies {
		regexes[i] = regexp.MustCompile(ignoredDiscrepancies[i])
	}

	return &DiffRecorder{
		ignoredDiscrepancies: regexes,
		Diffs:                make([]XmlDiff, 0),
		Messages:             make([]string, 0),
		namespaces:           make(map[keyValue]void),
	}
}

func (recorder *DiffRecorder) AddDiff(diff XmlDiff) {
	msg := diff.DescribeDiff()
	if len(msg) != 0 && !recorder.isIgnored(msg) {
		recorder.Diffs = append(recorder.Diffs, diff)
		recorder.Messages = append(recorder.Messages, msg)
	}
}

func (recorder *DiffRecorder) isIgnored(msg string) bool {
	for _, d := range recorder.ignoredDiscrepancies {
		if d.MatchString(msg) {
			return true
		}
	}
	return false
}

func (recorder *DiffRecorder) AreNamespacesNew(space1 string, space2 string) bool {
	aPair := keyValue{space1, space2}
	if _, ok := recorder.namespaces[aPair]; ok {
		return false
	}
	recorder.namespaces[aPair] = empty
	return true
}
