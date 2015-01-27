// +build js

package select_menu

import (
	"net/url"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"honnef.co/go/js/dom"
)

func init() {
	js.Global.Set("SelectMenuOnInput", jsutil.Wrap(SelectMenuOnInput))
}

func SelectMenuOnInput(event dom.Event, object dom.HTMLElement, defaultOption, queryParameter string) {
	rawQuery := strings.TrimPrefix(dom.GetWindow().Location().Search, "?")
	query, _ := url.ParseQuery(rawQuery)

	selectElement := object.(*dom.HTMLSelectElement)

	/*selectedIndex := selectElement.Underlying().Get("selectedIndex").Int()
	selected := selectElement.Options()[selectedIndex].Text*/
	selected := selectElement.Underlying().Get("selectedOptions").Index(0).Get("text").String()

	if selected == defaultOption {
		query.Del(queryParameter)
	} else {
		query.Set(queryParameter, selected)
	}

	var originalSelected int
	options := selectElement.Underlying().Get("options")
	for i := 0; i < options.Length(); i++ {
		if options.Index(i).Get("defaultSelected").Bool() {
			originalSelected = i
			break
		}
	}
	selectElement.Underlying().Set("selectedIndex", originalSelected)

	dom.GetWindow().Location().Search = "?" + query.Encode()
}
