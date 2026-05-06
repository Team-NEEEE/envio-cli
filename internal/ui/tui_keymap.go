package ui

import (
	"github.com/charmbracelet/bubbles/key"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

type tuiKeyMap struct {
	Confirm key.Binding
	Help    key.Binding
	Quit    key.Binding
}

func newTUIKeyMap(lang i18n.Language) tuiKeyMap {
	if lang == i18n.English {
		return tuiKeyMap{
			Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "close when done"), key.WithDisabled()),
			Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
			Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		}
	}
	return tuiKeyMap{
		Confirm: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "완료 후 닫기"), key.WithDisabled()),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "도움말")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "종료")),
	}
}

func (k tuiKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Help, k.Quit}
}

func (k tuiKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm},
		{k.Help, k.Quit},
	}
}
