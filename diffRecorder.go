package xmlcomparator

import "regexp"

type void struct{}

var empty void

// Discrepancy messages collected while walking the trees.
type DiffRecorder struct {
	ignoredDiscrepancies []*regexp.Regexp
	Messages             []string
	namespaces           map[pair]void
}

// Creates an instance of DiffRecorder.
func CreateDiffRecorder(ignoredDiscrepancies []string) *DiffRecorder {
	regexes := make([]*regexp.Regexp, len(ignoredDiscrepancies))
	for i, d := range ignoredDiscrepancies {
		regexes[i] = regexp.MustCompile(d)
	}

	return &DiffRecorder{
		ignoredDiscrepancies: regexes,
		Messages:             make([]string, 0),
		namespaces:           make(map[pair]void),
	}
}

func (recorder *DiffRecorder) AddMessage(msg string) {
	for _, d := range recorder.ignoredDiscrepancies {
		if d.MatchString(msg) {
			return
		}
	}
	recorder.Messages = append(recorder.Messages, msg)
}

func (recorder *DiffRecorder) AreNamespacesNew(space1 string, space2 string) bool{
	aPair := pair{space1, space2}
	if _, ok := recorder.namespaces[aPair]; ok {
		return false
	}
	recorder.namespaces[aPair] = empty
	return true
}
