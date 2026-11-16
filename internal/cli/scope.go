package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/AntonioIliev19/conmesh/internal/config"
)

var errScopeCancelled = errors.New("install cancelled; nothing was changed")

// needsScope reports whether args install packages without saying where.
func needsScope(args []string) bool {
	if len(args) < 2 || args[1] != "add" || args[0] != "agent" && args[0] != "skill" && args[0] != "pack" {
		return false
	}
	for _, a := range args[2:] {
		if a == "--global" || a == "--project" || a == "-h" || a == "--help" {
			return false
		}
	}
	return true
}

// scopeOptions are the two answers to "install for whom?". Project scope is
// offered even outside a project, with a note, so the choice is always visible.
func scopeOptions(project string) []menuOption {
	here := "Only this project; saved in its .conmesh folder"
	if project == "" {
		here = "Needs a ConMesh project here; run conmesh init first"
	} else {
		here = fmt.Sprintf("Only %s; saved in its .conmesh folder", filepath.Base(project))
	}
	return []menuOption{
		{id: "project", name: "This project", description: here},
		{id: "global", name: "All my projects", description: "Global: every project on this computer picks it up on its next conmesh sync"},
	}
}

// askScope asks where to install, on the terminal the question is drawn on.
func askScope(input io.Reader, out io.Writer, t terminal, width, height int, raw func() (func() error, error), project string) (string, error) {
	options := scopeOptions(project)
	choice, err := pickMenu(bufio.NewReader(input), out, t, width, height, t.paint(styleBold, "Install for which projects?"), options, raw)
	if err != nil && !errors.Is(err, errMenuQuit) && !errors.Is(err, errMenuInterrupt) && !errors.Is(err, io.EOF) {
		return "", err
	}
	if choice < 0 || err != nil {
		return "", errScopeCancelled
	}
	return options[choice].id, nil
}

// promptScope adds --project or --global to an interactive install that named
// neither. Scripts, --no-input and redirected input keep the project default.
func promptScope(args []string, stderr io.Writer, g globalOptions) ([]string, error) {
	if !needsScope(args) || g.noInput || !isTerminal(os.Stdin) {
		return args, nil
	}
	file, ok := stderr.(*os.File)
	t := terminalForOptions(stderr, g.noColor)
	if !ok || !t.live {
		return args, nil
	}
	project := ""
	if cwd, err := os.Getwd(); err == nil {
		if root, _, err := config.Find(cwd); err == nil {
			project = root
		}
	}
	scope, err := askScope(os.Stdin, stderr, t, terminalWidth(file), terminalHeight(file), terminalMenuMode(os.Stdin), project)
	if err != nil {
		return nil, err
	}
	return append(append([]string{}, args...), "--"+scope), nil
}
