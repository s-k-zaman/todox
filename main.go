package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	pflag "github.com/spf13/pflag"
)

type Args struct {
	add      string
	yes      bool
	todoFile string
	edit     bool
}

func parseFlags() Args {
	parsedFlags := Args{}
	parsedFlags.todoFile = "TODO.md"
	pflag.BoolVarP(&parsedFlags.yes, "yes", "y", false, "bypass confirm")
	pflag.BoolVarP(&parsedFlags.edit, "edit", "e", false, "edit in editor")
	pflag.StringVarP(&parsedFlags.add, "add", "a", "", "Add to to-dos")
	pflag.Parse()

	if parsedFlags.add == "" {
		args := pflag.Args()
		if len(args) > 0 {
			parsedFlags.add = args[0]
		}
	}

	return parsedFlags
}

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func get_working_dir() string {
	gitRoot, err := getGitRoot()
	if err != nil {
		gitRoot, _ = os.Getwd()
	}
	return gitRoot
}

func prepareTodo(todoPath string) {
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		err := os.WriteFile(todoPath, []byte("# TODOs\n\n- [ ] Add your first task here.\n"), 0644)
		if err != nil {
			log.Fatalln("Error creating TODO.md:", err)
		}
		// fmt.Println("Created TODO.md at:", todoPath)
	}
}

func getEditor() string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	return editor
}

func insideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func splitLines(s string) []string {
	lines := []string{}
	curr := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, curr)
			curr = ""
		} else {
			curr += string(r)
		}
	}
	if curr != "" {
		lines = append(lines, curr)
	}
	return lines
}

func containsLine(output, target string) bool {
	return slices.Contains(splitLines(output), target)
}

func openInTmux(editor, file string) {
	windowName := "TODOs"

	// Check if window exists
	checkCmd := exec.Command("tmux", "list-windows", "-F", "#{window_name}")
	output, err := checkCmd.Output()
	if err != nil {
		return
	}
	// Switch to
	if containsLine(string(output), windowName) {
		exec.Command("tmux", "select-window", "-t", windowName).Run()
		return
	}
	// Create new
	cmdStr := fmt.Sprintf("[[ -e %[1]q ]] && %[2]s %[1]q", file, editor)
	cmd := exec.Command("tmux", "neww",
		"-n", windowName,
		"-c", "#{pane_current_path}",
		cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func editTodos(todoPath string) error {
	editor := getEditor()
	if insideTmux() {
		openInTmux(editor, todoPath)
	} else {
		cmd := exec.Command(editor, todoPath)
		// Attach command’s input/output to the terminal
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()
	}
	return nil
}

func main() {
	args := parseFlags()
	w_dir := get_working_dir()
	todoPath := filepath.Join(w_dir, args.todoFile)
	prepareTodo(todoPath)
	if args.edit {
		editTodos(todoPath)
		return
	}
}
