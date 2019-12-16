// +build js

// Package tabsupport offers functionality to add tab support to a textarea element.
package tabsupport

import (
	"syscall/js"

	"honnef.co/go/js/dom/v2"
)

// Add is a helper that modifies a <textarea>, so that pressing tab key will insert tabs.
func Add(textArea *dom.HTMLTextAreaElement) {
	textArea.AddEventListener("keydown", false, func(event dom.Event) {
		switch ke := event.(*dom.KeyboardEvent); {
		case ke.KeyCode() == '\t' && !ke.CtrlKey() && !ke.AltKey() && !ke.MetaKey() && !ke.ShiftKey(): // Tab.
			event.PreventDefault()
			insertTab(textArea)
		}
	})
}

// KeyDownHandler is a keydown event handler for a <textarea> element.
// It makes it so that pressing tab key will insert tabs.
//
// To use it, first make it available to the JavaScript world, e.g.:
//
// 	js.Global().Set("TabSupportKeyDownHandler", jsutil.Wrap(tabsupport.KeyDownHandler))
//
// Then use it as follows in the HTML:
//
// 	<textarea onkeydown="TabSupportKeyDownHandler(this, event);"></textarea>
//
func KeyDownHandler(element dom.HTMLElement, event dom.Event) {
	switch ke := event.(*dom.KeyboardEvent); {
	case ke.KeyCode() == '\t' && !ke.CtrlKey() && !ke.AltKey() && !ke.MetaKey() && !ke.ShiftKey(): // Tab.
		event.PreventDefault()
		insertTab(element.(*dom.HTMLTextAreaElement))
	}
}

func insertTab(textArea *dom.HTMLTextAreaElement) {
	value, start, end := textArea.Value(), textArea.SelectionStart(), textArea.SelectionEnd()

	textArea.SetValue(value[:start] + "\t" + value[end:])
	textArea.SetSelectionStart(start + 1)
	textArea.SetSelectionEnd(start + 1)

	// Trigger "input" event listeners.
	inputEvent := js.Global().Get("CustomEvent").New("input")
	textArea.Underlying().Call("dispatchEvent", inputEvent)
}
