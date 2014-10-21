// Copyright 2013 Gary Burd. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tango

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type busterWriter struct {
	headerMap http.Header
	status    int
	io.Writer
}

func (bw *busterWriter) Header() http.Header {
	return bw.headerMap
}

func (bw *busterWriter) WriteHeader(status int) {
	bw.status = status
}

// CacheBusters maintains a cache of cache-busting tokens for static resources
// served by Handler.
//
// Tokens are computed by issuing a HEAD request to Handler for the static
// resource. If the response has an ETag header, then the ETag header is used
// as the token. Otherwise, the Last-Modified header is used.
type CacheBusters struct {
	Handler http.Handler

	mu     sync.Mutex
	tokens map[string]string
}

func sanitizeTokenRune(r rune) rune {
	if r <= ' ' || r >= 127 {
		return -1
	}
	// Convert percent encoding reserved characters to '-'.
	if strings.ContainsRune("!#$&'()*+,/:;=?@[]", r) {
		return '-'
	}
	return r
}

func (cb *CacheBusters) get(path string) string {
	cb.mu.Lock()
	if cb.tokens == nil {
		cb.tokens = make(map[string]string)
	}
	token, ok := cb.tokens[path]
	cb.mu.Unlock()
	if ok {
		return token
	}

	w := busterWriter{
		Writer:    ioutil.Discard,
		headerMap: make(http.Header),
	}
	r := &http.Request{URL: &url.URL{Path: path}, Method: "HEAD"}
	cb.Handler.ServeHTTP(&w, r)

	if w.status == 200 {
		token = w.headerMap.Get("Etag")
		if token == "" {
			token = w.headerMap.Get("Last-Modified")
		}
		token = strings.Trim(token, `" `)
		token = strings.Map(sanitizeTokenRune, token)
	}

	cb.mu.Lock()
	cb.tokens[path] = token
	cb.mu.Unlock()

	return token
}

// AppendQueryParam returns path with the cache-busting token for path added as
// a query parameter.
func (cb *CacheBusters) AppendQueryParam(path string, name string) string {
	token := cb.get(path)
	if token == "" {
		return path
	}
	return path + "?" + name + "=" + token
}
