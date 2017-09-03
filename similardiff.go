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
