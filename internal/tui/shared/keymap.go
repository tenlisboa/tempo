package shared

import "github.com/charmbracelet/bubbles/key"

type ListKeyMap struct {
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Run    key.Binding
	Logs   key.Binding
	Toggle key.Binding
	Quit   key.Binding
}

var ListKeys = ListKeyMap{
	New:    key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
	Edit:   key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	Delete: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
	Run:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run now")),
	Logs:   key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "logs")),
	Toggle: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "toggle")),
	Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

type FormKeyMap struct {
	Next   key.Binding
	Prev   key.Binding
	Save   key.Binding
	Cancel key.Binding
	Help   key.Binding
}

var FormKeys = FormKeyMap{
	Next:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	Prev:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
	Save:   key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "schedule help")),
}

type LogKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Back   key.Binding
	PgUp   key.Binding
	PgDown key.Binding
}

var LogKeys = LogKeyMap{
	Up:     key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
	Down:   key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
	Back:   key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc", "back")),
	PgUp:   key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "scroll up")),
	PgDown: key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdn", "scroll down")),
}
