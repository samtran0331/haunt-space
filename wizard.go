package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── types ─────────────────────────────────────────────────────────────────────

type wizardStep int

const (
	stepListTemplates wizardStep = iota
	stepChooseName
	stepChooseAction
	stepInputSize
	stepInputFolder
	stepInputCommand
	stepConfirm
	stepDone
	stepNotFound
)

type wizardMode int

const (
	modeNew  wizardMode = iota
	modeEdit
	modeCopy
)

type confirmKind int

const (
	confirmDeleteTemplate confirmKind = iota
	confirmDeletePane
	confirmExit
)

// pendingPath is a selectable leaf node in the current layout tree.
type pendingPath struct {
	path    []int
	paneNum int
}

type wizardModel struct {
	step         wizardStep
	prevStep     wizardStep
	mode         wizardMode
	root         *LayoutNode
	leaves       []pendingPath
	selectedLeaf int
	pendingDir   SplitDirection
	templateName string
	originalName string
	input        textinput.Model
	errorMsg     string
	saveMsg      string
	isDirty      bool
	cwd          string // directory where hsp was invoked
	tempFolder   string // folder staged between stepInputFolder and stepInputCommand
	launchName   string // template to launch after TUI exits
	notFoundName string // name attempted when not found
	// list view
	templateList []string
	listCursor   int
	// confirm
	confirmMsg  string
	confirmKind confirmKind
	// terminal size
	width  int
	height int
}

// ── styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	subtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	cursorStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
)

// ── tree helpers ──────────────────────────────────────────────────────────────

// collectLeaves walks the tree appending each leaf position (nil node or Direction==None).
func collectLeaves(node *LayoutNode, path []int, counter *int, out *[]pendingPath) {
	if node == nil || node.Direction == None {
		*out = append(*out, pendingPath{path: copyPath(path), paneNum: *counter})
		*counter++
		return
	}
	collectLeaves(node.LeftChild, append(path, 0), counter, out)
	collectLeaves(node.RightChild, append(path, 1), counter, out)
}

// buildLeaves returns all leaf nodes numbered from 1.
func buildLeaves(root *LayoutNode) []pendingPath {
	var out []pendingPath
	n := 1
	collectLeaves(root, []int{}, &n, &out)
	return out
}

// findLeafIndex returns the index in leaves whose path matches, or 0.
func findLeafIndex(leaves []pendingPath, path []int) int {
	for i, l := range leaves {
		if pathsEqual(l.path, path) {
			return i
		}
	}
	return 0
}

// deletePaneAt removes the leaf at path by promoting its sibling into the parent slot.
func deletePaneAt(root **LayoutNode, path []int) {
	if len(path) == 0 {
		*root = nil
		return
	}
	parentPath := path[:len(path)-1]
	childSide := path[len(path)-1]
	parentSlot := nodeAt(root, parentPath)
	parent := *parentSlot
	var sibling *LayoutNode
	if childSide == 0 {
		sibling = parent.RightChild
	} else {
		sibling = parent.LeftChild
	}
	if sibling == nil {
		*parentSlot = &LayoutNode{Direction: None}
	} else {
		*parentSlot = sibling
	}
}

// newNotFoundModel starts the wizard on the "template not found" screen.
func newNotFoundModel(name string) wizardModel {
	names, _ := listTemplateNames()
	return wizardModel{
		step:         stepNotFound,
		notFoundName: name,
		templateList: names,
		input:        newInput(),
	}
}

// ── node helpers ──────────────────────────────────────────────────────────────

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

func copyPath(p []int) []int {
	out := make([]int, len(p))
	copy(out, p)
	return out
}

func getNode(root *LayoutNode, path []int) *LayoutNode {
	cur := root
	for _, dir := range path {
		if cur == nil {
			return nil
		}
		if dir == 0 {
			cur = cur.LeftChild
		} else {
			cur = cur.RightChild
		}
	}
	return cur
}

// ── constructors ──────────────────────────────────────────────────────────────

func newInput() textinput.Model {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 40
	return ti
}

// newWizardModel starts in the template list view.
func newWizardModel() wizardModel {
	names, _ := listTemplateNames()
	cwd, _ := os.Getwd()
	return wizardModel{
		step:         stepListTemplates,
		templateList: names,
		input:        newInput(),
		cwd:          cwd,
	}
}

