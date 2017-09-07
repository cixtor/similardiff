package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type SimilarDiff struct {
	Cursor   int
	FileA    string
	FileB    string
	Lines    []string
	Pairs    []SimilarDiffPair
	Changes  []SimilarDiffChange
	Colorize bool
	Total    int
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

func (s *SimilarDiff) SetColorize(value string) {
	s.Colorize = (value == "true")
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

		if s.CaptureChangedLinesOne() {
			idx = s.Cursor
			continue
		}

		if s.CaptureChangedLinesMany() {
			idx = s.Cursor
			continue
		}
	}
}

// CaptureChangedLinesOne detects and captures single-line diff.
// 1c1 | changed line (file A, file B)
// < A | content in file A
// --- | change separator
// > B | content in file B
func (s *SimilarDiff) CaptureChangedLinesOne() bool {
	header := regexp.MustCompile(`^([0-9]+)c([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])

	s.Pairs = append(s.Pairs, SimilarDiffPair{
		Left:      s.Lines[s.Cursor+1][2:],
		Right:     s.Lines[s.Cursor+3][2:],
		LeftLine:  s.ConvertAtoi(m[1]),
		RightLine: s.ConvertAtoi(m[2]),
	})

	s.Cursor += 3

	return true
}

// CaptureChangedLinesMany detects and captures multiple-line diff.
// 1,3c7,9 | changed lines (file A, file B)
// < A     | content in file A, line 1
// < B     | content in file A, line 2
// < C     | content in file A, line 3
// ---     | change separator
// > X     | content in file B, line 7
// > Y     | content in file B, line 8
// > Z     | content in file B, line 9
func (s *SimilarDiff) CaptureChangedLinesMany() bool {
	header := regexp.MustCompile(`^([0-9]+),([0-9]+)c([0-9]+),([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	var howmany int

	m := header.FindStringSubmatch(s.Lines[s.Cursor])
	numLeftA := s.ConvertAtoi(m[1])  /* 1,3c7,9 -> 1 */
	numLeftB := s.ConvertAtoi(m[2])  /* 1,3c7,9 -> 3 */
	numRightA := s.ConvertAtoi(m[3]) /* 1,3c7,9 -> 7 */
	numRightB := s.ConvertAtoi(m[4]) /* 1,3c7,9 -> 9 */

	howmany = (numLeftB - numLeftA) + 1 /* inclusive */
	subgroups := make([]SimilarDiffPair, howmany)

	s.Cursor++ /* skip diff header */

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

	s.PrintRed("--- %s", s.FileA)
	s.PrintGreen("+++ %s", s.FileB)

	for _, group := range s.Pairs {
		if group.LeftLine > 0 {
			s.PrintRed("%d\t-%s", group.LeftLine, group.Left)
		}

		s.PrintGreen("%d\t+%s", group.RightLine, group.Right)
	}
}

func (s *SimilarDiff) ConvertAtoi(number string) int {
	num, err := strconv.Atoi(number)

	if err != nil {
		return 0
	}

	return num
}

func (s *SimilarDiff) PrintRed(format string, text ...interface{}) {
	if s.Colorize {
		fmt.Print("\033[0;31m")
	}

	fmt.Printf(format, text...)

	if s.Colorize {
		fmt.Print("\033[0m")
	}

	fmt.Print("\n")
}

func (s *SimilarDiff) PrintGreen(format string, text ...interface{}) {
	if s.Colorize {
		fmt.Print("\033[0;32m")
	}

	fmt.Printf(format, text...)

	if s.Colorize {
		fmt.Print("\033[0m")
	}

	fmt.Print("\n")
}

func main() {
	flag.Parse()

	s := NewSimilarDiff()

	s.SetFileA(flag.Arg(0))
	s.SetFileB(flag.Arg(1))
	s.SetChanges(flag.Arg(2))
	s.SetColorize(os.Getenv("SIMILARDIFF_COLOR"))

	s.PrettyPrint()
}
