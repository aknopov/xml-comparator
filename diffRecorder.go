package xmlcomparator

import "regexp"

// Discrepancy messages collected while walking the trees.
type DiffRecorder struct {
	ignoredDiscrepancies []*regexp.Regexp
	Messages             []string
}

// Creates an instance of DiffRecorder.
func CreateDiffRecorder(ignoredDiscrepancies []string) *DiffRecorder {
	regs := make([]*regexp.Regexp, len(ignoredDiscrepancies))
	for i, d := range ignoredDiscrepancies {
		regs[i] = regexp.MustCompile(d)
	}

	return &DiffRecorder{
		ignoredDiscrepancies: regs,
		Messages:             make([]string, 0),
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
