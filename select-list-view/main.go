//go:build js

package main

import (
	"html"
	"strings"

	"github.com/gopherjs/gopherjs/js"

	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

var headers []dom.Element

var selected int

var baseHash string
var baseX, baseY int

var entryHeight float64
var entries []dom.Node
var manuallyPicked string

func main() {}

func init() {
	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) {
		setup()
	})
}

func setup() {
	overlay := document.CreateElement("div").(*dom.HTMLDivElement)
	overlay.SetID("gts-overlay")

	container := document.CreateElement("div")
	overlay.AppendChild(container)
	container.SetOuterHTML(`<div><input id="gts-command"></input><div id="gts-results"></div></div>`)

	document.Body().AppendChild(overlay)

	command := document.GetElementByID("gts-command").(*dom.HTMLInputElement)
	results := document.GetElementByID("gts-results").(*dom.HTMLDivElement)

	command.AddEventListener("input", false, func(event dom.Event) {
		updateResults(false, nil)
	})

	/*mousedown := false
	results.AddEventListener("mousedown", false, func(event dom.Event) {
		mousedown = true

		command.Focus()

		me := event.(*dom.MouseEvent)
		y := (me.ClientY - results.GetBoundingClientRect().Top) + results.Underlying().Get("scrollTop").Int()
		selected = int(float64(y) / entryHeight)
		updateResultSelection()
	})
	results.AddEventListener("mouseup", false, func(event dom.Event) {
		mousedown = false
	})
	results.AddEventListener("mouseleave", false, func(event dom.Event) {
		mousedown = false
	})
	results.AddEventListener("mousemove", false, func(event dom.Event) {
		if !mousedown {
			return
		}

		command.Focus()

		me := event.(*dom.MouseEvent)
		y := (me.ClientY - results.GetBoundingClientRect().Top) + results.Underlying().Get("scrollTop").Int()
		selected = int(float64(y) / entryHeight)
		updateResultSelection()
	})*/
	results.AddEventListener("click", false, func(event dom.Event) {
		command.Focus()

		me, ok := event.(*dom.MouseEvent)
		if !ok {
			me = event.(*dom.PointerEvent).MouseEvent
		}
		y := (me.ClientY - int(results.GetBoundingClientRect().Top)) + results.Underlying().Get("scrollTop").Int()
		selected = int(float64(y) / entryHeight)
		updateResultSelection()
	})
	results.AddEventListener("dblclick", false, func(event dom.Event) {
		event.PreventDefault()

		hideOverlay(overlay)
	})

	overlay.AddEventListener("keydown", false, func(event dom.Event) {
		switch ke := event.(*dom.KeyboardEvent); {
		case ke.KeyCode == 27 && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Escape.
			ke.PreventDefault()

			if ke.Target().IsEqualNode(command) {
				js.Global.Get("window").Get("history").Call("replaceState", nil, nil, "#"+baseHash)
				dom.GetWindow().ScrollTo(baseX, baseY)
			}

			hideOverlay(overlay)
		case ke.KeyCode == 13 && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Enter.
			ke.PreventDefault()

			hideOverlay(overlay)
		case ke.KeyCode == 40 && !ke.CtrlKey && !ke.AltKey && ke.MetaKey && !ke.ShiftKey: // Down.
			ke.PreventDefault()
			selected = len(entries) - 1
			updateResultSelection()
		case ke.KeyCode == 40 && ke.CtrlKey && ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Down.
			ke.PreventDefault()
			results.Underlying().Set("scrollTop", results.Underlying().Get("scrollTop").Float()+entryHeight)
		case ke.KeyCode == 40 && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Down.
			ke.PreventDefault()
			selected++
			updateResultSelection()
		case ke.KeyCode == 38 && !ke.CtrlKey && !ke.AltKey && ke.MetaKey && !ke.ShiftKey: // Up.
			ke.PreventDefault()
			selected = 0
			updateResultSelection()
		case ke.KeyCode == 38 && ke.CtrlKey && ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Up.
			ke.PreventDefault()
			results.Underlying().Set("scrollTop", results.Underlying().Get("scrollTop").Float()-entryHeight)
		case ke.KeyCode == 38 && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Up.
			ke.PreventDefault()
			selected--
			updateResultSelection()
		}
	})

	document.Body().AddEventListener("keydown", false, func(event dom.Event) {
		switch ke := event.(*dom.KeyboardEvent); {
		case ke.KeyCode == int('F') && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // F.
			fallthrough
		case ke.KeyCode == int('R') && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Just R, since some browsers don't let us intercept Cmd+R.
			// Ignore just R when command elment has focus (it means the user is typing).
			if ke.Target().IsEqualNode(command) {
				break
			}
			fallthrough
		case ke.KeyCode == int('R') && !ke.CtrlKey && !ke.AltKey && ke.MetaKey && !ke.ShiftKey: // Cmd+R.
			ke.PreventDefault()

			// Is overlay already being displayed?
			if isOverlayVisible(overlay) {
				command.Select()
				break
			}

			command.Value = ""
			manuallyPicked = ""

			{
				headers = nil
				for _, header := range append(document.Body().GetElementsByTagName("h3"), document.Body().GetElementsByTagName("h4")...) {
					if header.ID() == "" {
						continue
					}
					headers = append(headers, header)
				}

				baseHash = strings.TrimPrefix(dom.GetWindow().Location().Hash, "#")
				baseX, baseY = dom.GetWindow().ScrollX(), dom.GetWindow().ScrollY()

				updateResults(true, overlay)
			}

			command.Select()
		case ke.KeyCode == 27 && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey: // Escape.
			// If overlay is not displayed, there's nothing to do.
			if !isOverlayVisible(overlay) {
				break
			}

			hideOverlay(overlay)

			ke.PreventDefault()
		}
	})
}

