// Copyright 2014 Gary Burd
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
    "net/http"
    "regexp"
    "log"
    "bufio"
)

var pats = []*regexp.Regexp{
    regexp.MustCompile(`(^/.*)`),
    regexp.MustCompile(`^\d+/\d+/\d+ \d+:\d+:\d+ go-app-builder: (.*)`),
    regexp.MustCompile(`^\d+/\d+/\d+ \d+:\d+:\d+ (.*)`),
}

func main() {
    log.SetFlags(0)
    r, err := http.Get("http://localhost:8080/")
    if err != nil {
        log.Fatal(err)
    }
    defer r.Body.Close()
    s := bufio.NewScanner(r.Body)
    for s.Scan() {
        for _, pat := range pats {
            if m := pat.FindSubmatch(s.Bytes()); m != nil {
                log.Printf("%s", m[1])
                break
            }
        }
    }
}
