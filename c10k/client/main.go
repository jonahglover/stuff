package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"github.com/garyburd/twister/expvar"
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
)

var (
	responseLineRegexp = regexp.MustCompile("^HTTP/[0-9.]+ ([0-9]+) ")
	connectionCount    = expvar.NewInt("connections")
	readCount          = expvar.NewInt("reads")
)

func dial(urlString string) (net.Conn, *bufio.Reader, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, nil, err
	}

	addr := u.Host
	if strings.LastIndex(addr, ":") <= strings.LastIndex(addr, "]") {
		if u.Scheme == "http" {
			addr = addr + ":80"
		} else {
			addr = addr + ":443"
		}
	}

	header := web.NewHeader(web.HeaderHost, u.Host)

	var request bytes.Buffer
	request.WriteString("GET ")
	request.WriteString(u.RawPath)
	request.WriteString(" HTTP/1.1\r\n")
	header.WriteHttpHeader(&request)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	if _, err := conn.Write(request.Bytes()); err != nil {
		conn.Close()
		return nil, nil, err
	}

	r, _ := bufio.NewReaderSize(conn, 512)
	p, err := r.ReadSlice('\n')
	if err != nil {
		conn.Close()
		return nil, nil, err
	}

	m := responseLineRegexp.FindSubmatch(p)
	if m == nil {
		conn.Close()
		return nil, nil, errors.New("bad response")
	}

	for {
		p, err = r.ReadSlice('\n')
		if err != nil {
			conn.Close()
			return nil, nil, err
		}
		if len(p) <= 2 {
			break
		}
	}

	if string(m[1]) != "200" {
		p, _ := ioutil.ReadAll(r)
		log.Println(string(p))
		conn.Close()
		return nil, nil, errors.New("bad response")
	}
	return conn, r, nil
}

func run() {
	conn, r, err := dial(*serverURL)
	if err != nil {
		log.Println("dial error", err)
		return
	}
	connectionCount.Add(1)
	defer connectionCount.Add(-1)
	defer conn.Close()
	b := make([]byte, 256)
	for {
		_, err := r.Read(b)
		if err != nil {
			log.Println("read error", err)
			return
		}
		readCount.Add(1)
	}
}

var (
	serverURL = flag.String("server", "http://localhost:8080/test", "")
	count     = flag.Int("count", 10, "")
)

func main() {
	flag.Parse()
	for i := 0; i < *count; i++ {
		go run()
	}
	h := web.HandlerFunc(expvar.ServeWeb)
	server.Run("127.0.0.1:8081", h)
}
