package index

import (
	"strings"
	"testing"
)

func TestCompactKeepsNumberedMatchingLines(t *testing.T) {
	filler := strings.Repeat("    private static int Other() => 0;\n", 40)
	chunk := "namespace Core;\n\n/// Joins model years into runs.\npublic static class YearRuns\n{\n    public static string Join(int[] years) => \"\";\n" + filler + "    // years again\n    // years once more\n}"
	got := Compact([]Evidence{{Path: "a.cs", StartLine: 10, EndLine: 59, Snippet: chunk}}, "model YEARS unbroken")
	want := "12: /// Joins model years into runs.\n15: public static string Join(int[] years) => \"\";\n56: // years again"
	if got[0].Snippet != want || !got[0].Truncated || got[0].StartLine != 10 || got[0].EndLine != 59 {
		t.Fatalf("compact snippet = %q truncated=%v", got[0].Snippet, got[0].Truncated)
	}
}

func TestCompactFallsBackToFirstLinesAndBoundsWidth(t *testing.T) {
	long := strings.Repeat("x", CompactKeepBytes)
	got := Compact([]Evidence{{StartLine: 3, EndLine: 4, Snippet: "\n  public class Greeter " + long}}, "Hello")
	if got[0].Snippet != "4: public class Greeter "+long[:CompactLineBytes-len("public class Greeter ")] || !got[0].Truncated {
		t.Fatalf("fallback snippet = %q truncated=%v", got[0].Snippet, got[0].Truncated)
	}
	small := Evidence{StartLine: 1, EndLine: 2, Snippet: "class A\n{ int hello; }"}
	if whole := Compact([]Evidence{small}, "hello"); whole[0] != small {
		t.Fatalf("a small snippet must be returned unchanged: %+v", whole[0])
	}
}
