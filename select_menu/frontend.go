//go:build js

package select_menu

import (
	"fmt"
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/go/gopherjs_http/jsutil"
	"honnef.co/go/js/dom"
)

func init() {
	js.Global.Set("SelectMenuOnInput", jsutil.Wrap(SelectMenuOnInput))
}

func SelectMenuOnInput(event dom.Event, selElem dom.HTMLElement, defaultOption, queryParameter string) {
	url, err := url.Parse(dom.GetWindow().Location().Href)
	if err != nil {
		// We don't expect this can ever happen, so treat it as an internal error if it does.
		panic(fmt.Errorf("internal error: parsing window.location.href as URL failed: %v", err))
	}
	query := url.Query()
	if selected := selElem.(*dom.HTMLSelectElement).SelectedOptions()[0].Text; selected == defaultOption {
		query.Del(queryParameter)
	} else {
		query.Set(queryParameter, selected)
	}
	url.RawQuery = query.Encode()
	dom.GetWindow().Location().Href = url.String()
}
