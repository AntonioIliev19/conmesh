package cli

import (
	"strings"
	"testing"

	"github.com/AntonioIliev19/conmesh/internal/packages"
)

func TestCellsCountsTerminalColumns(t *testing.T) {
	for text, want := range map[string]int{
		"abc":       3,
		"文言文":       6,
		"✓ ok":      4,
		"🔴 high":    7,
		"é":        1, // e + combining acute accent
		"a‍b":       2, // zero-width joiner
		"Ｆｕｌｌ":      8,
		"› Caveman": 9,
	} {
		if got := cells(text); got != want {
			t.Errorf("cells(%q) = %d, want %d", text, got, want)
		}
	}
	if got := fit("ab文言文cd", 6); got != "ab文~" || cells(got) > 6 {
		t.Fatalf("fit = %q (%d columns)", got, cells(got))
	}
}

// Every picker line must fit the terminal, or the line wraps and each redraw
// leaves a stale copy behind. The Caveman pack has Chinese text and emoji.
func TestPackPickerLinesFitWithWideCharacters(t *testing.T) {
	pack, err := packages.PackByID("caveman")
	if err != nil {
		t.Fatal(err)
	}
	options := []menuOption{{group: pack.Title, subtitle: packOrigin(pack), name: "Whole pack", description: packCounts(pack)}}
	for _, e := range pack.Members {
		options = append(options, menuOption{group: "Or just one", name: e.Title, description: e.Type + "  " + e.Description,
			sameAs: []string{"conmesh " + e.Type + " add " + e.Source + " --trust"}})
	}
	for _, width := range []int{40, 80, 117, 120, 200} {
		for selected := range options {
			for _, line := range strings.Split(renderPicker(terminal{}, width, options, selected), "\n") {
				if cells(line) > width-1 {
					t.Fatalf("width %d: %d-column line wraps: %q", width, cells(line), line)
				}
			}
		}
	}
}
