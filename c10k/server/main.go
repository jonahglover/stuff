// Copyright 2011 Gary Burd
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

package main

import (
	"flag"
	"github.com/garyburd/twister/expvar"
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

var (
	msg             = []byte("hello")
	connectionCount = expvar.NewInt("connections")
	messageCount    = expvar.NewInt("messges")
)

func testHandler(req *web.Request) {
	connectionCount.Add(1)
	defer connectionCount.Add(-1)
	w := req.Respond(web.StatusOK, web.HeaderContentType, web.ContentTypeHTML)
	for {
		<-time.After(time.Duration(10e9 + 1e9*rand.Int63n(10)))
		_, err := w.Write(msg)
		if err != nil {
			return
		}
		if w, ok := w.(web.Flusher); ok {
			err = w.Flush()
			if err != nil {
				return
			}
		}
		messageCount.Add(1)
	}
}

func gcHandler(req *web.Request) {
	t := time.Now()
	runtime.GC()
	d := time.Now().Sub(t)
	req.Respond(web.StatusOK, web.HeaderContentType, "text/plain").Write([]byte(strconv.FormatInt(int64(d), 10)))
}

func main() {
	flag.Parse()
	h := web.NewRouter().
		Register("/test", "GET", testHandler).
		Register("/gc", "GET", gcHandler).
		Register("/", "GET", expvar.ServeWeb)
	server.Run(":8080", h)
}
