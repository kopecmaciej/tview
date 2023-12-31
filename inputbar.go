package tview

import (
	"sort"
	"sync"

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

	fieldWidth int

	textArea *TextArea

	autocompleteListMutex sync.Mutex
	autocompleteList      *List

	autocompleteFunc func(word string, pos int) (items []AutocompleteItem)

	// An optional function which is called when the user selects an
	// autocomplete entry. The text and index of the selected entry (within the
	// list) is provided, as well as the user action causing the selection (one
	// of the "Autocompleted" values). The function should return true if the
	// autocomplete list should be closed. If nil, the input field will be
	// updated automatically when the user navigates the autocomplete list.
	autocompleted func(text string, index int, source int) bool

	// An optional function which may reject the last character that was entered.
	accept func(text string, ch rune) bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user indicated that they
	// are done entering text. The key which was pressed is provided (tab,
	// shift-tab, enter, or escape).
	done func(tcell.Key)

	// A callback function set by the Form class and called when the user leaves
	// this form item.
	finished func(tcell.Key)
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

func (e *InputBar) SetAutocompleteFunc(callback func(word string, pos int) (items []AutocompleteItem)) *InputBar {
	e.autocompleteFunc = callback
	e.Autocomplete()
	return e
}

func (e *InputBar) Autocomplete() *InputBar {
	if e.autocompleteFunc == nil {
		return e
	}

	// clear autocomplete list
	e.autocompleteList.Clear()

	text := e.textArea.GetText()
	if text == "" {
		return e
	}
	_, _, toRow, _ := e.textArea.GetCursor()

	items := e.autocompleteFunc(text, toRow)
	if len(items) == 0 {
		return e
	}

	e.SetAutocompleteList(items)

	return e
}

func (e *InputBar) GetCursor() (int, int, int, int) {
	return e.textArea.GetCursor()
}

// check if list is on the screen
func (e *InputBar) IsAutocompleteVisible() bool {
	return e.autocompleteList.GetItemCount() > 0
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
	labelWidth := e.textArea.GetLabelWidth()
	if labelWidth == 0 {
		labelWidth = TaggedStringWidth(e.textArea.GetLabel())
	}
	fieldWidth := e.fieldWidth
	if fieldWidth == 0 {
		fieldWidth = width - labelWidth
	}

	// Resize text area.
	e.textArea.SetRect(x, y, labelWidth+fieldWidth, 1)
	e.textArea.setMinCursorPadding(fieldWidth-1, 1)

	// Draw text area.
	e.textArea.hasFocus = e.HasFocus() // Force cursor positioning.
	e.textArea.Draw(screen)

	// Draw autocomplete list.
	e.autocompleteListMutex.Lock()
	defer e.autocompleteListMutex.Unlock()
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
}

func (e *InputBar) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return e.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {

		// Trigger changed events.
		var skipAutocomplete bool
		currentText := e.textArea.GetText()
		defer func() {
			newText := e.textArea.GetText()
			if newText != currentText {
				if !skipAutocomplete {
					e.Autocomplete()
				}
				if e.changed != nil {
					e.changed(newText)
				}
			}
		}()

		if e.autocompleteList != nil {
			e.autocompleteList.SetChangedFunc(nil)
			e.autocompleteList.SetSelectedFunc(nil)
			switch key := event.Key(); key {
			case tcell.KeyEscape: // Close the list.
				e.autocompleteList = nil
				return
			case tcell.KeyTab, tcell.KeyDown, tcell.KeyUp, tcell.KeyBackspace:
				e.autocompleteList.SetChangedFunc(func(index int, text, secondaryText string, shortcut rune) {
					text = stripTags(text)
					if e.autocompleted != nil {
						if e.autocompleted(text, index, AutocompletedNavigate) {
							e.autocompleteList = nil
							currentText = e.GetText()
						}
					} else {
						e.SetText(text)
						currentText = stripTags(text) // We want to keep the autocomplete list open and unchanged.
					}
				})
				e.autocompleteList.InputHandler()(event, setFocus)
				return
				// If the user presses the enter key, select the currently highlighted
				// autocomplete item, put it into the input field, and close the
				// autocomplete list.
			case tcell.KeyEnter:
				index := e.autocompleteList.GetCurrentItem()
				main, _ := e.autocompleteList.GetItemText(index)
				e.SetText(main)
				e.autocompleteList.Clear()
			}
		}
		e.textArea.InputHandler()(event, setFocus)
	})
}
