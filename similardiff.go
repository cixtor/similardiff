package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const changed rune = 'c'
const added rune = 'a'
const deleted rune = 'd'

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
	Group     rune
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
	folder, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	/* configuration file does not exists; skip changes */
	if _, err := os.Stat(folder + "/" + name); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(folder + "/" + name)

	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

	defer file.Close()

	var line string
	var parts []string

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line = scanner.Text()
		line = strings.TrimSpace(line)

		/* expect x=y */
		if len(line) < 3 {
			continue
		}

		/* skip comments */
		if line[0] == '#' {
			continue
		}

		parts = strings.Split(scanner.Text(), "=")

		s.Changes = append(s.Changes, SimilarDiffChange{
			Old: parts[0],
			New: parts[1],
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
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

		if s.CaptureChangedLinesManyBothSides() {
			idx = s.Cursor
			continue
		}

		if s.CaptureChangedLinesManyRightSide() {
			idx = s.Cursor
			continue
		}

		if s.CaptureChangedLinesManyLeftSide() {
			idx = s.Cursor
			continue
		}

		if s.CaptureDeletedLinesOne() {
			idx = s.Cursor
			continue
		}

		if s.CaptureDeletedLinesMany() {
			idx = s.Cursor
			continue
		}

		if s.CaptureAddedLinesOne() {
			idx = s.Cursor
			continue
		}

		if s.CaptureAddedLinesMany() {
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
		Group:     changed,
		Left:      s.Lines[s.Cursor+1][2:],
		Right:     s.Lines[s.Cursor+3][2:],
		LeftLine:  s.ConvertAtoi(m[1]),
		RightLine: s.ConvertAtoi(m[2]),
	})

	s.Cursor += 3

	return true
}

