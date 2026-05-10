package e2ediscover

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	componentv1 "github.com/plantonhq/openmcf/apis/org/openmcf/qa/componente2eprofile/v1"
	"github.com/plantonhq/openmcf/pkg/e2e/profile"
)

var (
	colorGreen   = lipgloss.Color("#69DB7C")
	colorYellow  = lipgloss.Color("#FFD43B")
	colorGray    = lipgloss.Color("#868E96")
	colorDimGray = lipgloss.Color("#495057")
	colorBlue    = lipgloss.Color("#74C0FC")
	colorWhite   = lipgloss.Color("#DEE2E6")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite)

	tierHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBlue)

	greenStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	yellowStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	grayStyle = lipgloss.NewStyle().
			Foreground(colorGray)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDimGray)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDimGray).
			Italic(true)

	summaryStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite)
)

type model struct {
	result     *profile.DiscoverResult
	lines      []styledLine
	cursor     int
	offset     int
	height     int
	width      int
	filter     string
	filtering  bool
	quitting   bool
	exportJSON bool
}

type styledLine struct {
	text     string
	isTier   bool
	isSpacer bool
}

// RunInteractive launches the bubbletea TUI for E2E discovery results.
func RunInteractive(result *profile.DiscoverResult) error {
	m := newModel(result)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	if fm, ok := finalModel.(model); ok && fm.exportJSON {
		matrix := profile.BuildGitHubMatrix(result)
		jsonStr, err := profile.MatrixJSON(matrix)
		if err != nil {
			return err
		}
		fmt.Println(jsonStr)
	}

	return nil
}

func newModel(result *profile.DiscoverResult) model {
	lines := buildLines(result, "")
	return model{
		result: result,
		lines:  lines,
		height: 24,
		width:  80,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			return m.handleFilterKey(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "g":
			m.exportJSON = true
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.lines)-1 {
				m.cursor++
			}
		case "pgup":
			m.cursor -= m.viewportHeight()
			if m.cursor < 0 {
				m.cursor = 0
			}
		case "pgdown":
			m.cursor += m.viewportHeight()
			if m.cursor >= len(m.lines) {
				m.cursor = len(m.lines) - 1
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.lines) - 1
		case "/":
			m.filtering = true
			m.filter = ""
		}
	}

	m.adjustOffset()
	return m, nil
}

func (m model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filtering = false
		return m, nil
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.filter += msg.String()
		}
	}

	m.lines = buildLines(m.result, m.filter)
	if m.cursor >= len(m.lines) {
		m.cursor = max(0, len(m.lines)-1)
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	provName := "unknown"
	if m.result.Provider != nil && m.result.Provider.Metadata != nil {
		provName = m.result.Provider.Metadata.Name
	}

	b.WriteString(titleStyle.Render(fmt.Sprintf("  OpenMCF E2E Discovery — %s", provName)))
	b.WriteString("\n\n")

	if m.result.Provider != nil && m.result.Provider.Spec != nil {
		spec := m.result.Provider.Spec
		substrate := strings.TrimPrefix(spec.TestSubstrate.String(), "test_substrate_")
		cost := strings.TrimPrefix(spec.DefaultCostClass.String(), "cost_class_")
		lane := strings.TrimPrefix(spec.DefaultScheduleLane.String(), "schedule_lane_")
		tools := strings.Join(spec.RequiredTools, ", ")

		b.WriteString(dimStyle.Render(fmt.Sprintf("  Substrate: %s  |  Cost: %s  |  Schedule: %s", substrate, cost, lane)))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("  Tools: %s", tools)))
		b.WriteString("\n\n")
	}

	vpHeight := m.viewportHeight()
	for i := m.offset; i < m.offset+vpHeight && i < len(m.lines); i++ {
		line := m.lines[i]
		prefix := "  "
		if i == m.cursor {
			prefix = "▸ "
		}

		if line.isSpacer {
			b.WriteString("\n")
			continue
		}

		if line.isTier {
			b.WriteString(tierHeaderStyle.Render(line.text))
		} else {
			b.WriteString(prefix + line.text)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	counts := profile.CountByStatus(m.result)
	b.WriteString(summaryStyle.Render(fmt.Sprintf("  %s %d  %s %d  %s %d  %s %d  ─  %d total",
		greenStyle.Render("●"), counts.Green,
		yellowStyle.Render("○"), counts.Deferred,
		grayStyle.Render("○"), counts.Skip,
		dimStyle.Render("○"), counts.Stub,
		counts.Total,
	)))
	b.WriteString("\n\n")

	if m.filtering {
		b.WriteString(fmt.Sprintf("  / %s█", m.filter))
	} else {
		b.WriteString(helpStyle.Render("  ↑↓ navigate  / filter  g github-matrix  q quit"))
	}
	b.WriteString("\n")

	return b.String()
}

func (m model) viewportHeight() int {
	// Reserve lines for: title(2) + provider info(3) + summary(2) + help(2)
	reserved := 9
	h := m.height - reserved
	if h < 5 {
		h = 5
	}
	return h
}

func (m *model) adjustOffset() {
	vpHeight := m.viewportHeight()
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+vpHeight {
		m.offset = m.cursor - vpHeight + 1
	}
}

func buildLines(result *profile.DiscoverResult, filter string) []styledLine {
	tierNames := map[int32]string{
		1: "Tier 1 — Native Kubernetes",
		2: "Tier 2 — Helm-based",
		3: "Tier 3 — Operator-dependent",
		4: "Tier 4 — Operators & Addons",
	}

	var lines []styledLine
	var currentTier int32

	// Column header
	lines = append(lines, styledLine{
		text: headerStyle.Render(fmt.Sprintf("  %-36s %-10s %-5s %s", "Component", "Status", "Prov", "Timeout")),
	})

	for _, ce := range result.Components {
		spec := ce.Profile.Spec
		if spec == nil {
			continue
		}

		if filter != "" && !strings.Contains(strings.ToLower(ce.Name), strings.ToLower(filter)) {
			continue
		}

		if spec.Tier != currentTier {
			currentTier = spec.Tier
			name := tierNames[currentTier]
			if name == "" {
				name = fmt.Sprintf("Tier %d", currentTier)
			}
			lines = append(lines, styledLine{isSpacer: true})
			lines = append(lines, styledLine{
				text:   fmt.Sprintf("  ─── %s %s", name, strings.Repeat("─", max(0, 58-len(name)))),
				isTier: true,
			})
		}

		statusStr := formatStatus(spec.Status)
		provStr := provisionerShorthand(spec.ValidatedProvisioners)
		timeout := fmt.Sprintf("%dm", spec.TimeoutMinutes)

		lines = append(lines, styledLine{
			text: fmt.Sprintf("%-36s %s  %-5s %s", ce.Name, statusStr, provStr, timeout),
		})
	}

	return lines
}

func formatStatus(s componentv1.ComponentE2EProfileSpec_Status) string {
	switch s {
	case componentv1.ComponentE2EProfileSpec_green:
		return greenStyle.Render("● GREEN ")
	case componentv1.ComponentE2EProfileSpec_deferred:
		return yellowStyle.Render("○ DEFER ")
	case componentv1.ComponentE2EProfileSpec_skip:
		return grayStyle.Render("○ SKIP  ")
	case componentv1.ComponentE2EProfileSpec_stub:
		return dimStyle.Render("○ STUB  ")
	default:
		return dimStyle.Render("? ???   ")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
