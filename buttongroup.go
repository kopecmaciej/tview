package tview

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ButtonGroup implements a horizontal row of labeled blocks where exactly one
// option is selected at a time. It functions like a segmented control or radio
// group, displayed inline as filled blocks.
//
// Navigation: Left/Right arrows (or h/l) move the cursor within the group.
// Enter or Space confirms the selection. Tab/Backtab/Escape exit to the next
// or previous form field.
type ButtonGroup struct {
	*Box

	// The text to be displayed before the buttons.
	label string

	// The screen width of the label area. 0 means use the label's natural width.
	labelWidth int

	// The label style.
	labelStyle tcell.Style

	// The available options.
	options []string

	// The index of the currently selected option (persisted value).
	currentOption int

	// The index of the option under the cursor (only meaningful while focused).
	focusedOption int

	// Style for the selected option when the group does not have focus.
	selectedStyle tcell.Style

	// Style for options that are not selected.
	unselectedStyle tcell.Style

	// Style for the option under the cursor when the group has focus.
	cursorStyle tcell.Style

	// Whether the item is disabled / read-only.
	disabled bool

	// Called when the user changes the current option.
	changed func(option string, index int)

	// Called when the user indicates they are done (Tab, Backtab, Escape).
	done func(tcell.Key)

	// Set by the Form class; called when the user leaves this form item.
	finished func(tcell.Key)

	// Optional icons displayed above each option label in card mode.
	icons []string
}

// NewButtonGroup returns a new ButtonGroup with the given label, options, and
// initial selection. The changed callback is invoked whenever the selection
// changes.
func NewButtonGroup(label string, options []string, initialOption int, changed func(string, int)) *ButtonGroup {
	if initialOption < 0 || initialOption >= len(options) {
		initialOption = 0
	}
	return &ButtonGroup{
		Box:                  NewBox(),
		label:                label,
		options:              options,
		currentOption:        initialOption,
		focusedOption:        initialOption,
		changed:              changed,
		labelStyle:           tcell.StyleDefault.Foreground(Styles.SecondaryTextColor),
		selectedStyle:   tcell.StyleDefault.Background(Styles.BorderColor).Foreground(Styles.PrimaryTextColor),
		unselectedStyle: tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		cursorStyle:     tcell.StyleDefault.Background(Styles.FocusColor).Foreground(Styles.PrimitiveBackgroundColor),
	}
}

// GetLabel returns the label text.
func (b *ButtonGroup) GetLabel() string {
	return b.label
}

// SetLabel sets the label text.
func (b *ButtonGroup) SetLabel(label string) *ButtonGroup {
	b.label = label
	return b
}

// SetLabelWidth sets the screen width of the label. 0 uses the natural width.
func (b *ButtonGroup) SetLabelWidth(width int) *ButtonGroup {
	b.labelWidth = width
	return b
}

// SetLabelColor sets the foreground color of the label.
func (b *ButtonGroup) SetLabelColor(color tcell.Color) *ButtonGroup {
	b.labelStyle = b.labelStyle.Foreground(color)
	return b
}

// GetCurrentOption returns the currently selected option string and its index.
func (b *ButtonGroup) GetCurrentOption() (string, int) {
	if len(b.options) == 0 {
		return "", -1
	}
	return b.options[b.currentOption], b.currentOption
}

// SetCurrentOption sets the selected option by index.
func (b *ButtonGroup) SetCurrentOption(index int) *ButtonGroup {
	if index >= 0 && index < len(b.options) {
		b.currentOption = index
		b.focusedOption = index
	}
	return b
}

// SetChangedFunc sets a handler called when the selection changes.
func (b *ButtonGroup) SetChangedFunc(handler func(option string, index int)) *ButtonGroup {
	b.changed = handler
	return b
}

// SetDoneFunc sets a handler called when the user leaves the group (Tab, Backtab, Escape).
func (b *ButtonGroup) SetDoneFunc(handler func(key tcell.Key)) *ButtonGroup {
	b.done = handler
	return b
}

// SetFinishedFunc sets the callback invoked when the user leaves this form item.
func (b *ButtonGroup) SetFinishedFunc(handler func(key tcell.Key)) FormItem {
	b.finished = handler
	return b
}

// SetDisabled sets whether the item is disabled / read-only.
func (b *ButtonGroup) SetDisabled(disabled bool) FormItem {
	b.disabled = disabled
	if b.finished != nil {
		b.finished(-1)
	}
	return b
}