// CaptureChangedLinesManyBothSides detects and captures multiple-line diff.
//
// Differences where the left and right side are balanced.
//
// 1,3c7,9 | changed lines (file A, file B)
// < A     | content in file A, line 1
// < B     | content in file A, line 2
// < C     | content in file A, line 3
// ---     | change separator
// > X     | content in file B, line 7
// > Y     | content in file B, line 8
// > Z     | content in file B, line 9
//
// Differences where the left and right side are unbalanced (added lines).
//
// 1,3c10,14 | changed lines (file A, file B)
// < A       | content in file A, line 1
// < B       | content in file A, line 2
// < C       | content in file A, line 3
// ---       | change separator
// > V       | content in file B, line 10
// > W       | content in file B, line 11
// > X       | content in file B, line 12
// > Y       | content in file B, line 13
// > Z       | content in file B, line 14
//
// Differences where the left and right side are unbalanced (deleted lines).
//
// 1,5c10,12 | changed lines (file A, file B)
// < A       | content in file A, line 1
// < B       | content in file A, line 2
// < C       | content in file A, line 3
// < D       | content in file A, line 4
// < E       | content in file A, line 5
// ---       | change separator
// > X       | content in file B, line 10
// > Y       | content in file B, line 11
// > Z       | content in file B, line 12
func (s *SimilarDiff) CaptureChangedLinesManyBothSides() bool {
	header := regexp.MustCompile(`^([0-9]+),([0-9]+)c([0-9]+),([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])
	numLeftA := s.ConvertAtoi(m[1])  /* 1,3c7,9 -> 1 */
	numLeftB := s.ConvertAtoi(m[2])  /* 1,3c7,9 -> 3 */
	numRightA := s.ConvertAtoi(m[3]) /* 1,3c7,9 -> 7 */
	numRightB := s.ConvertAtoi(m[4]) /* 1,3c7,9 -> 9 */

	numItemsLeft := (numLeftB - numLeftA) + 1    /* inclusive */
	numItemsRight := (numRightB - numRightA) + 1 /* inclusive */

	var howmany int
	var padding int
	var hasAddedLines bool
	var hasDeletedLines bool

	if numItemsLeft == numItemsRight {
		/* balanced; pairing sides */
		howmany = numItemsLeft
		padding = numItemsLeft
	} else if numItemsLeft > numItemsRight {
		/* unbalanced; deleted lines */
		howmany = numItemsRight
		padding = numItemsLeft
		hasDeletedLines = true
	} else if numItemsLeft < numItemsRight {
		/* unbalanced; added lines */
		howmany = numItemsLeft
		padding = numItemsLeft
		hasAddedLines = true
	}

	/* capture pairing differences */
	for i := 0; i < howmany; i++ {
		s.Cursor++ /* move cursor ahead */
		s.Pairs = append(s.Pairs, SimilarDiffPair{
			Group:     changed,
			Left:      s.Lines[s.Cursor][2:],
			Right:     s.Lines[s.Cursor+padding+1][2:],
			LeftLine:  numLeftA + i,
			RightLine: numRightA + i,
		})
	}

	if hasAddedLines {
		s.Cursor += howmany + 1 /* move ahead */
		/* how many items remain in the stack */
		remaining := numItemsRight - numItemsLeft
		/* unbalanced differences; added lines */
		for i := 0; i < remaining; i++ {
			s.Cursor++ /* move cursor ahead */
			s.Pairs = append(s.Pairs, SimilarDiffPair{
				Group:     added,
				Right:     s.Lines[s.Cursor][2:],
				RightLine: numRightA + howmany + i,
			})
		}
	}

	if hasDeletedLines {
		/* how many items remain in the stack */
		remaining := numItemsLeft - numItemsRight
		/* unbalanced differences; deleted lines */
		for i := 0; i < remaining; i++ {
			s.Cursor++ /* move cursor ahead */
			s.Pairs = append(s.Pairs, SimilarDiffPair{
				Group:    deleted,
				Left:     s.Lines[s.Cursor][2:],
				LeftLine: numLeftA + howmany + i,
			})
		}
	}

	return true
}

// CaptureChangedLinesManyRightSide detects and captures multiple-line diff.
// 296c300,303 | changed lines (file A, file B)
// < A         | content in file A, line 296
// ---         | change separator
// > B         | content in file B, line 300
// > X         | content in file B, line 301
// > Y         | content in file B, line 302
// > Z         | content in file B, line 303
func (s *SimilarDiff) CaptureChangedLinesManyRightSide() bool {
	header := regexp.MustCompile(`^([0-9]+)c([0-9]+),([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])
	numLeftA := s.ConvertAtoi(m[1])  /* 296c300,303 -> 296 */
	numRightA := s.ConvertAtoi(m[2]) /* 296c300,303 -> 300 */
	numRightB := s.ConvertAtoi(m[3]) /* 296c300,303 -> 303 */

	/* capture pairing differences */
	s.Pairs = append(s.Pairs, SimilarDiffPair{
		Group:     changed,
		Left:      s.Lines[s.Cursor+1][2:],
		Right:     s.Lines[s.Cursor+3][2:],
		LeftLine:  numLeftA,
		RightLine: numRightA,
	})
	s.Cursor += 3

	/* capture orphan differences; added lines */
	howmany := (numRightB - numRightA) /* exclusive */
	for i := 0; i < howmany; i++ {
		s.Cursor++ /* move cursor ahead */
		s.Pairs = append(s.Pairs, SimilarDiffPair{
			Group:     added,
			Right:     s.Lines[s.Cursor][2:],
			RightLine: numRightA + i + 1,
		})
	}

	return true
}

// CaptureChangedLinesManyLeftSide detects and captures multiple-line diff.
// 126,128c130 | changed lines (file A, file B)
// < A         | content in file A, line 126
// < B         | content in file A, line 127
// < C         | content in file A, line 128
// ---         | change separator
// > X         | content in file B, line 130
func (s *SimilarDiff) CaptureChangedLinesManyLeftSide() bool {
	header := regexp.MustCompile(`^([0-9]+),([0-9]+)c([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])
	numLeftA := s.ConvertAtoi(m[1])  /* 126,128c130 -> 126 */
	numLeftB := s.ConvertAtoi(m[2])  /* 126,128c130 -> 128 */
	numRightA := s.ConvertAtoi(m[3]) /* 126,128c130 -> 130 */

	/* capture pairing differences */
	s.Cursor++ /* move cursor ahead */
	padding := (numLeftB - numLeftA) + 2
	s.Pairs = append(s.Pairs, SimilarDiffPair{
		Group:     changed,
		Left:      s.Lines[s.Cursor][2:],
		Right:     s.Lines[s.Cursor+padding][2:],
		LeftLine:  numLeftA,
		RightLine: numRightA,
	})

	/* capture orphan differences; deleted lines */
	remaining := (numLeftB - numLeftA) /* exclusive */
	for i := 0; i < remaining; i++ {
		s.Cursor++ /* move cursor ahead */
		s.Pairs = append(s.Pairs, SimilarDiffPair{
			Group:    deleted,
			Left:     s.Lines[s.Cursor][2:],
			LeftLine: numLeftA + i + 1,
		})
	}

	return true
}

// CaptureDeletedLinesOne detects and captures deleted lines.
// 13d12 | changed lines (file A, file B)
// < A   | content in file A, line 13
// 43d22 | changed lines (file A, file B)
// < B   | content in file A, line 43
func (s *SimilarDiff) CaptureDeletedLinesOne() bool {
	header := regexp.MustCompile(`^([0-9]+)d([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])

	s.Cursor++ /* move cursor ahead */

	s.Pairs = append(s.Pairs, SimilarDiffPair{
		Group:    deleted,
		Left:     s.Lines[s.Cursor][2:],
		LeftLine: s.ConvertAtoi(m[1]),
	})

	return true
}

// CaptureDeletedLinesMany detects and captures deleted lines.
// 10,13d5 | changed lines (file A, file B)
// < W     | content in file A, line 10
// < X     | content in file A, line 11
// < Y     | content in file A, line 12
// < Z     | content in file A, line 13
func (s *SimilarDiff) CaptureDeletedLinesMany() bool {
	header := regexp.MustCompile(`^([0-9]+),([0-9]+)d([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])
	numLeftA := s.ConvertAtoi(m[1]) /* 10,13d5 -> 10 */
	numLeftB := s.ConvertAtoi(m[2]) /* 10,13d5 -> 13 */

	for i := numLeftA; i <= numLeftB; i++ {
		s.Cursor++ /* move cursor ahead */
		s.Pairs = append(s.Pairs, SimilarDiffPair{
			Group:    deleted,
			Left:     s.Lines[s.Cursor][2:],
			LeftLine: i, /* real line number */
		})
	}

	return true
}

// CaptureAddedLinesOne detects and captures added lines.
// 5a10 | changed lines (file A, file B)
// > A  | content in file B, line 10
// 5a20 | changed lines (file A, file B)
// > B  | content in file B, line 20
func (s *SimilarDiff) CaptureAddedLinesOne() bool {
	header := regexp.MustCompile(`^([0-9]+)a([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])

	s.Cursor++ /* move cursor ahead */

	s.Pairs = append(s.Pairs, SimilarDiffPair{
		Group:     added,
		Right:     s.Lines[s.Cursor][2:],
		RightLine: s.ConvertAtoi(m[2]),
	})

	return true
}

// CaptureAddedLinesMany detects and captures added lines.
// 5a10,13 | changed lines (file A, file B)
// > W     | content in file B, line 10
// > X     | content in file B, line 11
// > Y     | content in file B, line 12
// > Z     | content in file B, line 13
func (s *SimilarDiff) CaptureAddedLinesMany() bool {
	header := regexp.MustCompile(`^([0-9]+)a([0-9]+),([0-9]+)$`)

	if header.FindString(s.Lines[s.Cursor]) == "" {
		return false
	}

	m := header.FindStringSubmatch(s.Lines[s.Cursor])
	numRightA := s.ConvertAtoi(m[2]) /* 5a10,13 -> 10 */
	numRightB := s.ConvertAtoi(m[3]) /* 5a10,13 -> 13 */

	for i := numRightA; i <= numRightB; i++ {
		s.Cursor++ /* move cursor ahead */
		s.Pairs = append(s.Pairs, SimilarDiffPair{
			Group:     added,
			Right:     s.Lines[s.Cursor][2:],
			RightLine: i, /* real line number */
		})
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

		/* cannot compare lines that were added or deleted */
		if group.Group == added || group.Group == deleted {
			notDiscarded = append(notDiscarded, group)
			continue
		}

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

		if group.RightLine > 0 {
			s.PrintGreen("%d\t+%s", group.RightLine, group.Right)
		}
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
	flag.Usage = func() {
		fmt.Println("Similar Diff")
		fmt.Println("https://cixtor.com/")
		fmt.Println("https://github.com/cixtor/similardiff")
		fmt.Println("https://en.wikipedia.org/wiki/Edit_distance")
		fmt.Println("https://en.wikipedia.org/wiki/File_comparison")
		fmt.Println("https://en.wikipedia.org/wiki/Levenshtein_distance")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  similardiff [FILE_A] [FILE_B]")
		fmt.Println()
		fmt.Println("Settings:")
		fmt.Println("  export SIMILARDIFF_COLOR=true")
		fmt.Println("  echo \"#file_a:file_b\" 1>> similardiff.ini")
		fmt.Println("  echo \"import:include\" 1>> similardiff.ini")
		fmt.Println("  echo \"package:module\" 1>> similardiff.ini")
		fmt.Println("  similardiff file_a.txt file_b.txt")
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	s := NewSimilarDiff()

	s.SetFileA(flag.Arg(0))
	s.SetFileB(flag.Arg(1))
	s.SetChanges("similardiff.ini")
	s.SetColorize(os.Getenv("SIMILARDIFF_COLOR"))

	s.PrettyPrint()
}
