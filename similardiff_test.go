package main

import (
	"testing"
)

func CheckTestData(t *testing.T, s *SimilarDiff, total int, expected []SimilarDiffPair) {
	if len(s.Pairs) != total {
		t.Logf("-%d ~> %#v", total, expected)
		t.Logf("+%d ~> %#v", len(s.Pairs), s.Pairs)
		t.Fatal("Number of detected pairs is incorrect")
	}

	for i := 0; i < total; i++ {
		if s.Pairs[i] == expected[i] {
			continue
		}

		t.Logf("-%#v", expected[i])
		t.Logf("+%#v", s.Pairs[i])
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
		Group:     'c',
		Left:      "A | content in file A",
		Right:     "B | content in file B",
		LeftLine:  1,
		RightLine: 1,
	}
	expected[1] = SimilarDiffPair{
		Group:     'c',
		Left:      "X | content in file X",
		Right:     "Y | content in file Y",
		LeftLine:  10,
		RightLine: 10,
	}

	CheckTestData(t, s, 2, expected)
}

func TestCaptureChangedLinesManyBothSides(t *testing.T) {
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
		Group:     'c',
		Left:      "A     | content in file A, line 1",
		Right:     "X     | content in file B, line 7",
		LeftLine:  1,
		RightLine: 7,
	}
	expected[1] = SimilarDiffPair{
		Group:     'c',
		Left:      "B     | content in file A, line 2",
		Right:     "Y     | content in file B, line 8",
		LeftLine:  2,
		RightLine: 8,
	}
	expected[2] = SimilarDiffPair{
		Group:     'c',
		Left:      "C     | content in file A, line 3",
		Right:     "Z     | content in file B, line 9",
		LeftLine:  3,
		RightLine: 9,
	}
	expected[3] = SimilarDiffPair{
		Group:     'c',
		Left:      "A     | content in file A, line 10",
		Right:     "X     | content in file B, line 20",
		LeftLine:  10,
		RightLine: 20,
	}
	expected[4] = SimilarDiffPair{
		Group:     'c',
		Left:      "B     | content in file A, line 11",
		Right:     "Y     | content in file B, line 21",
		LeftLine:  11,
		RightLine: 21,
	}
	expected[5] = SimilarDiffPair{
		Group:     'c',
		Left:      "C     | content in file A, line 12",
		Right:     "Z     | content in file B, line 22",
		LeftLine:  12,
		RightLine: 22,
	}

	CheckTestData(t, s, 6, expected)
}

func TestCaptureChangedLinesManyBothSidesAdded(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"1,3c10,14",
		"< A | content in file A, line 1",
		"< B | content in file A, line 2",
		"< C | content in file A, line 3",
		"---",
		"> V | content in file B, line 10",
		"> W | content in file B, line 11",
		"> X | content in file B, line 12",
		"> Y | content in file B, line 13",
		"> Z | content in file B, line 14",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 5)

	expected[0] = SimilarDiffPair{
		Group:     'c',
		Left:      "A | content in file A, line 1",
		Right:     "V | content in file B, line 10",
		LeftLine:  1,
		RightLine: 10,
	}
	expected[1] = SimilarDiffPair{
		Group:     'c',
		Left:      "B | content in file A, line 2",
		Right:     "W | content in file B, line 11",
		LeftLine:  2,
		RightLine: 11,
	}
	expected[2] = SimilarDiffPair{
		Group:     'c',
		Left:      "C | content in file A, line 3",
		Right:     "X | content in file B, line 12",
		LeftLine:  3,
		RightLine: 12,
	}
	expected[3] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "Y | content in file B, line 13",
		LeftLine:  0,
		RightLine: 13,
	}
	expected[4] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "Z | content in file B, line 14",
		LeftLine:  0,
		RightLine: 14,
	}

	CheckTestData(t, s, 5, expected)
}

func TestCaptureChangedLinesManyBothSidesDeleted(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"1,5c10,12",
		"< A | content in file A, line 1",
		"< B | content in file A, line 2",
		"< C | content in file A, line 3",
		"< D | content in file A, line 4",
		"< E | content in file A, line 5",
		"---",
		"> X | content in file B, line 10",
		"> Y | content in file B, line 11",
		"> Z | content in file B, line 12",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 5)

	expected[0] = SimilarDiffPair{
		Group:     'c',
		Left:      "A | content in file A, line 1",
		Right:     "X | content in file B, line 10",
		LeftLine:  1,
		RightLine: 10,
	}
	expected[1] = SimilarDiffPair{
		Group:     'c',
		Left:      "B | content in file A, line 2",
		Right:     "Y | content in file B, line 11",
		LeftLine:  2,
		RightLine: 11,
	}
	expected[2] = SimilarDiffPair{
		Group:     'c',
		Left:      "C | content in file A, line 3",
		Right:     "Z | content in file B, line 12",
		LeftLine:  3,
		RightLine: 12,
	}
	expected[3] = SimilarDiffPair{
		Group:     'd',
		Left:      "D | content in file A, line 4",
		Right:     "",
		LeftLine:  4,
		RightLine: 0,
	}
	expected[4] = SimilarDiffPair{
		Group:     'd',
		Left:      "E | content in file A, line 5",
		Right:     "",
		LeftLine:  5,
		RightLine: 0,
	}

	CheckTestData(t, s, 5, expected)
}

