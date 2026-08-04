// Package tview implements rich widgets for terminal based user interfaces.
// This file centralises the global default for built-in vim-style navigation.

package tview

// DefaultVimKeys decides whether widgets handle built-in vim-style rune
// navigation (j/k/h/l/g/G and similar) in their input handlers. Individual
// widgets can override it with SetVimKeys. It is consulted at input-handling
// time, so changing it affects already-created widgets.
var DefaultVimKeys = true

// SetDefaultVimKeys sets the package-wide default for built-in vim-style
// navigation. Widgets without an explicit SetVimKeys call use this value.
func SetDefaultVimKeys(v bool) {
	DefaultVimKeys = v
}
