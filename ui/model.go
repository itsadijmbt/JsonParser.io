package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TickMsg time.Time

type Node struct {
	Key      string
	Value    interface{}
	Children []*Node
}

func buildNode(key string, v interface{}) *Node {
	n := &Node{Key: key}
	switch vv := v.(type) {
	case map[string]interface{}:
		for k, val := range vv {
			n.Children = append(n.Children, buildNode(k, val))
		}
	case []interface{}:
		for i, val := range vv {
			n.Children = append(n.Children, buildNode(fmt.Sprintf("[%d]", i), val))
		}
	default:
		n.Value = vv
	}
	return n
}

func renderTreeLines(n *Node, prefix string, isTail bool, indent int) []string {

	var branch string
	if isTail {
		branch = "└" + strings.Repeat("─", indent)
	} else {
		branch = "├" + strings.Repeat("─", indent)
	}
	
	line := prefix + branch + " " + n.Key
	if n.Value != nil && len(n.Children) == 0 {
		line += fmt.Sprintf(": %v", n.Value)
	}

	lines := []string{line}

	var nextPrefix string
	if isTail {
		nextPrefix = prefix + strings.Repeat(" ", indent+2)
	} else {
		nextPrefix = prefix + "│" + strings.Repeat(" ", indent+1)
	}
	// recurse
	for i, c := range n.Children {
		childLines := renderTreeLines(c, nextPrefix, i == len(n.Children)-1, indent)
		lines = append(lines, childLines...)
	}
	return lines
}

type model struct {
	lines     []string
	displayed int
	indent    int
	viewport  viewport.Model
	ready     bool
	style     lipgloss.Style
}


func NewModel(tree interface{}) tea.Model {

	vp := viewport.New(0, 0)

	vp.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2)
	
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#BD93F9")).
		Margin(1, 2)

	root := buildNode("root", tree)
	allLines := renderTreeLines(root, "", true, 3)
	return &model{
		lines:     allLines,
		displayed: 0,
		indent:    3,
		viewport:  vp,
		ready:     false,
		style:     containerStyle,
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case TickMsg:
		if m.displayed < len(m.lines) {
			m.displayed++
			cmd = tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {
				return TickMsg(t)
			})
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.viewport.LineUp(1)
		case "down", "j":
			m.viewport.LineDown(1)
		case "pgup":
			m.viewport.LineUp(m.viewport.Height)
		case "pgdown":
			m.viewport.LineDown(m.viewport.Height)
		case "left", "h":
			if m.indent > 1 {
				m.indent--
			}
		case "right", "l":
			if m.indent < 8 {
				m.indent++
			}
		}

	case tea.WindowSizeMsg:

		width := msg.Width - 6
		height := msg.Height - 6
	
		style := m.viewport.Style
		m.viewport = viewport.New(width, height)
		m.viewport.Style = style
		m.ready = true
	}
	return m, cmd
}

func (m *model) View() string {
	if !m.ready {
		return ""
	}
	
	var sb strings.Builder
	for i := 0; i < m.displayed && i < len(m.lines); i++ {
		line := m.lines[i]
		
		connector := strings.Repeat("─", m.indent)
		line = strings.ReplaceAll(line, strings.Repeat("─", 3), connector)
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Render(line) + "\n")
	}
	m.viewport.SetContent(sb.String())


	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#555555")).
		Padding(0, 1).
		Render(" JSON TreeView Parser ")


	status := lipgloss.NewStyle().
		Padding(0, 1).
		Render(fmt.Sprintf("Indent: %d  |  Lines: %d/%d  |  q: quit", m.indent, m.displayed, len(m.lines)))


	view := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		m.viewport.View(),
		status,
	)
	return m.style.Render(view)
}
