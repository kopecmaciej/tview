package tview

import (
	"sort"

	"github.com/gdamore/tcell/v2"
)

// AutocompleteItem is an item in the autocomplete list.
// The Value is the text that will be inserted into the input field when the
// user selects this item. The Description is the text that will be displayed
// in the autocomplete list.
type AutocompleteItem struct {
	Value       string
	Description string
	Priority    int
}

// InputBar is a one-line box (three lines if there is a title) where the user
// can enter text.
type InputBar struct {
	*Box

	textArea *TextArea

	autocompleteList *List
}

func NewInputBar() *InputBar {
	autoList := NewList()
	// add abbility to show secondary text in the same line
	autoList.ShowSecondaryText(true)
	autoList.SetHighlightFullLine(true)
  autoList.SetInlined(true)

	return &InputBar{
		Box:              NewBox(),
		textArea:         NewTextArea(),
		autocompleteList: autoList,
	}
}

func (e *InputBar) SetLabel(label string) *InputBar {
	e.textArea.SetLabel(label)
	return e
}

func (e *InputBar) SetText(text string) *InputBar {
	e.textArea.SetText(text, true)
	return e
}

func (e *InputBar) GetText() string {
	return e.textArea.GetText()
}

func (e *InputBar) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return e.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if e.autocompleteList.GetItemCount() > 0 {
			switch event.Key() {
			case tcell.KeyTab, tcell.KeyDown, tcell.KeyUp, tcell.KeyBackspace, tcell.KeyEscape:
				e.autocompleteList.InputHandler()(event, setFocus)
				return
			}
		}
		e.textArea.InputHandler()(event, setFocus)
	})
}

func (e *InputBar) SetTextSurroudings(left, right string, offset int) *InputBar {
	e.textArea.SetTextSurroudings(left, right, offset)
	return e
}

func (e *InputBar) Focus(delegate func(p Primitive)) {
	e.textArea.Focus(delegate)
}

func (e *InputBar) HasFocus() bool {
	return e.textArea.HasFocus() || e.Box.HasFocus()
}

func (e *InputBar) SetAutocompleteList(items []AutocompleteItem) *InputBar {
	e.autocompleteList.Clear()

	// Sort items by priority
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority > items[j].Priority
	})

	for _, item := range items {
		e.autocompleteList.AddItem(item.Value, item.Description, 0, nil)
	}
	return e
}

// Draw draws this primitive onto the screen.
func (e *InputBar) Draw(screen tcell.Screen) {
	e.Box.DrawForSubclass(screen, e)

	x, y, width, height := e.GetInnerRect()
	if height < 1 || width < 1 {
		return
	}

	// Resize text area.
	e.textArea.setMinCursorPadding(width-1, 1)

	// Draw autocomplete list
	if e.autocompleteList != nil {
		listHeight := e.autocompleteList.GetItemCount()
		listWidth := 0

		// Find the longest item
		for i := 0; i < listHeight; i++ {
			main, secondary := e.autocompleteList.GetItemText(i)
			compinedWidth := TaggedStringWidth(main + secondary)
			if compinedWidth > listWidth {
				// add with padding
				listWidth = compinedWidth + 2
			}
		}
		listHeight = listHeight + 2

		lx := x + e.textArea.GetLabelWidth()
		ly := y + 1
		_, sheight := screen.Size()
		if ly+listHeight >= sheight && ly-2 > listHeight-ly {
			ly = y - listHeight
			if ly < 0 {
				ly = 0
			}
		}
		if ly+listHeight >= sheight {
			listHeight = sheight - ly
		}
		lx = lx + 5

		e.autocompleteList.SetRect(lx, ly, listWidth, listHeight)
		e.autocompleteList.Draw(screen)
	}

	e.textArea.SetRect(x, y, width, height)
	e.textArea.Draw(screen)
}
