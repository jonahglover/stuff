// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// +build appengine

package app

import (
	"html/template"
	"net/http"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case strings.HasPrefix(p, "/pkg/"):
		p = p[len("/pkg"):]
	case p == "/index":
		p = "/-/index"
	}
	tmpl.Execute(w, p)
}

func init() {
	http.HandleFunc("/", handler)
}

var tmpl = template.Must(template.New("").Parse(html))
var html = `
<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
The page that you are looking for moved to <a href="http://godoc.org/{{.}}">http://godoc.org{{.}}</a>.
</body>
</html>
`
