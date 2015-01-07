// Copyright 2015 Gary Burd
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

package resp

import (
	"fmt"
	"io"
	"strconv"
	"sync"
)

const blockSize = 4096

type block struct {
	next *block
	buf  [blockSize]byte
}

var blockPool = sync.Pool{New: func() interface{} { return new(block) }}

type CommandBuffer struct {
	// Scratch space for formatting argument length.
	// '*' or '$', length, "\r\n"
	lenScratch [32]byte

	// Scratch space for formatting integers and floats.
	numScratch [40]byte

	// linked list of blocks
	head, tail *block

	// current position tail block
	i int
}

func (cb *CommandBuffer) WriteTo(w io.Writer) error {
	var err error
	for b := cb.head; b != cb.tail; {
		if err != nil {
			_, err = w.Write(b.buf[:])
		}
		b = b.next
		blockPool.Put(b)
	}
	if cb.tail != nil {
		if err != nil {
			_, err = w.Write(cb.tail.buf[:cb.i])
		}
		blockPool.Put(cb.tail)
	}
	cb.head, cb.tail, cb.i = nil, nil, 0
	return err
}

func (cb *CommandBuffer) appendRawBytes(p []byte) {
	for len(p) > 0 {
		if cb.i == blockSize {
			b := blockPool.Get().(*block)
			cb.tail.next = b
			cb.tail = b
			cb.i = 0
		}
		n := copy(cb.tail.buf[cb.i:], p)
		cb.i += n
		p = p[n:]
	}
}

func (cb *CommandBuffer) appendRawString(p string) {
	for len(p) > 0 {
		if cb.i == blockSize {
			b := blockPool.Get().(*block)
			cb.tail.next = b
			cb.tail = b
			cb.i = 0
		}
		n := copy(cb.tail.buf[cb.i:], p)
		cb.i += n
		p = p[n:]
	}
}

func (cb *CommandBuffer) appendLen(prefix byte, n int) {
	cb.lenScratch[len(cb.lenScratch)-1] = '\n'
	cb.lenScratch[len(cb.lenScratch)-2] = '\r'
	i := len(cb.lenScratch) - 3
	for {
		cb.lenScratch[i] = byte('0' + n%10)
		i -= 1
		n = n / 10
		if n == 0 {
			break
		}
	}
	cb.lenScratch[i] = prefix
	cb.appendRawBytes(cb.lenScratch[i:])
}

func (cb *CommandBuffer) appendBytes(p []byte) {
	cb.appendLen('$', len(p))
	cb.appendRawBytes(p)
	cb.appendRawString("\r\n")
}

func (cb *CommandBuffer) appendString(s string) {
	cb.appendLen('$', len(s))
	cb.appendRawString(s)
	cb.appendRawString("\r\n")
}

func (cb *CommandBuffer) appendInt64(n int64) {
	cb.appendBytes(strconv.AppendInt(cb.numScratch[:0], n, 10))
}

func (cb *CommandBuffer) appendFloat64(n float64) {
	cb.appendBytes(strconv.AppendFloat(cb.numScratch[:0], n, 'g', -1, 64))
}

func (cb *CommandBuffer) Add(commandName string, args []interface{}) error {
	if cb.head == nil {
		cb.head = blockPool.Get().(*block)
		cb.tail = cb.head
		cb.i = 0
	}

	head, tail, i := cb.head, cb.tail, cb.i

	n := 1
	for _, arg := range args {
		switch arg := arg.(type) {
		case []string:
			n += len(arg)
		case []interface{}:
			n += len(arg)
		case [][]byte:
			n += len(arg)
		case []float64:
			n += len(arg)
		default:
			n += 1
		}
	}

	cb.appendLen('*', n)
	cb.appendString(commandName)
	for _, x := range args {
		switch x := x.(type) {
		case string:
			cb.appendString(x)
		case []byte:
			cb.appendBytes(x)
		case int:
			cb.appendInt64(int64(x))
		case int64:
			cb.appendInt64(x)
		case float64:
			cb.appendFloat64(x)
		case []string:
			for _, x := range x {
				cb.appendString(x)
			}
		case [][]byte:
			for _, x := range x {
				cb.appendBytes(x)
			}
		case []float64:
			for _, x := range x {
				cb.appendFloat64(x)
			}
		case []interface{}:
			for _, x := range x {
				switch x := x.(type) {
				case string:
					cb.appendString(x)
				case []byte:
					cb.appendBytes(x)
				case int:
					cb.appendInt64(int64(x))
				case int64:
					cb.appendInt64(x)
				case float64:
					cb.appendFloat64(x)
				default:
					cb.head, cb.tail, cb.i = head, tail, i
					return fmt.Errorf("resp: argument type %T not supported", x)
				}
			}
		default:
			cb.head, cb.tail, cb.i = head, tail, i
			return fmt.Errorf("resp: argument type %T not supported", x)
		}
	}
	return nil
}
