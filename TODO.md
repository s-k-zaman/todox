# TODOs
- [x] make it a CLI app
- [x] Store to-dos in markdown file.
- [x] Add options to change tmux window name and number.
- [x] Ask for confirmation on creating new TODO.md file.
- [x] add a task(should be able to add subtasks also)
- [ ] Delete a task (id should be line number)(show heading and subtasks if available on confirmation, and delete subtasks as well, if deleting parent task)
- [ ] Complete a task (id should be line number)(show heading and subtasks if available on confirmation, and delete subtasks as well, if deleting parent task)
- [ ] view tasks/todos(show headings as well)


 # BUGs
- [x] opening using keyboard shortcut in tmux if present in `TODOs` window causing it to freeze(need to press ESC to unfreeze)(fixed using tmux run shell `tmux neww....`)
