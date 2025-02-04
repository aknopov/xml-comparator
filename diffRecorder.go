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

// Provides diocreapncies while walking the trees in raw and string formats.
type DiffRecorder interface {
	// List of differences
	GetDiffs() []XmlDiff
	// List of serialized differences
	GetMessages() []string
}

// Discrepancy messages collected while walking the trees.
type diffRecorder struct {
	ignoredDiscrepancies []*regexp.Regexp
	diffs                []XmlDiff
	messages             []string
	namespaces           map[keyValue]void
}

func (recorder diffRecorder) GetDiffs() []XmlDiff {
	return recorder.diffs
}

func (recorder diffRecorder) GetMessages() []string {
	return recorder.messages
}

// Creates an instance of DiffRecorder.
func createDiffRecorder(ignoredDiscrepancies []string) *diffRecorder {
	regexes := make([]*regexp.Regexp, len(ignoredDiscrepancies))
	for i := range ignoredDiscrepancies {
		regexes[i] = regexp.MustCompile(ignoredDiscrepancies[i])
	}

	return &diffRecorder{
		ignoredDiscrepancies: regexes,
		diffs:                make([]XmlDiff, 0),
		messages:             make([]string, 0),
		namespaces:           make(map[keyValue]void),
	}
}

func (recorder *diffRecorder) addDiff(diff XmlDiff) {
	msg := diff.DescribeDiff()
	if len(msg) != 0 && !recorder.isIgnored(msg) {
		recorder.diffs = append(recorder.diffs, diff)
		recorder.messages = append(recorder.messages, msg)
	}
}

func (recorder *diffRecorder) isIgnored(msg string) bool {
	for _, d := range recorder.ignoredDiscrepancies {
		if d.MatchString(msg) {
			return true
		}
	}
	return false
}

func (recorder *diffRecorder) areNamespacesNew(space1 string, space2 string) bool {
	aPair := keyValue{space1, space2}
	if _, ok := recorder.namespaces[aPair]; ok {
		return false
	}
	recorder.namespaces[aPair] = empty
	return true
}
