// +build !js

// Package checkbox provides a checkbox connected to a query parameter.
package checkbox

import (
	"fmt"
	"html/template"
	"net/url"
	"strconv"

	"github.com/shurcooL/htmlg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// New creates the HTML for a checkbox instance. Its checked value is directly connected
// to the presence of queryParameter.
// Changing either the presence of queryParameter, or checking/unchecking the checkbox
// will result in the other updating to match.
func New(defaultValue bool, query url.Values, queryParameter string) template.HTML {
	inputElement := &html.Node{
		Type: html.ElementNode,
		Data: "input",
		Attr: []html.Attribute{{Key: atom.Type.String(), Val: "checkbox"}},
	}

	var selectedValue = defaultValue
	if _, set := query[queryParameter]; set {
		selectedValue = !selectedValue
	}
	if selectedValue {
		inputElement.Attr = append(inputElement.Attr, html.Attribute{Key: "checked"})
	}

	inputElement.Attr = append(inputElement.Attr, html.Attribute{
		Key: "onchange",
		// HACK: Don't use Sprintf, properly encode (as json at this time).
		Val: fmt.Sprintf(`CheckboxOnChange(event, this, %v, %q);`, defaultValue, strconv.Quote(queryParameter)),
	})

	return template.HTML(htmlg.Render(inputElement))
}
