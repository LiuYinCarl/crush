package dialog

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/ui/common"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/crush/internal/ui/styles"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/sahilm/fuzzy"
)

// SearchID is the identifier for the message search dialog.
const SearchID = "search"

// MessageSearchEntry is a searchable chat message shown in the message
// search dialog.
type MessageSearchEntry struct {
	// ID is the message ID used to locate the message in the chat list.
	ID string
	// Role is the message role ("user" or "assistant").
	Role string
	// Text is a single-line snippet of the message content.
	Text string
}

// Search is a dialog for searching chat messages and jumping to one.
type Search struct {
	com   *common.Common
	help  help.Model
	list  *list.FilterableList
	input textinput.Model

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

var _ Dialog = (*Search)(nil)

// NewSearch creates a new message search dialog for the given entries.
func NewSearch(com *common.Common, entries []MessageSearchEntry) *Search {
	s := &Search{com: com}

	h := help.New()
	h.Styles = com.Styles.DialogHelpStyles()
	s.help = h

	items := make([]list.FilterableItem, len(entries))
	for i, entry := range entries {
		items[i] = &SearchItem{
			Versioned: list.NewVersioned(),
			entry:     entry,
			t:         com.Styles,
		}
	}
	s.list = list.NewFilterableList(items...)
	s.list.Focus()
	s.list.SetSelected(0)

	s.input = textinput.New()
	s.input.SetVirtualCursor(false)
	s.input.Placeholder = "Type to filter"
	s.input.SetStyles(com.Styles.TextInput)
	s.input.Focus()

	s.keyMap.Select = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "jump to message"),
	)
	s.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next item"),
	)
	s.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous item"),
	)
	s.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑↓", "choose"),
	)
	s.keyMap.Close = CloseKey

	return s
}

// ID implements Dialog.
func (s *Search) ID() string {
	return SearchID
}

// HandleMsg implements Dialog.
func (s *Search) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, s.keyMap.Previous):
			s.list.Focus()
			if s.list.IsSelectedFirst() {
				s.list.SelectLast()
			} else {
				s.list.SelectPrev()
			}
			s.list.ScrollToSelected()
		case key.Matches(msg, s.keyMap.Next):
			s.list.Focus()
			if s.list.IsSelectedLast() {
				s.list.SelectFirst()
			} else {
				s.list.SelectNext()
			}
			s.list.ScrollToSelected()
		case key.Matches(msg, s.keyMap.Select):
			if item, ok := s.list.SelectedItem().(*SearchItem); ok && item != nil {
				return ActionJumpToMessage{MessageID: item.entry.ID}
			}
		default:
			prevValue := s.input.Value()
			var cmd tea.Cmd
			s.input, cmd = s.input.Update(msg)
			if value := s.input.Value(); value != prevValue {
				s.list.SetFilter(value)
				s.list.ScrollToTop()
				s.list.SetSelected(0)
			}
			return ActionCmd{cmd}
		}
	}
	return nil
}

// Cursor returns the cursor position relative to the dialog.
func (s *Search) Cursor() *tea.Cursor {
	return InputCursor(s.com.Styles, s.input.Cursor())
}

// Draw implements [Dialog].
func (s *Search) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := s.com.Styles
	width := max(0, min(defaultDialogMaxWidth, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	height := max(0, min(defaultDialogHeight, area.Dy()-t.Dialog.View.GetVerticalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()

	s.input.SetWidth(dialogInputTextWidth(t, s.input, innerWidth))
	listHeight, listTotalHeight, _ := sizeDialogList(t, s.list, innerWidth, height)

	rc := NewRenderContext(t, width)
	rc.Title = "Search Messages"
	rc.AddPart(t.Dialog.InputPrompt.Render(s.input.View()))
	bodyView := t.Dialog.List.Height(s.list.Height()).Render(s.list.Render())
	bodyView = joinScrollbar(t, bodyView, listHeight, listTotalHeight, listHeight, s.list.Offset())
	rc.AddPart(bodyView)
	rc.Help = renderDialogHelp(t, &s.help, s, innerWidth)

	view := rc.Render()
	cur := s.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// ShortHelp implements [help.KeyMap].
func (s *Search) ShortHelp() []key.Binding {
	return []key.Binding{
		s.keyMap.UpDown,
		s.keyMap.Select,
		s.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (s *Search) FullHelp() [][]key.Binding {
	return [][]key.Binding{s.ShortHelp()}
}

// SearchItem wraps a [MessageSearchEntry] to implement the [ListItem]
// interface.
type SearchItem struct {
	*list.Versioned
	entry   MessageSearchEntry
	t       *styles.Styles
	m       fuzzy.Match
	cache   map[int]string
	focused bool
}

var _ ListItem = (*SearchItem)(nil)

// Finished implements list.Item. Search items are render-stable outside of
// explicit SetFocused / SetMatch calls.
func (s *SearchItem) Finished() bool {
	return true
}

// Filter implements ListItem.
func (s *SearchItem) Filter() string {
	return s.entry.Text
}

// ID implements ListItem.
func (s *SearchItem) ID() string {
	return s.entry.ID
}

// SetFocused implements ListItem.
func (s *SearchItem) SetFocused(focused bool) {
	if s.focused == focused {
		return
	}
	s.cache = nil
	s.focused = focused
	if s.Versioned != nil {
		s.Bump()
	}
}

// SetMatch implements ListItem.
func (s *SearchItem) SetMatch(m fuzzy.Match) {
	if sameFuzzyMatch(s.m, m) {
		return
	}
	s.cache = nil
	s.m = m
	if s.Versioned != nil {
		s.Bump()
	}
}

// Render implements ListItem.
func (s *SearchItem) Render(width int) string {
	styles := ListItemStyles{
		ItemBlurred:     s.t.Dialog.NormalItem,
		ItemFocused:     s.t.Dialog.SelectedItem,
		InfoTextBlurred: s.t.Dialog.ListItem.InfoBlurred,
		InfoTextFocused: s.t.Dialog.ListItem.InfoFocused,
	}
	return renderItem(styles, s.entry.Text, s.entry.Role, s.focused, width, s.cache, &s.m)
}
