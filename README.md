frontend
========

[![Build Status](https://travis-ci.org/shurcooL/frontend.svg?branch=master)](https://travis-ci.org/shurcooL/frontend) [![GoDoc](https://godoc.org/github.com/shurcooL/frontend?status.svg)](https://godoc.org/github.com/shurcooL/frontend)

Common frontend code.

Installation
------------

```bash
go get -u github.com/shurcooL/frontend/...
GOARCH=js go get -u -d github.com/shurcooL/frontend/...
```

Testing Locally
---------------

Note: `gopherjs_serve_html` is superceded by the official `gopherjs serve` command. These instructions should be updated to use that instead.

For packages that have any `_test.html` files, you can use [`gopherjs_serve_html`](http://godoc.org/github.com/shurcooL/cmd/gopherjs_serve_html) to serve said test. For example:

```bash
cd ./table-of-contents/
gopherjs_serve_html main_test.html    # Serves main_html.html at http://localhost:8080/index.html.
open http://localhost:8080/index.html # Open http://localhost:8080/index.html in browser.
```

Changes to .go code are reloaded on every request, so you can make changes, refresh browser to see new version. Watch browser console for errors.

Directories
-----------

| Path                                                                                                  | Synopsis                                                                          |
|-------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------|
| [checkbox](https://godoc.org/github.com/shurcooL/frontend/checkbox)                                   | Package checkbox provides a checkbox connected to a query parameter.              |
| [reactionsmenu](https://godoc.org/github.com/shurcooL/frontend/reactionsmenu)                         | Package reactionsmenu provides a reactions menu component.                        |
| [select_menu](https://godoc.org/github.com/shurcooL/frontend/select_menu)                             |                                                                                   |
| [table-of-contents/handler](https://godoc.org/github.com/shurcooL/frontend/table-of-contents/handler) |                                                                                   |
| [tabsupport](https://godoc.org/github.com/shurcooL/frontend/tabsupport)                               | Package tabsupport offers functionality to add tab support to a textarea element. |

License
-------

-	[MIT License](LICENSE)
