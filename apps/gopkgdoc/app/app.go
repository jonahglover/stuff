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
	var d struct {
		Path, LinkID, ASIN string
	}
	d.Path = r.URL.Path
	switch {
	case strings.HasPrefix(d.Path, "/pkg/"):
		d.Path = d.Path[len("/pkg"):]
	case d.Path == "/index":
		d.Path = "/-/index"
	}
	switch {
	case strings.HasPrefix(d.Path, "/labix.org/v2/mgo"):
		d.ASIN = "1449344682"
		d.LinkID = "Y4ZCMCCFWTSWNGDK"
	case strings.HasPrefix(d.Path, "/launchpad.net/gozk/zookeeper"):
		d.ASIN = "1449361307"
		d.LinkID = "V7MS2ZDNWPUE37MG"
	case strings.HasPrefix(d.Path, "/github.com/jmcvetta/facebook"):
		d.ASIN = "1449367615"
		d.LinkID = "O5HAY7GRH5BWBRUU"
	default:
		d.ASIN = "1934356980"
		d.LinkID = "HMMJZDJBRDHD4O3X"
	}
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, &d)
}

func init() {
	http.HandleFunc("/", handler)
}

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{"hasPrefix": strings.HasPrefix}).Parse(html))
var html = `
<!DOCTYPE html>
<html>
<head>
<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body>
<iframe style="width:120px;height:240px;" marginwidth="0" marginheight="0" scrolling="no" frameborder="0" src="//ws-na.amazon-adsystem.com/widgets/q?ServiceVersion=20070822&OneJS=1&Operation=GetAdHtml&MarketPlace=US&source=ss&ref=ss_til&ad_type=product_link&tracking_id=605030-20&marketplace=amazon&region=US&placement={{.ASIN}}&asins={{.ASIN}}&linkId=Y4ZCMCCFWTSWNGDK&show_border=false&link_opens_in_new_window=false"> </iframe>
<p>The page that you are looking for moved to <a href="http://godoc.org{{.Path}}">http://godoc.org{{.Path}}</a>.
</body>
</html>
`
