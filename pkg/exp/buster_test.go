// Copyright 2013 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tango

import (
	"net/http"
	"testing"
)

func TestCacheBusters(t *testing.T) {
	cbs := CacheBusters{Handler: http.FileServer(http.Dir("."))}

	token := cbs.get("/buster_test.go")
	if token == "" {
		t.Errorf("could not extract token from http.FileServer")
	}

	var ss StaticServer
	cbs = CacheBusters{Handler: ss.FileHandler("buster_test.go")}

	token = cbs.get("/xxx")
	if token == "" {
		t.Errorf("could not extract token from StaticServer FileHandler")
	}
}
