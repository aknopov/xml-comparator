package xmlcomparator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKnownMessagesFiltering(t *testing.T) {
	assert := assert.New(t)

	recorder := CreateDiffRecorder([]string{"^footer.*$"})
	recorder.AddMessage("header")
	recorder.AddMessage("body")
	recorder.AddMessage("footer")
	recorder.AddMessage(" footer2")

	assert.Equal([]string{"header", "body", " footer2"}, recorder.Messages)
}
