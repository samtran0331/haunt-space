package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// wizardStep enumerates the sequential states of the wizard.
type wizardStep int

const (
	stepChooseName   wizardStep = iota // enter template name
	stepChooseAction                   // choose split direction or leaf command
	stepInputSize                      // enter split percentage
	stepInputCommand                   // enter leaf command (optional)
	stepDone                           // template saved
)

// pendingPath represents a tree position that still needs to be configured.
type pendingPath struct {
	path []int
}

// wizardModel is the Bubble Tea application model for the template wizard.
type wizardModel struct {
	step         wizardStep
	root         *LayoutNode
	pending      []pendingPath // queue of tree paths awaiting configuration
	currentPath  []int
	pendingDir   SplitDirection
	templateName string
	input        textinput.Model
	errorMsg     string
	saved        bool
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
)

// newWizardModel creates a fresh wizard model.
func newWizardModel() wizardModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 40
	return wizardModel{
		step:  stepChooseName,
		input: ti,
	}
}

// nodeAt returns the node pointer at the given path within the tree,
// creating intermediate nodes as needed.
func nodeAt(root **LayoutNode, path []int) **LayoutNode {
	cur := root
	for _, dir := range path {
		if *cur == nil {
			*cur = &LayoutNode{}
		}
		if dir == 0 {
			cur = &(*cur).LeftChild
		} else {
			cur = &(*cur).RightChild
		}
	}
	if *cur == nil {
		*cur = &LayoutNode{}
	}
	return cur
}

// copyPath returns a deep copy of the path slice.
func copyPath(p []int) []int {
	out := make([]int, len(p))
	copy(out, p)
	return out
}

// advancePending removes the first pending path and sets the next current path.
func (m *wizardModel) advancePending() {
	if len(m.pending) > 0 {
		m.pending = m.pending[1:]
	}
	if len(m.pending) == 0 {
		// All nodes configured — save and finish.
		m.finalize()
		return
	}
	m.currentPath = copyPath(m.pending[0].path)
	m.step = stepChooseAction
}

// finalize serialises the blueprint and transitions to stepDone.
func (m *wizardModel) finalize() {
	if m.root == nil {
		m.root = &LayoutNode{}
	}
	bp := GlobalBlueprint{
		TemplateName: m.templateName,
		Root:         *m.root,
	}
	if err := saveBlueprint(bp); err != nil {
		m.errorMsg = "Save failed: " + err.Error()
	} else {
		m.saved = true
	}
	m.step = stepDone
}

// Init satisfies the tea.Model interface.
func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard events and state transitions.
func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.step == stepDone {
				return m, tea.Quit
			}
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		case "v", "h", "c":
			if m.step == stepChooseAction {
				return m.handleActionKey(msg.String())
			}
		}
	}

	if m.step == stepChooseName || m.step == stepInputSize || m.step == stepInputCommand {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleEnter processes the Enter key based on the current step.
func (m wizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case stepChooseName:
		name := strings.TrimSpace(m.input.Value())
		if name == "" {
			m.errorMsg = "Template name cannot be empty."
			return m, nil
		}
		if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
			m.errorMsg = "Template name must not contain path separators or '..'."
			return m, nil
		}
		m.templateName = name
		m.errorMsg = ""
		m.input.SetValue("")
		// Start at the root position.
		m.pending = []pendingPath{{path: []int{}}}
		m.currentPath = []int{}
		m.step = stepChooseAction

	case stepInputSize:
		val, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
		if err != nil || val < 1 || val > 99 {
			m.errorMsg = "Enter a whole number between 1 and 99."
			return m, nil
		}
		m.errorMsg = ""

		// Set the node at the current path.
		nodePtr := nodeAt(&m.root, m.currentPath)
		(*nodePtr).Direction = m.pendingDir
		(*nodePtr).Size = val

		// Replace current path in the queue with its two children.
		m.pending[0] = pendingPath{path: append(copyPath(m.currentPath), 0)}
		leftPath := copyPath(m.pending[0].path)
		rightPath := append(copyPath(m.currentPath), 1)
		m.pending = append([]pendingPath{{path: leftPath}, {path: rightPath}}, m.pending[1:]...)
		m.currentPath = copyPath(m.pending[0].path)
		m.input.SetValue("")
		m.step = stepChooseAction

	case stepInputCommand:
		cmd := strings.TrimSpace(m.input.Value())
		nodePtr := nodeAt(&m.root, m.currentPath)
		(*nodePtr).Direction = None
		(*nodePtr).Command = cmd
		m.input.SetValue("")
		m.errorMsg = ""
		m.advancePending()

	case stepDone:
		return m, tea.Quit
	}

	return m, nil
}

