package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	pflag "github.com/spf13/pflag"
)

type Args struct {
	add          string
	yes          bool
	todoFile     string
	edit         bool
	tmux_win     string
	tmux_win_num int
}

func parseFlags() Args {
	parsedFlags := Args{}
	parsedFlags.todoFile = "TODO.md"
	pflag.BoolVarP(&parsedFlags.yes, "yes", "y", false, "bypass confirm")
	pflag.BoolVarP(&parsedFlags.edit, "edit", "e", false, "edit in editor")
	pflag.StringVarP(&parsedFlags.add, "add", "a", "", "Add to to-dos")
	pflag.StringVarP(&parsedFlags.tmux_win, "twin", "w", "TODOs", "tmux window name")
	pflag.IntVarP(&parsedFlags.tmux_win_num, "tnum", "n", 9, "tmux window number")
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

func openInTmux(editor, file string, windowName string, windowNumber int) {
	preferredNumber := strconv.Itoa(windowNumber)

	// CHECK IF WINDOW EXISTS
	checkCmd := exec.Command("tmux", "list-windows", "-F", "#{window_index}:#{window_name}")
	output, err := checkCmd.Output()
	if err != nil {
		return
	}
	// Switch to
	lines := strings.SplitSeq(string(output), "\n")
	for line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		num, name := parts[0], parts[1]

		// If a window already exists with the same name or preferred number, just switch to it
		if name == windowName {
			exec.Command("tmux", "select-window", "-t", num).Run()
			return
		}
	}

	// CREATE NEW
	cmdStr := fmt.Sprintf("[[ -e %[1]q ]] && %[2]s %[1]q", file, editor)
	args := []string{"neww", "-n", windowName}

	// Determine if preferred number is free, number taken → let tmux auto-assign
	assign_num := "-t " + preferredNumber
	if strings.Contains(string(output), preferredNumber+":") {
		assign_num = ""
	}
	if assign_num != "" {
		args = append(args, assign_num)
	}

	args = append(args, "-c", "#{pane_current_path}")

	args = append(args, cmdStr)
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func editTodos(todoPath string, args *Args) error {
	editor := getEditor()
	if insideTmux() {
		openInTmux(editor, todoPath, args.tmux_win, args.tmux_win_num)
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
		editTodos(todoPath, &args)
		return
	}
}
