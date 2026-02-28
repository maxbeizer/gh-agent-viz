Output docs/demo.gif

Set FontSize 14
Set Width 1400
Set Height 800
Set Theme "Builtin Solarized Dark"
Set TypingSpeed 50ms

# Launch — lands on the dashboard
Type "gh agent-viz --demo"
Enter
Sleep 4s

# Navigate Active panel
Down
Sleep 600ms
Down
Sleep 600ms

# Tab to Recent panel
Tab
Sleep 1.5s

# Tab to Attention panel
Tab
Sleep 1.5s

# Tab to Repos panel
Tab
Sleep 1.5s

# Navigate repos
Down
Sleep 400ms
Down
Sleep 400ms

# Drill into repo → filtered list view
Enter
Sleep 2s

# Back to dashboard
Escape
Sleep 1s

# Drill into an active session (Active panel is focused after wrap)
Tab
Sleep 500ms
Tab
Sleep 500ms
Enter
Sleep 2.5s

# View logs
Type "l"
Sleep 2.5s

# Back to dashboard
Escape
Sleep 1s

# Fuzzy search
Type "/"
Sleep 500ms
Type@100ms "auth"
Sleep 1.5s
Escape
Sleep 1s

# Kanban view
Type "K"
Sleep 2s

# Tab through columns
Tab
Sleep 1s
Tab
Sleep 1s

# Dismiss done
Type "X"
Sleep 2s

# Back to dashboard
Escape
Sleep 1.5s

# Snapshot
Type "S"
Sleep 1.5s

# Help overlay
Type "?"
Sleep 2.5s
Type "?"
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