// newBooModel starts directly in new-template mode.
func newBooModel() wizardModel {
	m := newWizardModel()
	m.step = stepChooseName
	m.mode = modeNew
	m.input.Placeholder = "my-layout"
	return m
}

// newEditModel builds a model pre-loaded with an existing template.
func newEditModel(name string, root *LayoutNode, mode wizardMode) wizardModel {
	cwd, _ := os.Getwd()
	return wizardModel{
		step:         stepChooseAction,
		mode:         mode,
		root:         root,
		leaves:       buildLeaves(root),
		templateName: name,
		originalName: name,
		input:        newInput(),
		isDirty:      mode == modeCopy,
		cwd:          cwd,
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m wizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		if sz, ok := msg.(tea.WindowSizeMsg); ok {
			m.width = sz.Width
			m.height = sz.Height
			return m, nil
		}
		// Pass non-key events to text input for typing steps.
		switch m.step {
		case stepChooseName, stepInputSize, stepInputFolder, stepInputCommand:
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	key := keyMsg.String()
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.step {
	case stepListTemplates:
		return m.updateList(key)
	case stepChooseName:
		return m.updateChooseName(msg, key)
	case stepChooseAction:
		return m.updateChooseAction(key)
	case stepInputSize:
		return m.updateInputSize(msg, key)
	case stepInputFolder:
		return m.updateInputFolder(msg, key)
	case stepInputCommand:
		return m.updateInputCommand(msg, key)
	case stepConfirm:
		return m.updateConfirm(key)
	case stepNotFound:
		return m.updateNotFound(key)
	case stepDone:
		return m, tea.Quit
	}
	return m, nil
}

// ── List view ─────────────────────────────────────────────────────────────────

func (m wizardModel) updateList(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "x":
		return m, tea.Quit
	case "j", "down":
		if m.listCursor < len(m.templateList)-1 {
			m.listCursor++
		}
	case "k", "up":
		if m.listCursor > 0 {
			m.listCursor--
		}
	case "enter":
		if len(m.templateList) == 0 {
			break
		}
		m.launchName = m.templateList[m.listCursor]
		return m, tea.Quit
	case "n":
		m.step = stepChooseName
		m.mode = modeNew
		m.input.SetValue("")
		m.input.Placeholder = "my-layout"
	case "e":
		if len(m.templateList) == 0 {
			break
		}
		return m.openForEdit(m.templateList[m.listCursor], modeEdit)
	case "c":
		if len(m.templateList) == 0 {
			break
		}
		m.originalName = m.templateList[m.listCursor]
		m.mode = modeCopy
		m.step = stepChooseName
		m.input.SetValue("")
		m.input.Placeholder = m.originalName + "-copy"
	case "d":
		if len(m.templateList) == 0 {
			break
		}
		name := m.templateList[m.listCursor]
		m.confirmMsg = fmt.Sprintf("Delete template %q? This cannot be undone.", name)
		m.confirmKind = confirmDeleteTemplate
		m.prevStep = stepListTemplates
		m.step = stepConfirm
	}
	return m, nil
}

func (m wizardModel) updateNotFound(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "c":
		m.mode = modeNew
		m.input.SetValue(m.notFoundName)
		m.input.Placeholder = "my-layout"
		m.step = stepChooseName
	case "l":
		m.step = stepListTemplates
	case "q", "x":
		return m, tea.Quit
	}
	return m, nil
}

func (m wizardModel) openForEdit(name string, mode wizardMode) (tea.Model, tea.Cmd) {
	bp, err := loadBlueprint(name)
	if err != nil {
		m.errorMsg = "Load failed: " + err.Error()
		return m, nil
	}
	root := bp.Root
	em := newEditModel(name, &root, mode)
	return em, textinput.Blink
}

// ── Choose name ───────────────────────────────────────────────────────────────

