package handler

import (
	"go/build"
	"log"
	"net/http"
	"path/filepath"

	"github.com/shurcooL/go/gopherjs_http"
	"github.com/shurcooL/httpfs/httputil"
	"github.com/shurcooL/httpfs/vfsutil"
)

func init() {
	// HACK: This code registers routes at root on default mux... That's not very nice.
	http.Handle("/table-of-contents.js", httputil.FileHandler{File: gopherjs_http.Package("github.com/shurcooL/frontend/table-of-contents")})
	http.Handle("/table-of-contents.css", httputil.FileHandler{File: vfsutil.File(filepath.Join(importPathToDir("github.com/shurcooL/frontend/table-of-contents"), "style.css"))})
}

func importPathToDir(importPath string) string {
	p, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	return p.Dir
}