func TestCaptureChangedLinesManyBothSidesDeletedSmall(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"126,128c130",
		"< A | content in file A, line 126",
		"< B | content in file A, line 127",
		"< C | content in file A, line 128",
		"---",
		"> X | content in file B, line 130",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 3)

	expected[0] = SimilarDiffPair{
		Group:     'c',
		Left:      "A | content in file A, line 126",
		Right:     "X | content in file B, line 130",
		LeftLine:  126,
		RightLine: 130,
	}
	expected[1] = SimilarDiffPair{
		Group:     'd',
		Left:      "B | content in file A, line 127",
		Right:     "",
		LeftLine:  127,
		RightLine: 0,
	}
	expected[2] = SimilarDiffPair{
		Group:     'd',
		Left:      "C | content in file A, line 128",
		Right:     "",
		LeftLine:  128,
		RightLine: 0,
	}

	CheckTestData(t, s, 3, expected)
}

func TestCaptureChangedLinesManyRightSide(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"296c300,303",
		"< content in file A, line 296",
		"---",
		"> content in file B, line 300",
		"> content in file B, line 301",
		"> content in file B, line 302",
		"> content in file B, line 303",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 4)

	expected[0] = SimilarDiffPair{
		Group:     'c',
		Left:      "content in file A, line 296",
		Right:     "content in file B, line 300",
		LeftLine:  296,
		RightLine: 300,
	}
	expected[1] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "content in file B, line 301",
		LeftLine:  0,
		RightLine: 301,
	}
	expected[2] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "content in file B, line 302",
		LeftLine:  0,
		RightLine: 302,
	}
	expected[3] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "content in file B, line 303",
		LeftLine:  0,
		RightLine: 303,
	}

	CheckTestData(t, s, 4, expected)
}

func TestCaptureDeletedLinesOne(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"13d12",
		"< A | content in file A, line 13",
		"43d22",
		"< B | content in file A, line 43",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 2)

	expected[0] = SimilarDiffPair{
		Group:     'd',
		Left:      "A | content in file A, line 13",
		Right:     "",
		LeftLine:  13,
		RightLine: 0,
	}
	expected[1] = SimilarDiffPair{
		Group:     'd',
		Left:      "B | content in file A, line 43",
		Right:     "",
		LeftLine:  43,
		RightLine: 0,
	}

	CheckTestData(t, s, 2, expected)
}

func TestCaptureDeletedLinesMany(t *testing.T) {
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
		Group:     'd',
		Left:      "W     | content in file A, line 10",
		Right:     "",
		LeftLine:  10,
		RightLine: 0,
	}
	expected[1] = SimilarDiffPair{
		Group:     'd',
		Left:      "X     | content in file A, line 11",
		Right:     "",
		LeftLine:  11,
		RightLine: 0,
	}
	expected[2] = SimilarDiffPair{
		Group:     'd',
		Left:      "Y     | content in file A, line 12",
		Right:     "",
		LeftLine:  12,
		RightLine: 0,
	}
	expected[3] = SimilarDiffPair{
		Group:     'd',
		Left:      "Z     | content in file A, line 13",
		Right:     "",
		LeftLine:  13,
		RightLine: 0,
	}

	CheckTestData(t, s, 4, expected)
}

func TestCaptureAddedLinesOne(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"5a10",
		"> A | content in file B, line 10",
		"5a20",
		"> B | content in file B, line 20",
	}

	s.Total = len(s.Lines)

	s.CaptureChanges()

	expected := make([]SimilarDiffPair, 2)

	expected[0] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "A | content in file B, line 10",
		LeftLine:  0,
		RightLine: 10,
	}
	expected[1] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "B | content in file B, line 20",
		LeftLine:  0,
		RightLine: 20,
	}

	CheckTestData(t, s, 2, expected)
}

func TestCaptureAddedLinesMany(t *testing.T) {
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
		Group:     'a',
		Left:      "",
		Right:     "W     | content in file B, line 10",
		LeftLine:  0,
		RightLine: 10,
	}
	expected[1] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "X     | content in file B, line 11",
		LeftLine:  0,
		RightLine: 11,
	}
	expected[2] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "Y     | content in file B, line 12",
		LeftLine:  0,
		RightLine: 12,
	}
	expected[3] = SimilarDiffPair{
		Group:     'a',
		Left:      "",
		Right:     "Z     | content in file B, line 13",
		LeftLine:  0,
		RightLine: 13,
	}

	CheckTestData(t, s, 4, expected)
}

func TestDiscardSimilarities(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"13d12",
		"> A | content in file B, line 13",
		"173d172",
		"> B | content in file B, line 173",
	}

	s.Total = len(s.Lines)

	s.Changes = append(s.Changes, SimilarDiffChange{"content", "foo"})
	s.Changes = append(s.Changes, SimilarDiffChange{"file", "bar"})
	s.Changes = append(s.Changes, SimilarDiffChange{"line", "lorem"})

	s.CaptureChanges()

	s.DiscardSimilarities()

	expected := make([]SimilarDiffPair, 2)

	expected[0] = SimilarDiffPair{
		Group:     'd',
		Left:      "A | content in file B, line 13",
		Right:     "",
		LeftLine:  13,
		RightLine: 0,
	}
	expected[1] = SimilarDiffPair{
		Group:     'd',
		Left:      "B | content in file B, line 173",
		Right:     "",
		LeftLine:  173,
		RightLine: 0,
	}

	CheckTestData(t, s, 2, expected)
}

func TestNoConfiguration(t *testing.T) {
	s := NewSimilarDiff()

	s.Lines = []string{
		"13d12",
		"> A | content in file B, line 13",
	}

	s.Total = len(s.Lines)

	s.SetChanges("similardiff-not-found.ini")

	s.CaptureChanges()

	s.DiscardSimilarities()

	expected := make([]SimilarDiffPair, 1)

	expected[0] = SimilarDiffPair{
		Group:     'd',
		Left:      "A | content in file B, line 13",
		Right:     "",
		LeftLine:  13,
		RightLine: 0,
	}

	CheckTestData(t, s, 1, expected)
}