// SetFormAttributes sets attributes shared by all form items.
func (b *ButtonGroup) SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem {
	b.labelWidth = labelWidth
	b.SetLabelColor(labelColor)
	b.backgroundColor = bgColor
	// Only update text colors; backgrounds stay as set in the constructor so
	// that the three-level contrast (unselected/selected/cursor) is preserved.
	b.unselectedStyle = b.unselectedStyle.Foreground(fieldTextColor)
	b.selectedStyle = b.selectedStyle.Foreground(fieldTextColor)
	// Cursor uses FocusColor background — outer bg as foreground for contrast.
	b.cursorStyle = b.cursorStyle.Foreground(bgColor)
	return b
}

// GetFieldWidth returns 0 (flexible width).
func (b *ButtonGroup) GetFieldWidth() int {
	return 0
}

// SetSelectedStyle sets the style for the confirmed (current) option in flat mode.
func (b *ButtonGroup) SetSelectedStyle(style tcell.Style) *ButtonGroup {
	b.selectedStyle = style
	return b
}

// SetUnselectedStyle sets the style for options that are not selected or focused.
func (b *ButtonGroup) SetUnselectedStyle(style tcell.Style) *ButtonGroup {
	b.unselectedStyle = style
	return b
}

// SetCursorStyle sets the style for the option under the cursor while focused.
func (b *ButtonGroup) SetCursorStyle(style tcell.Style) *ButtonGroup {
	b.cursorStyle = style
	return b
}

// SetIcons sets per-option icons displayed above the option label (card mode).
// When icons are set, GetFieldHeight returns 4 and each option renders as a
// tall card with the icon on the first inner row and the label on the second.
func (b *ButtonGroup) SetIcons(icons []string) *ButtonGroup {
	b.icons = icons
	return b
}

// GetFieldHeight returns 6 in card mode (icons set), 1 otherwise.
func (b *ButtonGroup) GetFieldHeight() int {
	if len(b.icons) > 0 {
		return 6
	}
	return 1
}

// Focus is called when this primitive receives focus.
func (b *ButtonGroup) Focus(delegate func(p Primitive)) {
	if b.disabled && b.finished != nil {
		b.finished(-1)
		return
	}
	b.Box.Focus(delegate)
}

// Blur is called when this primitive loses focus.
func (b *ButtonGroup) Blur() {
	// Reset cursor to the selected option so next focus starts there.
	b.focusedOption = b.currentOption
	b.Box.Blur()
}

// Draw renders the button group onto the screen.
func (b *ButtonGroup) Draw(screen tcell.Screen) {
	b.Box.DrawForSubclass(screen, b)

	x, y, width, height := b.GetInnerRect()
	if height < 1 || width <= 0 {
		return
	}
	rightLimit := x + width

	// Draw label.
	_, labelBg, _ := b.labelStyle.Decompose()
	if b.labelWidth > 0 {
		lw := b.labelWidth
		if lw > width {
			lw = width
		}
		printWithStyle(screen, b.label, x, y, 0, lw, AlignLeft, b.labelStyle, labelBg == tcell.ColorDefault)
		x += lw
	} else if b.label != "" {
		_, _, drawnWidth := printWithStyle(screen, b.label, x, y, 0, width, AlignLeft, b.labelStyle, labelBg == tcell.ColorDefault)
		x += drawnWidth
	}

	if len(b.icons) > 0 {
		b.drawCards(screen, x, y, rightLimit)
	} else {
		b.drawInline(screen, x, y, rightLimit)
	}
}

func (b *ButtonGroup) styleFor(i int) tcell.Style {
	switch {
	case b.disabled:
		return b.unselectedStyle.Background(b.backgroundColor)
	case b.HasFocus() && i == b.focusedOption:
		return b.cursorStyle
	case len(b.icons) == 0 && i == b.currentOption:
		// In flat mode the confirmed selection is visually distinct.
		// In card mode there is no persistent selection — only the cursor matters.
		return b.selectedStyle
	default:
		return b.unselectedStyle
	}
}

