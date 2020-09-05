// Package reactionsmenu provides a reactions menu component.
package reactionsmenu

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"github.com/shurcooL/htmlg"
	"github.com/shurcooL/reactions"
	reactionscomponent "github.com/shurcooL/reactions/component"
	"github.com/shurcooL/users"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

var rm reactionsMenu

type reactionsMenu struct {
	reactableURI      string
	reactionsService  reactions.Service
	authenticatedUser users.User

	menu    *dom.HTMLDivElement
	filter  *dom.HTMLInputElement
	results *dom.HTMLDivElement

	filtered []string // Filtered emojiIDs from reactions.Sorted list. With colons, e.g., ":+1:".

	// From last Show, needed to rerender reactableContainer after toggling a reaction.
	reactableID        string
	reactableContainer dom.Element
}

// Setup sets up the reaction menu on the current page.
// It must be called exactly once when document.Body() already exists.
func Setup(reactableURI string, reactionsService reactions.Service, authenticatedUser users.User) {
	rm = reactionsMenu{
		reactableURI:      reactableURI,
		reactionsService:  reactionsService,
		authenticatedUser: authenticatedUser,
	}

	js.Global.Set("ShowReactionMenu", jsutil.Wrap(rm.Show))
	js.Global.Set("ToggleReaction", jsutil.Wrap(rm.ToggleReaction))

	rm.menu = document.CreateElement("div").(*dom.HTMLDivElement)
	rm.menu.SetID("rm-reactions-menu")
	rm.menu.Style().SetProperty("display", "none", "")

	container := document.CreateElement("div").(*dom.HTMLDivElement)
	container.SetClass("rm-reactions-menu-container")
	rm.menu.AppendChild(container)

	// Disable for unauthenticated user.
	if rm.authenticatedUser.ID == 0 {
		disabled := document.CreateElement("div").(*dom.HTMLDivElement)
		disabled.SetClass("rm-reactions-menu-disabled")
		signIn := document.CreateElement("div").(*dom.HTMLDivElement)
		signIn.SetClass("rm-reactions-menu-signin")
		returnURL := dom.GetWindow().Location().Pathname + dom.GetWindow().Location().Search
		signIn.SetInnerHTML(signInHTML(returnURL) + " to react.")
		disabled.AppendChild(signIn)
		container.AppendChild(disabled)
	}

	rm.filter = document.CreateElement("input").(*dom.HTMLInputElement)
	rm.filter.SetClass("rm-reactions-filter")
	rm.filter.Placeholder = "Search"
	rm.menu.AddEventListener("click", false, func(event dom.Event) {
		if rm.authenticatedUser.ID != 0 {
			rm.filter.Focus()
		}
	})
	container.AppendChild(rm.filter)
	rm.results = document.CreateElement("div").(*dom.HTMLDivElement)
	rm.results.SetClass("rm-reactions-results")
	rm.results.AddEventListener("click", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		x := (me.ClientX - int(rm.results.GetBoundingClientRect().Left) + rm.results.Underlying().Get("scrollLeft").Int()) / 30
		if x >= 9 {
			return // Out of bounds to the right, likely because of scrollbar.
		}
		y := (me.ClientY - int(rm.results.GetBoundingClientRect().Top) + rm.results.Underlying().Get("scrollTop").Int()) / 30
		i := y*9 + x
		if i < 0 || i >= len(rm.filtered) {
			return
		}
		reaction := reactions.EmojiID(strings.Trim(rm.filtered[i], ":")) // Trim colons because reactions.Sorted has them, but reactions.EmojiID shouldn't.
		go func() {
			reactions, err := rm.reactionsService.Toggle(context.TODO(), rm.reactableURI, rm.reactableID, reactions.ToggleRequest{Reaction: reaction})
			if err != nil {
				log.Println(err)
				return
			}
			inner := reactionscomponent.ReactionsBarInner{
				Reactions:   reactions,
				CurrentUser: rm.authenticatedUser,
				ReactableID: rm.reactableID,
			}
			rm.reactableContainer.SetInnerHTML(htmlg.Render(inner.Render()...))
		}()
		rm.hide()
	})
	container.AppendChild(rm.results)
	preview := document.CreateElement("div").(*dom.HTMLDivElement)
	preview.SetClass("rm-reactions-preview")
	preview.SetInnerHTML(`<span id="rm-reactions-preview-emoji"><span class="rm-emoji rm-large"></span></span><span id="rm-reactions-preview-label"></span>`)
	container.AppendChild(preview)

	rm.updateFilteredResults()
	rm.filter.AddEventListener("input", false, func(dom.Event) {
		rm.updateFilteredResults()
	})

	rm.results.AddEventListener("mousemove", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		x := (me.ClientX - int(rm.results.GetBoundingClientRect().Left) + rm.results.Underlying().Get("scrollLeft").Int()) / 30
		if x >= 9 {
			return // Out of bounds to the right, likely because of scrollbar.
		}
		y := (me.ClientY - int(rm.results.GetBoundingClientRect().Top) + rm.results.Underlying().Get("scrollTop").Int()) / 30
		i := y*9 + x
		rm.updateSelected(i)
	})

	document.Body().AddEventListener("keydown", false, func(event dom.Event) {
		if event.DefaultPrevented() {
			return
		}

		ke, ok := event.(*dom.KeyboardEvent)
		if !ok {
			// Modern browsers send a custom event with "keydown" type when an auto-complete entry
			// is picked, so we may get a *dom.BasicEvent{Type: "keydown"} or *dom.CustomEvent{...}
			// rather than *dom.KeyboardEvent. Ignore such events.
			return
		}
		switch {
		// Escape.
		case ke.KeyCode == 27 && !ke.Repeat && !ke.CtrlKey && !ke.AltKey && !ke.MetaKey && !ke.ShiftKey:
			if rm.isHidden() {
				return
			}
			rm.hide()
			ke.PreventDefault()
		}
	})

	backgroundClick := func(event dom.Event) {
		if event.DefaultPrevented() {
			return
		}
		if rm.menu.Contains(event.Target()) {
			return
		}
		if rm.isHidden() {
			return
		}
		rm.hide()
		event.PreventDefault()
	}
	// Use capture because we want click-outside-to-close to get handled before others.
	document.AddEventListener("click", true, backgroundClick)
	addTapEventListener(document, true, backgroundClick)

	document.Body().AppendChild(rm.menu)
}