var previouslyHighlightedHeader dom.HTMLElement

func hideOverlay(overlay dom.HTMLElement) {
	overlay.Style().SetProperty("display", "none", "")

	if previouslyHighlightedHeader != nil {
		previouslyHighlightedHeader.Class().Remove("highlighted")
		previouslyHighlightedHeader.Class().Add("highlighted-fade")
	}

	document.GetElementByID("gts-command").(dom.HTMLElement).Blur() // Deselect the command input; needed in Firefox so body regains focus.
}

// isOverlayVisible returns true iff overlay is being displayed.
func isOverlayVisible(overlay dom.HTMLElement) bool {
	return (overlay.Style().GetPropertyValue("display") == "initial")
}

var previouslySelected int

func updateResultSelection() {
	results := document.GetElementByID("gts-results").(*dom.HTMLDivElement)

	if selected < 0 {
		selected = 0
	} else if selected > len(entries)-1 {
		selected = len(entries) - 1
	}

	if selected == previouslySelected {
		return
	}

	entries[previouslySelected].(dom.Element).Class().Remove("gts-highlighted")
	if previouslyHighlightedHeader != nil {
		previouslyHighlightedHeader.Class().Remove("highlighted")
	}

	{
		element := entries[selected].(dom.Element)

		if element.GetBoundingClientRect().Top < results.GetBoundingClientRect().Top {
			element.Underlying().Call("scrollIntoView", true)
		} else if element.GetBoundingClientRect().Bottom > results.GetBoundingClientRect().Bottom {
			element.Underlying().Call("scrollIntoView", false)
		}

		element.Class().Add("gts-highlighted")
		//dom.GetWindow().Location().Hash = "#" + element.GetAttribute("data-id")
		//dom.GetWindow().History().ReplaceState(nil, nil, "#"+element.GetAttribute("data-id"))
		js.Global.Get("window").Get("history").Call("replaceState", nil, nil, "#"+element.GetAttribute("data-id"))
		target := document.GetElementByID(element.GetAttribute("data-id")).(dom.HTMLElement)
		target.Class().Add("highlighted")
		previouslyHighlightedHeader = target
		centerOnTargetIfOffscreen(target)

		manuallyPicked = element.GetAttribute("data-id")
	}

	previouslySelected = selected
}

// offsetTopRoot returns the offset top of element e relative to root element.
func offsetTopRoot(e dom.HTMLElement) float64 {
	var offsetTopRoot float64
	for ; e != nil; e = e.OffsetParent() {
		offsetTopRoot += e.OffsetTop()
	}
	return offsetTopRoot
}

