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