// TODO: Come up with API that will allow providing a custom signInHTML implementation.
func signInHTML(returnURL string) string {
	const style = `a.Login { color: #4183c4; text-decoration: none; } a.Login:hover { text-decoration: underline; }`
	u := url.URL{Path: "/login", RawQuery: url.Values{"return": {returnURL}}.Encode()}
	signInViaURL := &html.Node{
		Type: html.ElementNode, Data: atom.A.String(),
		Attr: []html.Attribute{
			{Key: atom.Class.String(), Val: "Login"},
			{Key: atom.Href.String(), Val: u.String()},
		},
		FirstChild: htmlg.Text("Sign in via URL"),
	}
	return "<style>" + style + "</style>" + htmlg.Render(signInViaURL)
}

// addTapEventListener adds a virtual tap event listener to et.
// A tap is considered to occur when the first touch point starts and ends in same spot,
// without moving or canceling or other touches starting.
func addTapEventListener(et dom.EventTarget, useCapture bool, listener func(dom.Event)) {
	var tapArmed bool
	et.AddEventListener("touchstart", useCapture, func(evt dom.Event) {
		firstTouch := len(evt.(*dom.TouchEvent).Touches()) == 1
		tapArmed = firstTouch // Tap is armed only when the first touch starts.
	})
	et.AddEventListener("touchmove", useCapture, func(dom.Event) { tapArmed = false })
	et.AddEventListener("touchcancel", useCapture, func(dom.Event) { tapArmed = false })
	et.AddEventListener("touchend", useCapture, func(evt dom.Event) {
		if !tapArmed {
			return
		}
		listener(evt)
	})
}