func (b *ButtonGroup) drawInline(screen tcell.Screen, x, y, rightLimit int) {
	for i, option := range b.options {
		text := " " + option + " "
		textWidth := len([]rune(text))
		if x >= rightLimit {
			break
		}
		style := b.styleFor(i)
		_, bg, _ := style.Decompose()
		printWithStyle(screen, text, x, y, 0, textWidth, AlignLeft, style, bg == tcell.ColorDefault)
		x += textWidth + 1
	}
}

func (b *ButtonGroup) drawCards(screen tcell.Screen, startX, y, rightLimit int) {
	// Uniform card width so all cards look identical.
	cardWidth := 14
	for i, option := range b.options {
		if w := TaggedStringWidth(option) + 6; w > cardWidth {
			cardWidth = w
		}
		if i < len(b.icons) {
			if w := TaggedStringWidth(b.icons[i]) + 6; w > cardWidth {
				cardWidth = w
			}
		}
	}

	const gap = 3
	availWidth := rightLimit - startX
	totalCardsWidth := len(b.options)*cardWidth + max(len(b.options)-1, 0)*gap

	// Center the block of cards within the available space.
	x := startX
	if totalCardsWidth < availWidth {
		x += (availWidth - totalCardsWidth) / 2
	}

	innerWidth := cardWidth - 2 // space between the two border bars
	topBorder := "╭" + strings.Repeat("─", innerWidth) + "╮"
	botBorder := "╰" + strings.Repeat("─", innerWidth) + "╯"
	blankRow := "│" + strings.Repeat(" ", innerWidth) + "│"

	for i, option := range b.options {
		if x+cardWidth > rightLimit {
			break
		}
		style := b.styleFor(i)
		_, bg, _ := style.Decompose()
		transparent := bg == tcell.ColorDefault

		icon := ""
		if i < len(b.icons) {
			icon = b.icons[i]
		}

		printWithStyle(screen, topBorder, x, y, 0, cardWidth, AlignLeft, style, transparent)
		printWithStyle(screen, blankRow, x, y+1, 0, cardWidth, AlignLeft, style, transparent)
		printWithStyle(screen, "│"+cardCenteredRow(icon, innerWidth)+"│", x, y+2, 0, cardWidth, AlignLeft, style, transparent)
		printWithStyle(screen, "│"+cardCenteredRow(option, innerWidth)+"│", x, y+3, 0, cardWidth, AlignLeft, style, transparent)
		printWithStyle(screen, blankRow, x, y+4, 0, cardWidth, AlignLeft, style, transparent)
		printWithStyle(screen, botBorder, x, y+5, 0, cardWidth, AlignLeft, style, transparent)

		x += cardWidth + gap
	}
}

// cardCenteredRow returns text padded with spaces so its display width equals width.
func cardCenteredRow(text string, width int) string {
	textW := TaggedStringWidth(text)
	if textW >= width {
		return text
	}
	left := (width - textW) / 2
	right := width - textW - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

// InputHandler returns the handler for this primitive.
func (b *ButtonGroup) InputHandler() func(event *tcell.EventKey, setFocus func(p Primitive)) {
	return b.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p Primitive)) {
		if b.disabled {
			return
		}

		finish := func(key tcell.Key) {
			if b.done != nil {
				b.done(key)
			}
			if b.finished != nil {
				b.finished(key)
			}
		}

		switch key := event.Key(); key {
		case tcell.KeyLeft:
			if b.focusedOption > 0 {
				b.focusedOption--
			}
		case tcell.KeyRight:
			if b.focusedOption < len(b.options)-1 {
				b.focusedOption++
			}
		case tcell.KeyRune:
			switch event.Rune() {
			case 'h':
				if b.focusedOption > 0 {
					b.focusedOption--
				}
			case 'l':
				if b.focusedOption < len(b.options)-1 {
					b.focusedOption++
				}
			case ' ':
				b.currentOption = b.focusedOption
				if b.changed != nil {
					b.changed(b.options[b.currentOption], b.currentOption)
				}
			}
		case tcell.KeyEnter:
			b.currentOption = b.focusedOption
			if b.changed != nil {
				b.changed(b.options[b.currentOption], b.currentOption)
			}
		case tcell.KeyTab:
			finish(tcell.KeyTab)
		case tcell.KeyBacktab:
			finish(tcell.KeyBacktab)
		case tcell.KeyEscape:
			finish(tcell.KeyEscape)
		case tcell.KeyUp:
			finish(tcell.KeyBacktab)
		case tcell.KeyDown:
			finish(tcell.KeyTab)
		}
	})
}
