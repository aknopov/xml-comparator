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

func (diff testDiff) XmlPath() string {
	return "/a/b"
}

func TestKnownMessagesFiltering(t *testing.T) {
	assertT := assert.New(t)

	recorder := createDiffRecorder([]string{"^footer.*$"})
	recorder.addDiff(testDiff{"header"})
	recorder.addDiff(testDiff{"body"})
	recorder.addDiff(testDiff{"footer"})
	recorder.addDiff(testDiff{" footer2"})

	assertT.Equal([]string{"header", "body", " footer2"}, recorder.Messages)
}

func TestAreNamespacesNew(t *testing.T) {
	assertT := assert.New(t)

	recorder := createDiffRecorder([]string{"^footer.*$"})

	assertT.True(recorder.areNamespacesNew("space1", "space2"))
	assertT.False(recorder.areNamespacesNew("space1", "space2"))
}

func TestAccessToDetails(t *testing.T) {
	assertT := assert.New(t)

	recorder := createDiffRecorder([]string{})
	recorder.addDiff(testDiff{"header"})
	recorder.addDiff(testDiff{"body"})

	assertT.Equal(2, len(recorder.Diffs))
	diff := recorder.Diffs[0]
	assertT.Equal("header", diff.DescribeDiff())
	assertT.Equal("/a/b", diff.XmlPath())
	diff = recorder.Diffs[1]
	assertT.Equal("/a/b", diff.XmlPath())
}
