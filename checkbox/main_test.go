// +build browsertest

package checkbox_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shurcooL/frontend/checkbox"
	"github.com/shurcooL/go/gopherjs_http"
	"github.com/shurcooL/go/open"
	"github.com/shurcooL/httpfs/httputil"
)

func Test(t *testing.T) {
	http.Handle("/script.js", httputil.FileHandler{gopherjs_http.Package("github.com/shurcooL/frontend/checkbox")})

	{
		defaultValue := false
		queryParameter := "some-optional-thing"

		http.HandleFunc("/index.html", func(w http.ResponseWriter, req *http.Request) {
			query := req.URL.Query()

			checkboxHtml := checkbox.New(defaultValue, query, queryParameter)

			io.WriteString(w, `<html><head><script type="text/javascript" src="/script.js"></script></head><body>`+string(checkboxHtml)+"</body></html>")
		})
	}

	ts := httptest.NewServer(nil)
	defer ts.Close()

	open.Open(ts.URL + "/index.html")

	select {}
}