func centerOnTargetIfOffscreen(target dom.HTMLElement) {
	isOffscreen := int(offsetTopRoot(target)) < dom.GetWindow().ScrollY() ||
		int(offsetTopRoot(target)+target.OffsetHeight()) > dom.GetWindow().ScrollY()+dom.GetWindow().InnerHeight()

	if isOffscreen {
		windowHalfHeight := dom.GetWindow().InnerHeight() / 2

		dom.GetWindow().ScrollTo(dom.GetWindow().ScrollX(), int(offsetTopRoot(target)+target.OffsetHeight())-windowHalfHeight)
	}
}

var initialSelected int

func updateResults(init bool, overlay dom.HTMLElement) {
	windowHalfHeight := dom.GetWindow().InnerHeight() / 2
	filter := document.GetElementByID("gts-command").(*dom.HTMLInputElement).Value

	results := document.GetElementByID("gts-results").(*dom.HTMLDivElement)

	var selectionPreserved = false

	results.SetInnerHTML("")
	var visibleIndex int
	for _, header := range headers {
		if filter != "" && !strings.Contains(strings.ToLower(header.TextContent()), strings.ToLower(filter)) {
			continue
		}

		element := document.CreateElement("div")
		element.Class().Add("gts-entry")
		element.SetAttribute("data-id", header.ID())
		{
			entry := header.TextContent()
			index := strings.Index(strings.ToLower(entry), strings.ToLower(filter))
			element.SetInnerHTML(html.EscapeString(entry[:index]) + "<strong>" + html.EscapeString(entry[index:index+len(filter)]) + "</strong>" + html.EscapeString(entry[index+len(filter):]))
		}
		if header.ID() == manuallyPicked {
			selectionPreserved = true

			selected = visibleIndex
			previouslySelected = visibleIndex
		}

		results.AppendChild(element)

		visibleIndex++
	}

	entries = results.ChildNodes()

	if !selectionPreserved {
		manuallyPicked = ""

		if init {
			// Find the nearest entry.
			for i := len(entries) - 1; i >= 0; i-- {
				element := entries[i].(dom.Element)
				header := document.GetElementByID(element.GetAttribute("data-id"))

				if int(header.GetBoundingClientRect().Top) <= windowHalfHeight || i == 0 {
					selected = i
					previouslySelected = i

					initialSelected = i

					break
				}
			}
		} else {
			if filter == "" {
				selected = initialSelected
				previouslySelected = initialSelected
			} else {
				selected = 0
				previouslySelected = 0
			}
		}
	}

	if init {
		if previouslyHighlightedHeader != nil {
			previouslyHighlightedHeader.Class().Remove("highlighted-fade")
		}

		overlay.Style().SetProperty("display", "initial", "")
		entryHeight = results.FirstChild().(dom.Element).GetBoundingClientRect().Height
	} else {
		if previouslyHighlightedHeader != nil {
			previouslyHighlightedHeader.Class().Remove("highlighted")
		}
	}

	if len(entries) > 0 {
		element := entries[selected].(dom.Element)

		if init {
			y := float64(selected) * entryHeight
			results.Underlying().Set("scrollTop", y-float64(results.GetBoundingClientRect().Height/2))
		} else {
			if element.GetBoundingClientRect().Top <= results.GetBoundingClientRect().Top {
				element.Underlying().Call("scrollIntoView", true)
			} else if element.GetBoundingClientRect().Bottom >= results.GetBoundingClientRect().Bottom {
				element.Underlying().Call("scrollIntoView", false)
			}
		}

		element.Class().Add("gts-highlighted")
		//dom.GetWindow().Location().Hash = "#" + element.GetAttribute("data-id")
		//dom.GetWindow().History().ReplaceState(nil, nil, "#"+element.GetAttribute("data-id"))
		js.Global.Get("window").Get("history").Call("replaceState", nil, nil, "#"+element.GetAttribute("data-id"))
		target := document.GetElementByID(element.GetAttribute("data-id")).(dom.HTMLElement)
		target.Class().Add("highlighted")
		previouslyHighlightedHeader = target
		centerOnTargetIfOffscreen(target)
	}
}