func (m wizardModel) updateChooseName(msg tea.Msg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "x":
		if m.mode == modeEdit || m.mode == modeCopy {
			names, _ := listTemplateNames()
			m.templateList = names
			m.step = stepListTemplates
			return m, nil
		}
		return m, tea.Quit
	case "enter":
		return m.handleChooseName()
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m wizardModel) handleChooseName() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.input.Value())
	if name == "" {
		m.errorMsg = "Name cannot be empty."
		return m, nil
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		m.errorMsg = "Name must not contain path separators or '..'."
		return m, nil
	}

	if m.mode == modeNew || m.mode == modeCopy {
		if _, err := os.Stat(templatePath(name)); err == nil {
			m.errorMsg = fmt.Sprintf("Template %q already exists — choose a different name.", name)
			return m, nil
		}
	}

	m.errorMsg = ""
	m.templateName = name
	m.input.SetValue("")

	if m.mode == modeNew {
		m.root = nil
		m.leaves = buildLeaves(nil)
		m.selectedLeaf = 0
		m.isDirty = false
		m.step = stepChooseAction
		return m, nil
	}

	// modeCopy: load original, open in edit mode under new name
	bp, err := loadBlueprint(m.originalName)
	if err != nil {
		m.errorMsg = "Load failed: " + err.Error()
		return m, nil
	}
	root := bp.Root
	em := newEditModel(name, &root, modeCopy)
	em.originalName = m.originalName
	return em, textinput.Blink
}

// ── Choose action (main editing screen) ───────────────────────────────────────

func (m wizardModel) updateChooseAction(key string) (tea.Model, tea.Cmd) {
	m.saveMsg = "" // clear save banner on any action
	m.errorMsg = ""

	switch key {
	case "q", "x":
		if m.isDirty {
			m.confirmMsg = "You have unsaved changes. Exit without saving?"
			m.confirmKind = confirmExit
			m.prevStep = stepChooseAction
			m.step = stepConfirm
			return m, nil
		}
		if m.mode == modeEdit || m.mode == modeCopy {
			names, _ := listTemplateNames()
			m.templateList = names
			m.step = stepListTemplates
			return m, nil
		}
		return m, tea.Quit

	case "v", "h":
		m.pendingDir = Vertical
		if key == "h" {
			m.pendingDir = Horizontal
		}
		m.input.SetValue("")
		m.input.Placeholder = "e.g. 50"
		m.step = stepInputSize

	case "c":
		// Pre-fill folder with the node's current folder, or cwd as default.
		currentFolder := m.cwd
		if len(m.leaves) > 0 {
			if n := getNode(m.root, m.leaves[m.selectedLeaf].path); n != nil && n.Folder != "" {
				currentFolder = n.Folder
			}
		}
		m.input.SetValue(currentFolder)
		m.input.Placeholder = m.cwd
		m.step = stepInputFolder

	case "d":
		if len(m.leaves) <= 1 {
			m.errorMsg = "Cannot delete the only pane."
			return m, nil
		}
		sel := m.leaves[m.selectedLeaf]
		m.confirmMsg = fmt.Sprintf("Delete pane %d? Its sibling will take its place.", sel.paneNum)
		m.confirmKind = confirmDeletePane
		m.prevStep = stepChooseAction
		m.step = stepConfirm

	case "s":
		m.doSave()

	default:
		if target, err := strconv.Atoi(key); err == nil {
			for i, l := range m.leaves {
				if l.paneNum == target {
					m.selectedLeaf = i
					break
				}
			}
		}
	}
	return m, nil
}

func (m *wizardModel) doSave() {
	root := m.root
	if root == nil {
		root = &LayoutNode{Direction: None}
	}
	bp := GlobalBlueprint{
		TemplateName: m.templateName,
		Root:         *root,
	}
	if err := saveBlueprint(bp); err != nil {
		m.errorMsg = "Save failed: " + err.Error()
	} else {
		m.isDirty = false
		m.saveMsg = "✓ Saved"
	}
}

// ── Input size ────────────────────────────────────────────────────────────────

func (m wizardModel) updateInputSize(msg tea.Msg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		return m.handleInputSize()
	case "esc", "q":
		m.step = stepChooseAction
		return m, nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m wizardModel) handleInputSize() (tea.Model, tea.Cmd) {
	val, err := strconv.Atoi(strings.TrimSpace(m.input.Value()))
	if err != nil || val < 1 || val > 99 {
		m.errorMsg = "Enter a whole number between 1 and 99."
		return m, nil
	}
	m.errorMsg = ""
	if len(m.leaves) == 0 {
		return m, nil
	}
	path := copyPath(m.leaves[m.selectedLeaf].path)
	nodePtr := nodeAt(&m.root, path)
	(*nodePtr).Direction = m.pendingDir
	(*nodePtr).Size = val
	(*nodePtr).LeftChild = nil
	(*nodePtr).RightChild = nil

	m.leaves = buildLeaves(m.root)
	m.selectedLeaf = findLeafIndex(m.leaves, append(copyPath(path), 0))
	m.input.SetValue("")
	m.isDirty = true
	m.step = stepChooseAction
	return m, nil
}

