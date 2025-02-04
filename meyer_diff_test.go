package xmlcomparator

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func equalsFun[T comparable](x, y T) bool {
	return x == y
}

func TestSerialization(t *testing.T) {
	assert := assert.New(t)

	diffs := CompareSequencesEx([]rune("abc"), []rune("abd"), equalsFun, true, 13)
	sDiff := SerializeDiffs(diffs)
	assert.Equal("=97[0<->0]\n=98[1<->1]\n-99[2<->2]\n+100[2<->2]\n", sDiff)
}

func TestStringDiff(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name string
		a    string
		b    string
		diff []Diff[rune]
	}{
		{
			name: "string diff1",
			a:    "abc",
			b:    "abd",
			diff: []Diff[rune]{
				{e: 'c', t: diffDelete, aIdx: 2, bIdx: 2},
				{e: 'd', t: diffAdd, aIdx: 2, bIdx: 2},
			},
		},
		{
			name: "string diff2",
			a:    "abcdef",
			b:    "dacfea",
			diff: []Diff[rune]{
				{e: 'd', t: diffAdd, aIdx: 0, bIdx: 0},
				{e: 'b', t: diffDelete, aIdx: 1, bIdx: 1},
				{e: 'd', t: diffDelete, aIdx: 3, bIdx: 3},
				{e: 'e', t: diffDelete, aIdx: 4, bIdx: 4},
				{e: 'e', t: diffAdd, aIdx: 4, bIdx: 4},
				{e: 'a', t: diffAdd, aIdx: 5, bIdx: 5},
			},
		},
		{
			name: "string diff3",
			a:    "acbdeacbed",
			b:    "acebdabbabed",
			diff: []Diff[rune]{
				{e: 'e', t: diffAdd, aIdx: 2, bIdx: 2},
				{e: 'e', t: diffDelete, aIdx: 4, bIdx: 4},
				{e: 'c', t: diffDelete, aIdx: 6, bIdx: 6},
				{e: 'b', t: diffAdd, aIdx: 7, bIdx: 7},
				{e: 'a', t: diffAdd, aIdx: 8, bIdx: 8},
				{e: 'b', t: diffAdd, aIdx: 9, bIdx: 9},
			},
		},
		{
			name: "string diff4",
			a:    "acebdabbabed",
			b:    "acbdeacbed",
			diff: []Diff[rune]{
				{e: 'e', t: diffDelete, aIdx: 2, bIdx: 2},
				{e: 'e', t: diffAdd, aIdx: 4, bIdx: 4},
				{e: 'c', t: diffAdd, aIdx: 6, bIdx: 6},
				{e: 'b', t: diffDelete, aIdx: 7, bIdx: 7},
				{e: 'a', t: diffDelete, aIdx: 8, bIdx: 8},
				{e: 'b', t: diffDelete, aIdx: 9, bIdx: 9},
			},
		},
		{
			name: "string diff5",
			a:    "abcbda",
			b:    "bdcaba",
			diff: []Diff[rune]{
				{e: 'a', t: diffDelete, aIdx: 0, bIdx: 0},
				{e: 'd', t: diffAdd, aIdx: 1, bIdx: 1},
				{e: 'a', t: diffAdd, aIdx: 3, bIdx: 3},
				{e: 'd', t: diffDelete, aIdx: 4, bIdx: 4},
			},
		},
		{
			name: "string diff6",
			a:    "bokko",
			b:    "bokkko",
			diff: []Diff[rune]{
				{e: 'k', t: diffAdd, aIdx: 4, bIdx: 4},
			},
		},
		{
			name: "string diff7",
			a:    "abcaaaaaabd",
			b:    "abdaaaaaabc",
			diff: []Diff[rune]{
				{e: 'c', t: diffDelete, aIdx: 2, bIdx: 2},
				{e: 'd', t: diffAdd, aIdx: 2, bIdx: 2},
				{e: 'd', t: diffDelete, aIdx: 10, bIdx: 10},
				{e: 'c', t: diffAdd, aIdx: 10, bIdx: 10},
			},
		},
		{
			name: "empty string diff1",
			a:    "",
			b:    "",
			diff: []Diff[rune]{},
		},
		{
			name: "empty string diff2",
			a:    "a",
			b:    "",
			diff: []Diff[rune]{
				{e: 'a', t: diffDelete, aIdx: 0, bIdx: 0},
			},
		},
		{
			name: "empty string diff3",
			a:    "",
			b:    "b",
			diff: []Diff[rune]{
				{e: 'b', t: diffAdd, aIdx: 0, bIdx: 0},
			},
		},
		{
			name: "Unicode string diff",
			a:    "Привет!",
			b:    "Прювет!",
			diff: []Diff[rune]{
				{e: 'и', t: diffDelete, aIdx: 2, bIdx: 2},
				{e: 'ю', t: diffAdd, aIdx: 2, bIdx: 2},
			},
		},
	}

	for _, tt := range tests {
		diffs := CompareSequences([]rune(tt.a), []rune(tt.b), equalsFun)
		assert.True(slices.Equal(tt.diff, diffs), ":%s:diff: want: %v, got: %v", tt.name, tt.diff, diffs)
	}
}

func TestSliceDiff(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		name string
		a    []int
		b    []int
		diff []Diff[int]
	}{
		{
			name: "int slice samediff with repeated elements",
			a:    []int{1, 2, 3, 4, 5, 6, 6, 6, 7, 8, 9},
			b:    []int{1, 2, 3, 4, 5, 0, 7, 8, 9},
			diff: []Diff[int]{
				{e: 6, t: diffDelete, aIdx: 5, bIdx: 5},
				{e: 6, t: diffDelete, aIdx: 6, bIdx: 6},
				{e: 6, t: diffDelete, aIdx: 7, bIdx: 7},
				{e: 0, t: diffAdd, aIdx: 5, bIdx: 5},
			},
		},
		{
			name: "int slice diff",
			a:    []int{1, 2, 3},
			b:    []int{1, 5, 3},
			diff: []Diff[int]{
				{e: 2, t: diffDelete, aIdx: 1, bIdx: 1},
				{e: 5, t: diffAdd, aIdx: 1, bIdx: 1},
			},
		},
		{
			name: "empty slice diff",
			a:    []int{},
			b:    []int{},
			diff: []Diff[int]{},
		},
	}

	for _, tt := range tests {
		diffs := CompareSequences(tt.a, tt.b, equalsFun)
		assert.True(slices.Equal(tt.diff, diffs), ":%s:diff: want: %v, got: %v", tt.name, tt.diff, diffs)
	}
}

func TestMaxDiff(t *testing.T) {
	assert := assert.New(t)

	a := []rune("abcd")
	b := []rune("dcba")

	diff1 := CompareSequences(a, b, equalsFun)
	assert.Equal(6, len(diff1), "want: 6 diffs, actual: %d", len(diff1))

	diff2 := CompareSequencesEx(a, b, equalsFun, false, 1)
	assert.Equal(2, len(diff2), "want: 2 diffs, actual: %d", len(diff2))
}

func TestDiffPluralSubsequence(t *testing.T) {
	a := []rune("abcaaaaaabd")
	b := []rune("abdaaaaaabc")
	// dividing sequences forcibly
	diffActual := CompareSequencesEx(a, b, equalsFun, false, 1)
	if len(diffActual) != 2 {
		t.Fatalf("diffs length is %d, want 2", len(diffActual))
	}
}

func BenchmarkStringDiffCompose(b *testing.B) {
	s1 := []rune("abc")
	s2 := []rune("abd")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareSequences(s1, s2, equalsFun)
	}
}
