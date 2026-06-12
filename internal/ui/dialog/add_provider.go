package dialog

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/catwalk/pkg/catwalk"
	uv "github.com/charmbracelet/ultraviolet"

	"github.com/SecDuckOps/duckops/internal/config"
	"github.com/SecDuckOps/duckops/internal/ui/common"
	"github.com/SecDuckOps/duckops/internal/ui/util"
)

const AddProviderID = "add_provider"

type AddProvider struct {
	com    *common.Common
	fields []string
	inputs []textinput.Model
	focus  int
	width  int
	help   help.Model

	keyMap struct {
		Next     key.Binding
		Previous key.Binding
		Submit   key.Binding
		Close    key.Binding
	}
}

var _ Dialog = (*AddProvider)(nil)

func NewAddProvider(com *common.Common) (*AddProvider, tea.Cmd) {
	t := com.Styles
	m := &AddProvider{
		com:    com,
		fields: []string{"Provider Name", "Base URL", "Model ID"},
		width:  60,
		focus:  0,
	}

	m.inputs = make([]textinput.Model, 3)
	for i := range m.fields {
		in := textinput.New()
		in.SetVirtualCursor(false)
		in.SetStyles(com.Styles.TextInput)
		in.Prompt = "> "
		switch i {
		case 0:
			in.Placeholder = "e.g. ollama, lm-studio"
		case 1:
			in.Placeholder = "e.g. http://localhost:11434/v1"
		case 2:
			in.Placeholder = "e.g. llama3, qwen2.5-coder:7b"
		}
		m.inputs[i] = in
	}
	m.inputs[0].Focus()

	help := help.New()
	help.Styles = t.DialogHelpStyles()
	m.help = help

	m.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "tab"),
		key.WithHelp("↓/tab", "next"),
	)
	m.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "shift+tab"),
		key.WithHelp("↑/shift+tab", "previous"),
	)
	m.keyMap.Submit = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	)
	m.keyMap.Close = CloseKey

	return m, nil
}

func (m *AddProvider) ID() string { return AddProviderID }

func (m *AddProvider) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, m.keyMap.Previous):
			m.prevField()
		case key.Matches(msg, m.keyMap.Submit):
			if m.focus == len(m.inputs)-1 {
				return m.save()
			}
			m.nextField()
		case key.Matches(msg, m.keyMap.Next):
			m.nextField()
		default:
			var cmd tea.Cmd
			m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
			return ActionCmd{cmd}
		}
	}
	return nil
}

func (m *AddProvider) nextField() {
	m.inputs[m.focus].Blur()
	if m.focus < len(m.inputs)-1 {
		m.focus++
	} else {
		m.focus = 0
	}
	m.inputs[m.focus].Focus()
}

func (m *AddProvider) prevField() {
	m.inputs[m.focus].Blur()
	if m.focus > 0 {
		m.focus--
	} else {
		m.focus = len(m.inputs) - 1
	}
	m.inputs[m.focus].Focus()
}

func (m *AddProvider) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := m.com.Styles
	rc := NewRenderContext(t, m.width)
	rc.Title = "Add Custom Provider"

	var b strings.Builder
	b.WriteString(t.Dialog.PrimaryText.Render("Fill in the fields below:") + "\n\n")
	for i, field := range m.fields {
		label := t.Dialog.NormalItem.Render("  " + field + ":")
		if m.focus == i {
			label = t.Dialog.SelectedItem.Render("> " + field + ":")
		}
		b.WriteString(label + "\n")
		m.inputs[i].SetWidth(m.width - 6)
		b.WriteString("  " + m.inputs[i].View() + "\n\n")
	}
	b.WriteString(t.Dialog.HelpView.Render(m.help.View(m)))

	rc.AddPart(b.String())
	view := rc.Render()
	DrawCenterCursor(scr, area, view, m.Cursor())
	return m.Cursor()
}

func (m *AddProvider) Cursor() *tea.Cursor {
	return InputCursor(m.com.Styles, m.inputs[m.focus].Cursor())
}

func (m *AddProvider) save() Action {
	name := strings.TrimSpace(m.inputs[0].Value())
	baseURL := strings.TrimSpace(m.inputs[1].Value())
	modelID := strings.TrimSpace(m.inputs[2].Value())
	if name == "" || baseURL == "" || modelID == "" {
		return ActionCmd{util.ReportError(fmt.Errorf("all fields are required"))}
	}
	p := config.ProviderConfig{
		ID:      name,
		Name:    name,
		Type:    catwalk.TypeOpenAICompat,
		BaseURL: baseURL,
		Models: []catwalk.Model{{
			ID:               modelID,
			Name:             modelID,
			ContextWindow:    4096,
			DefaultMaxTokens: 2048,
		}},
	}
	err := m.com.Workspace.SetConfigField(config.ScopeGlobal, fmt.Sprintf("providers.%s", name), p)
	if err != nil {
		return ActionCmd{util.ReportError(fmt.Errorf("failed to save provider: %w", err))}
	}
	_ = m.com.Workspace.SetConfigField(config.ScopeGlobal,
		"models.local",
		config.SelectedModel{Provider: name, Model: modelID},
	)
	return ActionClose{}
}

func (m *AddProvider) ShortHelp() []key.Binding {
	return []key.Binding{m.keyMap.Previous, m.keyMap.Next, m.keyMap.Submit, m.keyMap.Close}
}

func (m *AddProvider) FullHelp() [][]key.Binding {
	return [][]key.Binding{{m.keyMap.Previous, m.keyMap.Next, m.keyMap.Submit, m.keyMap.Close}}
}