// ── Input folder ──────────────────────────────────────────────────────────────

func (m wizardModel) updateInputFolder(msg tea.Msg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		return m.handleInputFolder()
	case "esc":
		m.step = stepChooseAction
		return m, nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m wizardModel) handleInputFolder() (tea.Model, tea.Cmd) {
	raw := strings.TrimSpace(m.input.Value())
	folder := raw
	m.errorMsg = ""

	if raw == "" {
		// Blank → home directory
		folder = "~"
		m.errorMsg = "Blank folder — pane will open at home directory (~)."
	}

	m.tempFolder = folder

	// Advance to command input, pre-filling current command if set.
	m.input.SetValue("")
	if len(m.leaves) > 0 {
		if n := getNode(m.root, m.leaves[m.selectedLeaf].path); n != nil {
			m.input.SetValue(n.Command)
		}
	}
	m.input.Placeholder = "e.g. lazygit  (leave blank for plain shell)"
	m.step = stepInputCommand
	return m, nil
}

// ── Input command ─────────────────────────────────────────────────────────────

func (m wizardModel) updateInputCommand(msg tea.Msg, key string) (tea.Model, tea.Cmd) {
	switch key {
	case "enter":
		return m.handleInputCommand()
	case "esc":
		m.step = stepChooseAction
		return m, nil
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m wizardModel) handleInputCommand() (tea.Model, tea.Cmd) {
	if len(m.leaves) == 0 {
		return m, nil
	}
	path := copyPath(m.leaves[m.selectedLeaf].path)
	command := strings.TrimSpace(m.input.Value())
	nodePtr := nodeAt(&m.root, path)
	(*nodePtr).Direction = None
	(*nodePtr).Command = command
	(*nodePtr).Folder = m.tempFolder
	m.tempFolder = ""

	m.leaves = buildLeaves(m.root)
	m.selectedLeaf = findLeafIndex(m.leaves, path)
	m.input.SetValue("")
	m.errorMsg = ""
	m.isDirty = true
	m.step = stepChooseAction
	return m, nil
}

// ── Confirm ───────────────────────────────────────────────────────────────────

func (m wizardModel) updateConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y":
		return m.executeConfirm()
	case "n", "N", "esc", "q", "x":
		m.step = m.prevStep
		return m, nil
	}
	return m, nil
}

