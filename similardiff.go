package main

import (
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SimilarDiff struct {
	Cursor  int
	FileA   string
	FileB   string
	Lines   []string
	Pairs   []SimilarDiffPair
	Changes []SimilarDiffChange
	Total   int
}

type SimilarDiffPair struct {
	Left      string
	Right     string
	LeftLine  int
	RightLine int
}

type SimilarDiffChange struct {
	Old string
	New string
}

func NewSimilarDiff() *SimilarDiff {
	return &SimilarDiff{}
}

func (s *SimilarDiff) SetFileA(name string) {
	s.FileA = name
}

func (s *SimilarDiff) SetFileB(name string) {
	s.FileB = name
}

func (s *SimilarDiff) SetChanges(name string) {
	var parts []string

	changes := strings.Split(name, ",")

	for _, change := range changes {
		if len(change) < 3 {
			continue
		}

		if !strings.Contains(change, ":") {
			continue
		}

		parts = strings.Split(change, ":")
		s.Changes = append(s.Changes, SimilarDiffChange{
			Old: parts[0],
			New: parts[1],
		})
	}
}

func (s *SimilarDiff) FindChanges() {
	/* discard errors; exit(1) means there are differences */
	out, _ := exec.Command("/usr/bin/diff", s.FileA, s.FileB).CombinedOutput()
	s.Lines = strings.Split(string(out), "\n")
	s.Total = len(s.Lines)
}

func (s *SimilarDiff) CaptureChanges() {
	for idx := 0; idx < s.Total; idx++ {
		s.Cursor = idx

		if s.CaptureSingleLine() {
			idx = s.Cursor
			continue
		}

		if s.CaptureMultipleLine() {
			idx = s.Cursor
			continue
		}
	}
}

// CaptureSingleLine processes single-line diff.
// 1c1 | changed line (file A, file B)
// < A | content in file A
// --- | change separator
// > B | content in file B
func (s *SimilarDiff) CaptureSingleLine() bool {
	heads := regexp.MustCompile(`^([0-9]+)c([0-9]+)$`)

	if heads.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := heads.FindStringSubmatch(s.Lines[s.Cursor])

	s.Pairs = append(s.Pairs, SimilarDiffPair{
		Left:      s.Lines[s.Cursor+1][2:],
		Right:     s.Lines[s.Cursor+3][2:],
		LeftLine:  s.ConvertAtoi(m[1]),
		RightLine: s.ConvertAtoi(m[2]),
	})

	s.Cursor += 3

	return true
}

// CaptureMultipleLine processes multiple-line diff.
// 1,3c7,9 | changed lines (file A, file B)
// < A     | content in file A, line 1
// < B     | content in file A, line 2
// < C     | content in file A, line 3
// ---     | change separator
// > X     | content in file B, line 7
// > Y     | content in file B, line 8
// > Z     | content in file B, line 9
func (s *SimilarDiff) CaptureMultipleLine() bool {
	headm := regexp.MustCompile(`^([0-9]+),([0-9]+)c([0-9]+),([0-9]+)$`)

	if headm.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	var howmany int

	m := headm.FindStringSubmatch(s.Lines[s.Cursor])
	numLeftA := s.ConvertAtoi(m[1])  /* 1,3c7,9 -> 1 */
	numLeftB := s.ConvertAtoi(m[2])  /* 1,3c7,9 -> 3 */
	numRightA := s.ConvertAtoi(m[3]) /* 1,3c7,9 -> 7 */
	numRightB := s.ConvertAtoi(m[4]) /* 1,3c7,9 -> 9 */

	howmany = (numLeftB - numLeftA) + 1 /* inclusive */
	subgroups := make([]SimilarDiffPair, howmany)

	s.Cursor++ /* skip multiple-line diff header */

	for i := 0; i < howmany; i++ {
		subgroups[i].Left = s.Lines[s.Cursor+i][2:]
		subgroups[i].LeftLine = numLeftA + i
	}

	s.Cursor += howmany + 1 /* left matches + separator */

	howmany = (numRightB - numRightA) + 1 /* inclusive */

	for i := 0; i < howmany; i++ {
		subgroups[i].Right = s.Lines[s.Cursor+i][2:]
		subgroups[i].RightLine = numRightA + i
	}

	for _, group := range subgroups {
		s.Pairs = append(s.Pairs, group)
	}

	return true
}

func (s *SimilarDiff) DiscardSimilarities() {
	var temp string
	var group SimilarDiffPair

	totalPairs := len(s.Pairs)
	notDiscarded := make([]SimilarDiffPair, 0)

	for i := 0; i < totalPairs; i++ {
		group = s.Pairs[i]
		temp = group.Left

		for _, change := range s.Changes {
			temp = strings.Replace(temp, change.Old, change.New, -1)
		}

		/* lines are similar */
		if temp == group.Right {
			continue
		}

		notDiscarded = append(notDiscarded, group)
	}

	s.Pairs = notDiscarded
}

func (s *SimilarDiff) PrettyPrint() {
	s.FindChanges() /* read and run diff */

	s.CaptureChanges() /* find and process */

	s.DiscardSimilarities()

	/* there are no changes */
	if len(s.Pairs) <= 0 {
		return
	}

	fmt.Printf("--- %s\n", s.FileA)
	fmt.Printf("+++ %s\n", s.FileB)

	for _, group := range s.Pairs {
		fmt.Printf("@@ -%d +%d @@\n", group.LeftLine, group.RightLine)
		fmt.Printf("-%s\n", group.Left)
		fmt.Printf("+%s\n", group.Right)
	}
}