// handleActionKey handles the v/h/c shortcut keys during stepChooseAction.
func (m wizardModel) handleActionKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "v":
		m.pendingDir = Vertical
		m.errorMsg = ""
		m.input.SetValue("")
		m.input.Placeholder = "e.g. 30"
		m.step = stepInputSize
	case "h":
		m.pendingDir = Horizontal
		m.errorMsg = ""
		m.input.SetValue("")
		m.input.Placeholder = "e.g. 50"
		m.step = stepInputSize
	case "c":
		m.errorMsg = ""
		m.input.SetValue("")
		m.input.Placeholder = "e.g. lazygit  (leave blank for plain shell)"
		m.step = stepInputCommand
	}
	return m, nil
}

// View renders the current wizard state.
func (m wizardModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("✦ haunt-space wizard") + "\n\n")

	switch m.step {
	case stepChooseName:
		b.WriteString("Template name:\n")
		b.WriteString(m.input.View() + "\n")

	case stepChooseAction:
		b.WriteString(fmt.Sprintf("Configuring node at path %v\n\n", m.currentPath))
		b.WriteString(activeStyle.Render("[v]") + " Vertical split\n")
		b.WriteString(activeStyle.Render("[h]") + " Horizontal split\n")
		b.WriteString(activeStyle.Render("[c]") + " Set pane command (leaf)\n")
		b.WriteString("\n" + m.renderPreview())

	case stepInputSize:
		dir := string(m.pendingDir)
		b.WriteString(fmt.Sprintf("Size percentage for %s split (1–99):\n", dir))
		b.WriteString(m.input.View() + "\n")
		b.WriteString("\n" + m.renderPreview())

	case stepInputCommand:
		b.WriteString("Command to run in this pane (optional, press Enter to skip):\n")
		b.WriteString(m.input.View() + "\n")
		b.WriteString("\n" + m.renderPreview())

	case stepDone:
		if m.saved {
			b.WriteString(successStyle.Render(fmt.Sprintf(
				"✓ Template %q saved to %s\n", m.templateName, templatePath(m.templateName),
			)))
		} else {
			b.WriteString(m.errorMsg + "\n")
		}
		b.WriteString(subtleStyle.Render("\nPress Enter or q to exit.") + "\n")
	}

	if m.errorMsg != "" && m.step != stepDone {
		b.WriteString("\n" + errorStyle.Render("⚠ "+m.errorMsg) + "\n")
	}

	return b.String()
}

// renderPreview builds the ASCII canvas and returns it as a string.
func (m *wizardModel) renderPreview() string {
	const cw, ch = 60, 18
	c := NewCanvas(cw, ch)
	if m.root != nil {
		RenderNodePreview(&c, m.root, 0, 0, cw, ch, 0, m.currentPath)
	} else {
		c.DrawBox(0, 0, cw, ch)
		c.WriteText(0, 0, cw, ch, "[ new template ]")
	}
	return subtleStyle.Render(c.Render()) + "\n"
}

// runWizard starts the Bubble Tea program for the template wizard.
func runWizard() error {
	p := tea.NewProgram(newWizardModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
