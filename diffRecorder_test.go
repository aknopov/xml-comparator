package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testDiff struct {
	text string
}

func (diff testDiff) DescribeDiff() string {
	return diff.text
}

func (diff testDiff) GetType() DiffType {
	return DiffContent
}

func TestKnownMessagesFiltering(t *testing.T) {
	assertT := assert.New(t)

	recorder := CreateDiffRecorder([]string{"^footer.*$"})
	recorder.AddDiff(testDiff{"header"})
	recorder.AddDiff(testDiff{"body"})
	recorder.AddDiff(testDiff{"footer"})
	recorder.AddDiff(testDiff{" footer2"})

	assertT.Equal([]string{"header", "body", " footer2"}, recorder.Messages)
}

func TestAreNamespacesNew(t *testing.T) {
	assertT := assert.New(t)

	recorder := CreateDiffRecorder([]string{"^footer.*$"})

	assertT.True(recorder.AreNamespacesNew("space1", "space2"))
	assertT.False(recorder.AreNamespacesNew("space1", "space2"))
}

func TestAccessToDetails(t *testing.T) {
	assertT := assert.New(t)

	recorder := CreateDiffRecorder([]string{})
	recorder.AddDiff(testDiff{"header"})
	recorder.AddDiff(testDiff{"body"})

	assertT.Equal(2, len(recorder.Diffs))
	diff := recorder.Diffs[0]
	assertT.Equal("header", diff.DescribeDiff())
	diff = recorder.Diffs[1]
	assertT.Equal("body", diff.DescribeDiff())
}
