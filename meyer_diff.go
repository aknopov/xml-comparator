package xmlcomparator

import (
	"bytes"
	"fmt"
)

// Implementation of O(NP) Myers' diff algorithm (http://www.xmailserver.org/diff2.pdf)

const (
	// Limit of max differences
	defaultMaxDiffs = 2000000
)

// editType is manipulation type
type editType int

// Type of modification
const (
	diffDelete editType = iota // deleted element
	diffSame                   // same element
	diffAdd                    // added element
)

// `coord` is a coordinate in edit graph
type coord struct {
	x, y int
}

// `graph` is a coordinate in edit graph with attached route
type graph struct {
	x, y, r int
}

// `Diff` is shortest edit script
type Diff[T any] struct {
	e    T
	t    editType
	aIdx int
	bIdx int
}

// `algData` is algorithm data for calculating difference between a and b
type algData[T any] struct {
	a, b         []T
	m, n         int
	diffs        []Diff[T]
	reverse      bool
	paths        []int
	graphs       []graph
	maxDiffs     int
	recordEquals bool
	equals       func(x, y T) bool
}

// CompareSequences compares two sequences of any type and returns a list of differences.
// The comparison is done using the provided `equals` function.
//
// Type Parameters:
//
//	T - The type of the elements in the sequences.
//
// Parameters:
//
//	a - The first sequence to compare.
//	b - The second sequence to compare.
//	equals - A comparison function that checks equality of its arguments.
//
// Returns:
//
//	A slice of Diff[T] representing the differences between the two sequences.
func CompareSequences[T any](a, b []T, equals func(x, y T) bool) []Diff[T] {
	return CompareSequencesEx(a, b, equals, false, defaultMaxDiffs)
}

// CompareSequencesEx compares two sequences of any type and returns a list of differences.
// The comparison is done using the provided `equals` function.
// The maxDiffs parameter specifies the maximum number of differences to analyse. The default values is 2000000.
//
// Type Parameters:
//
//	T - The type of the elements in the sequences.
//
// Parameters:
//   - a - The first sequence to compare.
//   - b - The second sequence to compare.
//   - equals - A comparison function that checks equality of its arguments.
//   - recordEquals - Whether to record equal elements
//   - maxDiffs - The maximum number of edit graphs to analyse.
//
// Returns:
//
//	A slice of Diff[T] representing the differences between the two sequences.
func CompareSequencesEx[T any](a, b []T, equals func(x, y T) bool, recordEquals bool, maxDiffs int) []Diff[T] {
	diff := create(a, b, equals)
	diff.recordEquals = recordEquals
	diff.maxDiffs = maxDiffs

	diff.recordDiffs(diff.compose())

	return diff.Diffs()
}

// SerializeDiffs returns string presentation of supplied differences
func SerializeDiffs[T any](diffs []Diff[T]) string {
	var buf bytes.Buffer
	for i := range diffs {
		switch diffs[i].t {
		case diffDelete:
			fmt.Fprintf(&buf, "-%v[%d<->%d]\n", diffs[i].e, diffs[i].aIdx, diffs[i].bIdx)
		case diffAdd:
			fmt.Fprintf(&buf, "+%v[%d<->%d]\n", diffs[i].e, diffs[i].aIdx, diffs[i].bIdx)
		case diffSame:
			fmt.Fprintf(&buf, "=%v[%d<->%d]\n", diffs[i].e, diffs[i].aIdx, diffs[i].bIdx)
		}
	}
	return buf.String()
}

//-------------------------------------------------------------------------

// Initializes algorithm data
func create[T any](a, b []T, equals func(x, y T) bool) *algData[T] {
	diff := new(algData[T])
	m, n := len(a), len(b)
	reverse := false
	if m >= n {
		a, b = b, a
		m, n = n, m
		reverse = true
	}
	diff.a = a
	diff.b = b
	diff.m = m
	diff.n = n
	diff.reverse = reverse
	diff.maxDiffs = defaultMaxDiffs
	diff.equals = equals
	return diff
}

// Diffs return the list of differences between samples
func (diff *algData[T]) Diffs() []Diff[T] {
	return diff.diffs
}

// Compose diff between samplea
func (diff *algData[T]) compose() []coord {
	fp := make([]int, diff.m+diff.n+3)
	diff.paths = make([]int, diff.m+diff.n+3)
	diff.graphs = make([]graph, 0)

	for i := range fp {
		fp[i] = -1
		diff.paths[i] = -1
	}

	offset := diff.m + 1
	delta := diff.n - diff.m
	for p := 0; ; p++ {
		for k := -p; k <= delta-1; k++ {
			fp[k+offset] = diff.snake(k, fp[k-1+offset]+1, fp[k+1+offset], offset)
		}

		for k := delta + p; k >= delta+1; k-- {
			fp[k+offset] = diff.snake(k, fp[k-1+offset]+1, fp[k+1+offset], offset)
		}

		fp[delta+offset] = diff.snake(delta, fp[delta-1+offset]+1, fp[delta+1+offset], offset)

		if fp[delta+offset] >= diff.n || len(diff.graphs) > diff.maxDiffs {
			break
		}
	}

	r := diff.paths[delta+offset]
	comparePoints := make([]coord, 0)
	for r != -1 {
		comparePoints = append(comparePoints, coord{x: diff.graphs[r].x, y: diff.graphs[r].y})
		r = diff.graphs[r].r
	}

	return comparePoints
}

func (diff *algData[T]) snake(k, p, pp, offset int) int {
	r := 0
	if p > pp {
		r = diff.paths[k-1+offset]
	} else {
		r = diff.paths[k+1+offset]
	}

	y := max(p, pp)
	x := y - k

	for x < diff.m && y < diff.n && diff.equals(diff.a[x], diff.b[y]) {
		x++
		y++
	}

	diff.paths[k+offset] = len(diff.graphs)
	diff.graphs = append(diff.graphs, graph{x: x, y: y, r: r})

	return y
}

//nolint:cyclop // cyclomatic complexity = 13, ignoring for this
func (diff *algData[T]) recordDiffs(comparePoints []coord) {
	x, y := 1, 1
	px, py := 0, 0
	for i := len(comparePoints) - 1; i >= 0; i-- {
		for (px < comparePoints[i].x) || (py < comparePoints[i].y) {
			switch {
			case (comparePoints[i].y - comparePoints[i].x) > (py - px):
				if diff.reverse {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.b[py], t: diffDelete, aIdx: y - 1, bIdx: y - 1})
				} else {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.b[py], t: diffAdd, aIdx: y - 1, bIdx: y - 1})
				}
				y++
				py++
			case (comparePoints[i].y - comparePoints[i].x) < (py - px):
				if diff.reverse {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.a[px], t: diffAdd, aIdx: x - 1, bIdx: x - 1})
				} else {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.a[px], t: diffDelete, aIdx: x - 1, bIdx: x - 1})
				}
				x++
				px++
			default:
				if diff.recordEquals {
					if diff.reverse {
						diff.diffs = append(diff.diffs, Diff[T]{e: diff.b[py], t: diffSame, aIdx: y - 1, bIdx: x - 1})
					} else {
						diff.diffs = append(diff.diffs, Diff[T]{e: diff.a[px], t: diffSame, aIdx: x - 1, bIdx: y - 1})
					}
				}
				x++
				y++
				px++
				py++
			}
		}
	}
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
