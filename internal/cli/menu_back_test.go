package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AntonioIliev19/conmesh/internal/packages"
)

const back = "\x7f" // Backspace

// Back steps out one screen at a time: review -> the pack's skills and agents
// -> the pack list -> the main menu. Nothing is installed on the way.
func TestMenuInstallBackReturnsToPreviousScreen(t *testing.T) {
	var out bytes.Buffer
	var calls [][]string
	input := moveTo(t, "install") +
		"\r" + down + down + "\r" + // first pack, its second item
		back + // scope screen -> the pack's items
		down + "\r" + "\r" + "\r" + // first item, this project, Cancel at review -> the pack's items
		back + // -> the pack list
		down + "\r" + back + // second pack, then back -> the pack list
		back + // -> the main menu
		"q"
	if err := runMenu(strings.NewReader(input), &out, terminal{}, "dev", func(args []string) {
		calls = append(calls, args)
	}, menuUpdates{}, false, nil, menuInfo{packages: menuPackages{packs: testPacks()}}); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("installed while going back: %v", calls)
	}
	screen := out.String()
	if n := strings.Count(screen, "Install the whole pack or one item?"); n != 4 {
		t.Fatalf("pack items shown %d times, want 4:\n%s", n, screen)
	}
	if n := strings.Count(screen, "Choose what to install"); n != 3 {
		t.Fatalf("pack list shown %d times, want 3:\n%s", n, screen)
	}
}

// Quitting deep in the install flow ends the menu instead of looping.
func TestMenuInstallQuitFromReview(t *testing.T) {
	var calls [][]string
	input := moveTo(t, "install") + "\r" + "\r" + "q"
	if err := runMenu(strings.NewReader(input), &bytes.Buffer{}, terminal{}, "dev", func(args []string) {
		calls = append(calls, args)
	}, menuUpdates{}, false, nil, menuInfo{packages: menuPackages{packs: testPacks()}}); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("commands = %v", calls)
	}
}

func TestCodeAwarePackIsFirstAndRecommended(t *testing.T) {
	packs, err := packages.Packs()
	if err != nil {
		t.Fatal(err)
	}
	if packs[0].ID != "conmesh-code-aware" || !packs[0].Recommended {
		t.Fatalf("first pack %s recommended=%v", packs[0].ID, packs[0].Recommended)
	}
	for _, p := range packs[1:] {
		if p.Recommended {
			t.Fatalf("%s is also recommended", p.ID)
		}
	}
	listing := packListing(packs)
	if !strings.HasPrefix(listing, "Code-aware (ConMesh index)\n  conmesh-code-aware") || !strings.Contains(listing, "[recommended]") {
		t.Fatalf("pack list:\n%s", listing)
	}
	var out bytes.Buffer
	input := moveTo(t, "install") + back + "q"
	if err := runMenu(strings.NewReader(input), &out, terminal{}, "dev", func([]string) {}, menuUpdates{}, false, nil, menuInfo{packages: menuPackages{packs: packs}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "> ConMesh code-aware (recommended)") {
		t.Fatalf("recommended pack not selected first:\n%s", out.String())
	}
}