func (m wizardModel) executeConfirm() (tea.Model, tea.Cmd) {
	switch m.confirmKind {
	case confirmDeleteTemplate:
		name := m.templateList[m.listCursor]
		os.Remove(templatePath(name))
		names, _ := listTemplateNames()
		m.templateList = names
		if m.listCursor >= len(m.templateList) && m.listCursor > 0 {
			m.listCursor--
		}
		m.step = stepListTemplates

	case confirmDeletePane:
		if len(m.leaves) > 0 {
			deletePaneAt(&m.root, m.leaves[m.selectedLeaf].path)
			m.leaves = buildLeaves(m.root)
			if m.selectedLeaf >= len(m.leaves) {
				m.selectedLeaf = 0
			}
			m.isDirty = true
		}
		m.step = stepChooseAction

	case confirmExit:
		return m, tea.Quit
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m wizardModel) View() string {
	switch m.step {
	case stepListTemplates:
		var b strings.Builder
		b.WriteString(titleStyle.Render("✦ haunt-space") + "\n\n")
		b.WriteString(m.viewList())
		if m.errorMsg != "" {
			b.WriteString("\n" + errorStyle.Render("⚠  "+m.errorMsg) + "\n")
		}
		return b.String()
	case stepChooseAction:
		var b strings.Builder
		b.WriteString(titleStyle.Render("✦ haunt-space") + "\n\n")
		b.WriteString(m.viewChooseAction())
		if m.errorMsg != "" {
			b.WriteString("\n" + errorStyle.Render("⚠  "+m.errorMsg) + "\n")
		}
		return b.String()
	case stepChooseName:
		return m.renderModal("✦ haunt-space", m.modalChooseName())
	case stepInputSize:
		return m.renderModal("Split Size", m.modalInputSize())
	case stepInputFolder:
		return m.renderModal("Pane Folder", m.modalInputFolder())
	case stepInputCommand:
		return m.renderModal("Pane Command", m.modalInputCommand())
	case stepConfirm:
		return m.renderModal("Confirm", m.modalConfirm())
	case stepNotFound:
		return m.renderModal("✦ haunt-space", m.modalNotFound())
	case stepDone:
		return m.renderModal("✦ haunt-space", m.modalDone())
	}
	return ""
}

func (m wizardModel) viewList() string {
	var b strings.Builder
	b.WriteString("Templates\n\n")
	if len(m.templateList) == 0 {
		b.WriteString(subtleStyle.Render("No templates yet.") + "\n\n")
	} else {
		for i, name := range m.templateList {
			if i == m.listCursor {
				b.WriteString(cursorStyle.Render("▶ "+name) + "\n")
			} else {
				b.WriteString("  " + subtleStyle.Render(name) + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(activeStyle.Render("[↵]") + " Launch   ")
		b.WriteString(activeStyle.Render("[e]") + " Edit   ")
		b.WriteString(activeStyle.Render("[c]") + " Copy   ")
		b.WriteString(activeStyle.Render("[d]") + " Delete\n")
	}
	b.WriteString(activeStyle.Render("[n]") + " New   ")
	b.WriteString(subtleStyle.Render("[j/k]") + " Navigate   ")
	b.WriteString(subtleStyle.Render("[q]") + " Quit\n")
	return b.String()
}

// ── Modal renderer ────────────────────────────────────────────────────────────

func (m wizardModel) renderModal(title, body string) string {
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(54)
	content := titleStyle.Render(title) + "\n\n" + body
	box := modalStyle.Render(content)
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}

func (m wizardModel) modalChooseName() string {
	var b strings.Builder
	switch m.mode {
	case modeNew:
		b.WriteString("New template name:\n\n")
	case modeCopy:
		b.WriteString(fmt.Sprintf("Copy of %q\nNew name:\n\n", m.originalName))
	}
	b.WriteString(m.input.View() + "\n")
	if m.errorMsg != "" {
		b.WriteString("\n" + errorStyle.Render("⚠  "+m.errorMsg) + "\n")
	}
	b.WriteString("\n" + subtleStyle.Render("[↵] confirm   [x] cancel"))
	return b.String()
}

func (m wizardModel) modalInputSize() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Size %% for %s split (1–99):\n\n", m.pendingDir))
	b.WriteString(m.input.View() + "\n")
	if m.errorMsg != "" {
		b.WriteString("\n" + errorStyle.Render("⚠  "+m.errorMsg) + "\n")
	}
	b.WriteString("\n" + subtleStyle.Render("[↵] confirm   [esc] cancel"))
	return b.String()
}

func (m wizardModel) modalInputFolder() string {
	var b strings.Builder
	b.WriteString("Folder for this pane:\n")
	b.WriteString(subtleStyle.Render("blank = home   relative paths OK") + "\n\n")
	b.WriteString(m.input.View() + "\n")
	if m.errorMsg != "" {
		b.WriteString("\n" + errorStyle.Render(m.errorMsg) + "\n")
	}
	b.WriteString("\n" + subtleStyle.Render("[↵] confirm   [esc] cancel"))
	return b.String()
}

func (m wizardModel) modalInputCommand() string {
	var b strings.Builder
	folderLabel := m.tempFolder
	if folderLabel == "" {
		folderLabel = m.cwd
	} else if folderLabel == "~" {
		folderLabel = "~ (home)"
	}
	b.WriteString(subtleStyle.Render("Folder: "+folderLabel) + "\n\n")
	b.WriteString("Command (blank = plain shell):\n\n")
	b.WriteString(m.input.View() + "\n")
	if m.errorMsg != "" {
		b.WriteString("\n" + errorStyle.Render("⚠  "+m.errorMsg) + "\n")
	}
	b.WriteString("\n" + subtleStyle.Render("[↵] confirm   [esc] cancel"))
	return b.String()
}

func (m wizardModel) modalConfirm() string {
	var b strings.Builder
	b.WriteString(warningStyle.Render("⚠  "+m.confirmMsg) + "\n\n")
	b.WriteString(activeStyle.Render("[y]") + " Yes   " + subtleStyle.Render("[n]") + " No\n")
	return b.String()
}

func (m wizardModel) modalNotFound() string {
	var b strings.Builder
	b.WriteString(warningStyle.Render(fmt.Sprintf("Template %q not found.\n\n", m.notFoundName)))
	b.WriteString(activeStyle.Render("[c]") + " Create new template\n")
	b.WriteString(activeStyle.Render("[l]") + " List templates\n")
	b.WriteString(activeStyle.Render("[x]") + " Exit\n")
	return b.String()
}

func (m wizardModel) modalDone() string {
	return successStyle.Render("✓ Saved — press any key to exit.")
}

func (m wizardModel) viewChooseAction() string {
	var b strings.Builder

	// Template name with dirty indicator
	nameLabel := m.templateName
	if m.isDirty {
		nameLabel += " " + warningStyle.Render("●")
	}
	b.WriteString(nameLabel + "\n\n")

	if len(m.leaves) > 0 {
		b.WriteString(fmt.Sprintf("Pane %d selected\n\n", m.leaves[m.selectedLeaf].paneNum))
	}

	b.WriteString(activeStyle.Render("[v]") + " Vertical split\n")
	b.WriteString(activeStyle.Render("[h]") + " Horizontal split\n")
	b.WriteString(activeStyle.Render("[c]") + " Set pane command\n")
	b.WriteString(activeStyle.Render("[d]") + " Delete pane\n")
	b.WriteString(activeStyle.Render("[s]") + " Save\n")
	b.WriteString(activeStyle.Render("[x]") + " Exit\n")

	// Current pane folder + command info
	if len(m.leaves) > 0 {
		num := m.leaves[m.selectedLeaf].paneNum
		paneFolder := "(current directory)"
		paneCmd := "(not set)"
		if n := getNode(m.root, m.leaves[m.selectedLeaf].path); n != nil {
			if n.Folder == "~" {
				paneFolder = "~ (home)"
			} else if n.Folder != "" {
				paneFolder = n.Folder
			}
			if n.Command != "" {
				paneCmd = n.Command
			}
		}
		b.WriteString(subtleStyle.Render(fmt.Sprintf("\n     Pane %d folder:  %s", num, paneFolder)) + "\n")
		b.WriteString(subtleStyle.Render(fmt.Sprintf("     Pane %d command: %s", num, paneCmd)) + "\n")
	}

	// Save confirmation banner
	if m.saveMsg != "" {
		b.WriteString(successStyle.Render(m.saveMsg) + "\n")
	}

	// Pane selector
	if len(m.leaves) > 1 {
		b.WriteString("\n")
		for i, l := range m.leaves {
			if i == m.selectedLeaf {
				b.WriteString(activeStyle.Render(fmt.Sprintf("[%d]", l.paneNum)) + " ")
			} else {
				b.WriteString(subtleStyle.Render(fmt.Sprintf("[%d]", l.paneNum)) + " ")
			}
		}
		b.WriteString(subtleStyle.Render(" ← number to select") + "\n")
	}

	b.WriteString("\n" + m.renderPreview())
	return b.String()
}

// ── Preview ───────────────────────────────────────────────────────────────────

func (m *wizardModel) renderPreview() string {
	const cw, ch = 90, 27
	c := NewCanvas(cw, ch)
	var activePath []int
	if len(m.leaves) > 0 {
		activePath = m.leaves[m.selectedLeaf].path
	}
	if m.root != nil {
		RenderNodePreview(&c, m.root, 0, 0, cw, ch, []int{}, m.leaves, activePath)
	} else {
		c.DrawBox(0, 0, cw, ch)
		if len(m.leaves) > 0 {
			c.WriteText(0, 0, cw, ch, fmt.Sprintf("►%d◄", m.leaves[0].paneNum))
		}
	}
	rendered := subtleStyle.Render(c.Render())
	if len(m.leaves) > 0 {
		marker := fmt.Sprintf("►%d◄", m.leaves[m.selectedLeaf].paneNum)
		rendered = strings.ReplaceAll(rendered, marker, activeStyle.Render(marker))
	}
	return rendered + "\n"
}

// ── Entry points ──────────────────────────────────────────────────────────────

func runTUI(initial wizardModel) error {
	p := tea.NewProgram(initial, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return err
	}
	if m, ok := final.(wizardModel); ok && m.launchName != "" {
		return launchTemplate(m.launchName)
	}
	return nil
}

// runWizard starts in the template list view (summon with no arg).
func runWizard() error {
	return runTUI(newWizardModel())
}

// runBoo starts directly in new-template mode.
func runBoo() error {
	return runTUI(newBooModel())
}

// runNotFound shows the "template not found" screen.
func runNotFound(name string) error {
	return runTUI(newNotFoundModel(name))
}
