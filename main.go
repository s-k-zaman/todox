package main

import (
	"bufio"
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
	w_dir        string
	todo_path    string
	add          string
	yes          bool
	todoFile     string
	edit         bool
	tmux_win     string
	tmux_win_num int
}

func parseFlags() Args {
	parsedFlags := Args{}
	pflag.BoolVarP(&parsedFlags.yes, "yes", "y", false, "bypass confirm")
	pflag.BoolVarP(&parsedFlags.edit, "edit", "e", false, "edit in editor")
	pflag.StringVarP(&parsedFlags.todoFile, "file", "f", "TODO.md", "markdown file name for TODOs")
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

func confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + " [y/N]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func project_name(w_dir string) string {
	// need to get working dir, do i need this?
	folders := strings.Split(w_dir, "/")
	return folders[len(folders)-1]
}

func tmuxSupportsPopup() bool {
	cmd := exec.Command("tmux", "display-popup", "-E", "true")
	err := cmd.Run()
	if err == nil {
		return true
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr := string(exitErr.Stderr)
		if strings.Contains(stderr, "unknown command") || strings.Contains(stderr, "unknown option") {
			return false
		}
	}
	return true
}

func relativePathWithTilde(target string) string {
	home, _ := os.UserHomeDir()
	if after, ok := strings.CutPrefix(target, home); ok {
		return "~" + after
	}
	cwd, _ := os.Getwd()
	rel, err := filepath.Rel(cwd, target)
	if err != nil {
		return target
	}
	return rel
}

func prepareTodo(args *Args) {
	if _, err := os.Stat(args.todo_path); os.IsNotExist(err) {
		if args.yes || confirm(fmt.Sprintf("File %s does not exist in %s/ Create it?", args.todoFile, relativePathWithTilde(args.w_dir))) {
			err := os.WriteFile(args.todo_path, []byte("# TODOs\n\n- [ ] Add your first task here.\n"), 0644)
			if err != nil {
				log.Fatalln("Error creating TODO.md:", err)
			}
			// fmt.Println("Created TODO.md at:", todoPath)
		} else {
			fmt.Println("Aborted!")
			os.Exit(0)
		}
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
			c := exec.Command("tmux", "select-window", "-t", num)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			c.Run()
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

func editTodos(args *Args) error {
	editor := getEditor()
	if insideTmux() {
		openInTmux(editor, args.todo_path, args.tmux_win, args.tmux_win_num)
	} else {
		cmd := exec.Command(editor, args.todo_path)
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
	args.w_dir = get_working_dir()
	args.todo_path = filepath.Join(args.w_dir, args.todoFile)
	prepareTodo(&args)
	if args.add != "" {
		content, err := os.ReadFile(args.todo_path)
		if err != nil {
			log.Fatalln("Error reading TODO.md:", err)
		}
		newTask := "\n- [ ] " + args.add
		newContent := string(content) + newTask
		err = os.WriteFile(args.todo_path, []byte(newContent), 0644)
		if err != nil {
			log.Fatalln("Error writing TODO.md:", err)
		}
		fmt.Println("Added task:", args.add)
		return
	}
	if args.edit {
		editTodos(&args)
		return
	}
}
