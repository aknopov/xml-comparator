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

// DiffType is manipulation type
type DiffType int

// Type of modification
const (
	DiffDelete DiffType = iota // deleted element
	DiffSame                   // same element
	DiffAdd                    // added element
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
	t    DiffType
	aIdx int
	bIdx int
}

// `algData` is algorithm data for calculating difference between a and b
type algData[T any] struct {
	a, b           []T
	m, n           int
	ox, oy         int
	diffs          []Diff[T]
	reverse        bool
	path           []int
	pointWithRoute []graph
	contextSize    int
	maxDiffs       int
	recordEquals   bool
	equals         func(x, y T) bool
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
	return CompareSequencesEx(a, b, equals, defaultMaxDiffs)
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
//   - a The first sequence to compare.
//   - b The second sequence to compare.
//   - equals - A comparison function that checks equality of its arguments.
//   - maxDiffs: The maximum number of differences to analyse.
//
// Returns:
//
//	A slice of Diff[T] representing the differences between the two sequences.
func CompareSequencesEx[T any](a, b []T, equals func(x, y T) bool, maxDiffs int) []Diff[T] {
	diff := create(a, b, equals)
	diff.recordEquals = false
	diff.maxDiffs = maxDiffs
	diff.doCompare()
	return diff.Diffs()
}

// SerializeDiffs returns string presentation of supplied differences
func SerializeDiffs[T any](diffs []Diff[T]) string {
	var buf bytes.Buffer
	for i := range diffs {
		switch diffs[i].t {
		case DiffDelete:
			fmt.Fprintf(&buf, "-%v[%d<->%d]\n", diffs[i].e, diffs[i].aIdx, diffs[i].bIdx)
		case DiffAdd:
			fmt.Fprintf(&buf, "+%v[%d<->%d]\n", diffs[i].e, diffs[i].aIdx, diffs[i].bIdx)
		case DiffSame:
			fmt.Fprintf(&buf, "=%v[%d<->%d]\n", diffs[i].e, diffs[i].aIdx, diffs[i].bIdx)
		}
	}
	return buf.String()
}

//-------------------------------------------------------------------------

// create is initializer of Diff
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
	diff.contextSize = 3
	diff.maxDiffs = defaultMaxDiffs
	diff.equals = equals
	return diff
}

// Diffs return llist of differences between samplea
func (diff *algData[T]) Diffs() []Diff[T] {
	return diff.diffs
}

// Compare slices till reaching the end
func (diff *algData[T]) doCompare() {
	done := false
	for !done {
		done = diff.recordDiffs(diff.compose())
	}
}

// Compose diff between samplea
func (diff *algData[T]) compose() []coord {
	fp := make([]int, diff.m+diff.n+3)
	diff.path = make([]int, diff.m+diff.n+3)
	diff.pointWithRoute = make([]graph, 0)

	for i := range fp {
		fp[i] = -1
		diff.path[i] = -1
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

		if fp[delta+offset] >= diff.n || len(diff.pointWithRoute) > diff.maxDiffs {
			break
		}
	}

	r := diff.path[delta+offset]
	comparePoints := make([]coord, 0)
	for r != -1 {
		comparePoints = append(comparePoints, coord{x: diff.pointWithRoute[r].x, y: diff.pointWithRoute[r].y})
		r = diff.pointWithRoute[r].r
	}

	return comparePoints
}

func (diff *algData[T]) snake(k, p, pp, offset int) int {
	r := 0
	if p > pp {
		r = diff.path[k-1+offset]
	} else {
		r = diff.path[k+1+offset]
	}

	y := max(p, pp)
	x := y - k

	for x < diff.m && y < diff.n && diff.equals(diff.a[x], diff.b[y]) {
		x++
		y++
	}

	diff.path[k+offset] = len(diff.pointWithRoute)
	diff.pointWithRoute = append(diff.pointWithRoute, graph{x: x, y: y, r: r})

	return y
}

//nolint:cyclop // cyclomatic complexity = 13, ignoring for this
func (diff *algData[T]) recordDiffs(comparerPoints []coord) bool {
	x, y := 1, 1
	px, py := 0, 0
	for i := len(comparerPoints) - 1; i >= 0; i-- {
		for (px < comparerPoints[i].x) || (py < comparerPoints[i].y) {
			switch {
			case (comparerPoints[i].y - comparerPoints[i].x) > (py - px):
				if diff.reverse {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.b[py], t: DiffDelete, aIdx: y + diff.oy - 1, bIdx: -1})
				} else {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.b[py], t: DiffAdd, aIdx: -1, bIdx: y + diff.oy - 1})
				}
				y++
				py++
			case (comparerPoints[i].y - comparerPoints[i].x) < (py - px):
				if diff.reverse {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.a[px], t: DiffAdd, aIdx: -1, bIdx: x + diff.ox - 1})
				} else {
					diff.diffs = append(diff.diffs, Diff[T]{e: diff.a[px], t: DiffDelete, aIdx: x + diff.ox - 1, bIdx: -1})
				}
				x++
				px++
			default:
				if diff.recordEquals {
					if diff.reverse {
						diff.diffs = append(diff.diffs, Diff[T]{e: diff.b[py], t: DiffSame, aIdx: y + diff.oy - 1, bIdx: x + diff.ox - 1})
					} else {
						diff.diffs = append(diff.diffs, Diff[T]{e: diff.a[px], t: DiffSame, aIdx: x + diff.ox - 1, bIdx: y + diff.oy - 1})
					}
				}
				x++
				y++
				px++
				py++
			}
		}
	}

	if x <= diff.m && y <= diff.n {
		diff.a = diff.a[x-1:]
		diff.b = diff.b[y-1:]
		diff.m = len(diff.a)
		diff.n = len(diff.b)
		diff.ox = x - 1
		diff.oy = y - 1
		return false
	}

	// all recording succeeded
	return true
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}