// Show shows the reactions menu.
func (rm *reactionsMenu) Show(this dom.HTMLElement, event dom.Event, reactableID string) {
	if event.DefaultPrevented() {
		return
	}

	rm.reactableID = reactableID
	rm.reactableContainer = document.GetElementByID("reactable-container-" + reactableID)

	rm.filter.Value = ""
	rm.filter.Underlying().Call("dispatchEvent", js.Global.Get("CustomEvent").New("input")) // Trigger "input" event listeners.
	rm.updateSelected(0)

	rm.menu.Style().SetProperty("display", "initial", "")

	// rm.menu aims to have 270px client width. Due to optional scrollbars
	// taking up some of that space, we may need to compensate and increase width.
	if scrollbarWidth := rm.results.OffsetWidth() - rm.results.Get("clientWidth").Float(); scrollbarWidth > 0 {
		rm.menu.Style().SetProperty("width", fmt.Sprintf("%fpx", 270+scrollbarWidth+1), "")
	}

	rm.results.Set("scrollTop", 0)
	top := float64(dom.GetWindow().ScrollY()) + this.GetBoundingClientRect().Top - rm.menu.GetBoundingClientRect().Height - 10
	if minTop := float64(dom.GetWindow().ScrollY()) + 12; top < minTop {
		top = minTop
	}
	rm.menu.Style().SetProperty("top", fmt.Sprintf("%vpx", top), "")
	left := float64(dom.GetWindow().ScrollX()) + this.GetBoundingClientRect().Left
	if maxLeft := float64(dom.GetWindow().InnerWidth()+dom.GetWindow().ScrollX()) - rm.menu.GetBoundingClientRect().Width - 12; left > maxLeft {
		left = maxLeft
	}
	if minLeft := float64(dom.GetWindow().ScrollX()) + 12; left < minLeft {
		left = minLeft
	}
	rm.menu.Style().SetProperty("left", fmt.Sprintf("%vpx", left), "")
	if rm.authenticatedUser.ID != 0 {
		rm.filter.Focus()
	}

	event.PreventDefault()
}

func (rm *reactionsMenu) hide() {
	rm.menu.Style().SetProperty("display", "none", "")
}

func (rm *reactionsMenu) isHidden() bool {
	return rm.menu.Style().GetPropertyValue("display") == "none"
}

// ToggleReaction toggles reaction. reaction is withot colons, e.g., "+1".
func (rm *reactionsMenu) ToggleReaction(this dom.HTMLElement, event dom.Event, reaction reactions.EmojiID) {
	container := getAncestorByClassName(this, "reactable-container")
	reactableID := container.GetAttribute("data-reactableID")

	if rm.authenticatedUser.ID == 0 {
		rm.Show(this, event, reactableID)
		return
	}

	go func() {
		reactions, err := rm.reactionsService.Toggle(context.TODO(), rm.reactableURI, reactableID, reactions.ToggleRequest{Reaction: reaction})
		if err != nil {
			log.Println(err)
			return
		}
		inner := reactionscomponent.ReactionsBarInner{
			Reactions:   reactions,
			CurrentUser: rm.authenticatedUser,
			ReactableID: reactableID,
		}
		container.SetInnerHTML(htmlg.Render(inner.Render()...))
	}()
}

func (rm *reactionsMenu) updateFilteredResults() {
	lower := strings.ToLower(strings.TrimSpace(rm.filter.Value))
	rm.results.SetInnerHTML("")
	rm.filtered = nil
	for _, emojiID := range reactions.Sorted {
		if lower != "" && !strings.Contains(emojiID, lower) {
			continue
		}
		element := document.CreateElement("div")
		rm.results.AppendChild(element)
		element.SetOuterHTML(`<div class="rm-reaction"><span class="rm-emoji" style="background-position: ` + reactions.Position(emojiID) + `;"></span></div>`)
		rm.filtered = append(rm.filtered, emojiID)
	}
}

// updateSelected updates selected reaction to rm.filtered[index].
func (rm *reactionsMenu) updateSelected(index int) {
	if index < 0 || index >= len(rm.filtered) {
		return
	}
	emojiID := rm.filtered[index]

	label := document.GetElementByID("rm-reactions-preview-label").(*dom.HTMLSpanElement)
	label.SetTextContent(strings.Trim(emojiID, ":"))
	emoji := document.GetElementByID("rm-reactions-preview-emoji").(*dom.HTMLSpanElement)
	emoji.FirstChild().(dom.HTMLElement).Style().SetProperty("background-position", reactions.Position(emojiID), "")
}

func getAncestorByClassName(el dom.Element, class string) dom.Element {
	for ; el != nil && !el.Class().Contains(class); el = el.ParentElement() {
	}
	return el
}
