package main

import (
	"testing"
)

func CheckTestData(t *testing.T, s *SimilarDiff, total int, expected []SimilarDiffPair) {
	if len(s.Pairs) != total {
		t.Logf("-%d", total)
		t.Logf("+%d\n", len(s.Pairs))
		t.Fatal("Number of detected pairs is incorrect")
	}

	if s.Pairs[0] != expected[0] {
		t.Logf("-%#v\n", expected[0])
		t.Logf("+%#v\n", s.Pairs[0])
		t.Fatal("Failure detecting changes in single lines: Index[0]")
	}

	for i := 0; i < total; i++ {
		if s.Pairs[i] == expected[i] {
			continue
		}

		t.Logf("-%#v\n", expected[i])
		t.Logf("+%#v\n", s.Pairs[i])
		t.Fatalf("Failure detecting and processing diff: Index[%d]", i)
	}
}

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

	CheckTestData(t, s, 2, expected)
}

func TestCaptureChangedLinesMany(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"1,3c7,9",
		"< A     | content in file A, line 1",
		"< B     | content in file A, line 2",
		"< C     | content in file A, line 3",
		"---     | change separator",
		"> X     | content in file B, line 7",
		"> Y     | content in file B, line 8",
		"> Z     | content in file B, line 9",
		"10,12c20,22",
		"< A     | content in file A, line 10",
		"< B     | content in file A, line 11",
		"< C     | content in file A, line 12",
		"---     | change separator",
		"> X     | content in file B, line 20",
		"> Y     | content in file B, line 21",
		"> Z     | content in file B, line 22",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 6)

	expected[0] = SimilarDiffPair{
		Left:      "A     | content in file A, line 1",
		Right:     "X     | content in file B, line 7",
		LeftLine:  1,
		RightLine: 7,
	}
	expected[1] = SimilarDiffPair{
		Left:      "B     | content in file A, line 2",
		Right:     "Y     | content in file B, line 8",
		LeftLine:  2,
		RightLine: 8,
	}
	expected[2] = SimilarDiffPair{
		Left:      "C     | content in file A, line 3",
		Right:     "Z     | content in file B, line 9",
		LeftLine:  3,
		RightLine: 9,
	}
	expected[3] = SimilarDiffPair{
		Left:      "A     | content in file A, line 10",
		Right:     "X     | content in file B, line 20",
		LeftLine:  10,
		RightLine: 20,
	}
	expected[4] = SimilarDiffPair{
		Left:      "B     | content in file A, line 11",
		Right:     "Y     | content in file B, line 21",
		LeftLine:  11,
		RightLine: 21,
	}
	expected[5] = SimilarDiffPair{
		Left:      "C     | content in file A, line 12",
		Right:     "Z     | content in file B, line 22",
		LeftLine:  12,
		RightLine: 22,
	}

	CheckTestData(t, s, 6, expected)
}

func TestCaptureDeletedLines(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"10,13d5",
		"< W     | content in file A, line 10",
		"< X     | content in file A, line 11",
		"< Y     | content in file A, line 12",
		"< Z     | content in file A, line 13",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 4)

	expected[0] = SimilarDiffPair{
		Left:      "W     | content in file A, line 10",
		Right:     "",
		LeftLine:  10,
		RightLine: 0,
	}
	expected[1] = SimilarDiffPair{
		Left:      "X     | content in file A, line 11",
		Right:     "",
		LeftLine:  11,
		RightLine: 0,
	}
	expected[2] = SimilarDiffPair{
		Left:      "Y     | content in file A, line 12",
		Right:     "",
		LeftLine:  12,
		RightLine: 0,
	}
	expected[3] = SimilarDiffPair{
		Left:      "Z     | content in file A, line 13",
		Right:     "",
		LeftLine:  13,
		RightLine: 0,
	}

	CheckTestData(t, s, 4, expected)
}

func TestCaptureAddedLines(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"5a10,13",
		"> W     | content in file B, line 10",
		"> X     | content in file B, line 11",
		"> Y     | content in file B, line 12",
		"> Z     | content in file B, line 13",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 4)

	expected[0] = SimilarDiffPair{
		Left:      "",
		Right:     "W     | content in file B, line 10",
		LeftLine:  0,
		RightLine: 10,
	}
	expected[1] = SimilarDiffPair{
		Left:      "",
		Right:     "X     | content in file B, line 11",
		LeftLine:  0,
		RightLine: 11,
	}
	expected[2] = SimilarDiffPair{
		Left:      "",
		Right:     "Y     | content in file B, line 12",
		LeftLine:  0,
		RightLine: 12,
	}
	expected[3] = SimilarDiffPair{
		Left:      "",
		Right:     "Z     | content in file B, line 13",
		LeftLine:  0,
		RightLine: 13,
	}

	CheckTestData(t, s, 4, expected)
}
