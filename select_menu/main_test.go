//go:build browsertest

package select_menu_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shurcooL/frontend/select_menu"
	"github.com/shurcooL/go/gopherjs_http"
	"github.com/shurcooL/go/open"
	"github.com/shurcooL/httpfs/httputil"
)

func Test(t *testing.T) {
	http.Handle("/script.js", httputil.FileHandler{gopherjs_http.Package("github.com/shurcooL/frontend/select_menu")})

	{
		options := []string{"option one", "option two (default)", "option three", "option four", "option five"}
		defaultOption := "option two (default)"
		queryParameter := "parameter"

		http.HandleFunc("/index.html", func(w http.ResponseWriter, req *http.Request) {
			query := req.URL.Query()

			selectMenuHtml := select_menu.New(options, defaultOption, query, queryParameter)

			io.WriteString(w, `<html><head><script type="text/javascript" src="/script.js"></script></head><body>`+string(selectMenuHtml)+"</body></html>")
		})
	}

	ts := httptest.NewServer(nil)
	defer ts.Close()

	open.Open(ts.URL + "/index.html")

	select {}
}
