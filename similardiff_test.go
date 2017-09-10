package main

import (
	"testing"
)

func TestCaptureChangedLinesOne(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"1c1",
		"< A | content in file A",
		"--- | change separator",
		"> B | content in file B",
		"10c10",
		"< X | content in file X",
		"--- | change separator",
		"> Y | content in file Y",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 2)

	expected[0] = SimilarDiffPair{
		Left:      "A | content in file A",
		Right:     "B | content in file B",
		LeftLine:  1,
		RightLine: 1,
	}
	expected[1] = SimilarDiffPair{
		Left:      "X | content in file X",
		Right:     "Y | content in file Y",
		LeftLine:  10,
		RightLine: 10,
	}

	if len(s.Pairs) != 2 {
		t.Fatal("Number of detected pairs is incorrect")
	}

	if s.Pairs[0] != expected[0] {
		t.Logf("-%#v\n", expected[0])
		t.Logf("+%#v\n", s.Pairs[0])
		t.Fatal("Failure detecting changes in single lines: Index[0]")
	}

	if s.Pairs[1] != expected[1] {
		t.Logf("-%#v\n", expected[1])
		t.Logf("+%#v\n", s.Pairs[1])
		t.Fatal("Failure detecting changes in single lines: Index[1]")
	}
}
