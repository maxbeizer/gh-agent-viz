Output docs/demo.gif

Set FontSize 14
Set Width 1400
Set Height 800
Set Theme "Builtin Solarized Dark"

Type "gh agent-viz --demo"
Enter
Sleep 3s

# Dashboard is the landing page — browse the panels
Sleep 1s

# Tab through panels: Active → Recent → Attention → Repos
Type "	"
Sleep 1s
Type "	"
Sleep 1s
Type "	"
Sleep 1s
Type "j"
Sleep 400ms
Type "j"
Sleep 400ms

# Drill into a repo to see its sessions
Type "enter"
Sleep 2s

# Back to dashboard
Type "esc"
Sleep 1s

# Tab back to Active panel
Type "	"
Sleep 500ms
Type "	"
Sleep 500ms

# Drill into an active session
Type "enter"
Sleep 2s

# Check the logs
Type "l"
Sleep 2s

# Back to dashboard
Type "esc"
Sleep 1s

# Try the search
Type "/"
Sleep 500ms
Type "auth"
Sleep 1s
Type "esc"
Sleep 500ms

# Switch to kanban
Type "K"
Sleep 2s

# Tab through columns
Type "	"
Sleep 800ms
Type "	"
Sleep 800ms
Type "	"
Sleep 800ms

# Dismiss completed
Type "X"
Sleep 2s

# Back to dashboard
Type "esc"
Sleep 1s

# Take a snapshot
Type "S"
Sleep 1s

# Show help
Type "?"
Sleep 2s
Type "?"
Sleep 500ms

# Quit
Type "q"
Sleep 500ms
