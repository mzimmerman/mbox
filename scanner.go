// "THE BEER-WARE LICENSE" (Revision 42):
// <tobias.rehbein@web.de> wrote this file. As long as you retain this notice
// you can do whatever you want with this stuff. If we meet some day, and you
// think this stuff is worth it, you can buy me a beer in return.
//                                                             Tobias Rehbein

package mbox

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"
)

// ErrInvalidMboxFormat is the error returned by the Next method of type Mbox if
// its content is malformed in a way that it is not possible to extract a
// message.
var ErrInvalidMboxFormat = errors.New("invalid mbox format")

// scanHeader is a split function for a bufio.Scanner that returns a messages headers in
// RFC 822 format or an error.
func scanHeader(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 && atEOF {
		return 0, nil, nil
	}
	e := bytes.Index(data, []byte("\n\n\n"))
	if e == -1 && !atEOF {
		// request more data
		return 0, nil, nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return e + 3, data[:e+3], nil
}

var fromLine = regexp.MustCompile("\nFrom .*(19|20)[0-9]{2}")

// scanMessage is a split function for a bufio.Scanner that returns a message in
// RFC 822 format or an error.
func scanMessage(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 && atEOF {
		return 0, nil, nil
	}

	var n int
	e := -1
	r := fromLine.FindIndex(data)
	if len(r) > 0 {
		e = r[0]
	}
	//e := bytes.Index(data, []byte("\nFrom "))
	advanceExtra := 0
	if e == 0 {
		data = data[1:] // advance past the leading LF
		advanceExtra = 1
		r = fromLine.FindIndex(data)
		if len(r) > 0 {
			e = r[0]
		} else {
			e = -1
		}
		//e = bytes.Index(data, []byte("\nFrom "))
	}
	if e == -1 && !atEOF {
		// request more data
		return advanceExtra, nil, nil
	}

	if !bytes.HasPrefix(data, []byte("From ")) {
		return advanceExtra, nil, ErrInvalidMboxFormat
	}
	n = bytes.IndexByte(data, '\n')
	if n == -1 {
		return advanceExtra, nil, ErrInvalidMboxFormat
	}

	if atEOF {
		if data[len(data)-1] != '\n' {
			return advanceExtra, nil, ErrInvalidMboxFormat
		}
		return len(data) + advanceExtra, data[n+1:], nil
	}
	tpr := textproto.NewReader(bufio.NewReader(bytes.NewReader(data[n+1:])))
	header, err := tpr.ReadMIMEHeader()
	if err != nil {
		return 0, nil, err
	}
	cth := header.Get("Content-Type")
	boundaryEnd := ""
	if strings.Contains(cth, "multipart") {
		splt := strings.Split(cth, "; ")
		for _, v := range splt {
			if strings.HasPrefix(v, "boundary=") {
				c := strings.Index(v, "=") + 1
				boundaryEnd = "--" + strings.Trim(strings.TrimRight(v[c:], ";"), `"'`) + "--"
				break
			}
		}
	}
	if boundaryEnd != "" {
		b := bytes.Index(data, []byte(boundaryEnd))
		if b == -1 {
			c := bytes.Index(data, []byte("\nContent-Type: multipart"))
			// c == the first boundary
			d := bytes.Index(data[c+1:], []byte("\nContent-Type: multipart"))
			if d+c+1 > e { // assume that the boundary end was never received
				return e + 1 + advanceExtra, data[n+1 : e], nil
			}
			return 0, nil, nil // need more data!
		}
		if e < b {
			r = fromLine.FindIndex(data)
			if len(r) > 0 {
				e = r[0]
			} else {
				e = -1
			}
			//e = bytes.Index(data[b:], []byte("\nFrom "))
			e += b
		}
	}
	if data[e-1] != '\n' {
		return e + 1 + advanceExtra, data[n+1 : e+1], nil
	}
	return e + 1 + advanceExtra, data[n+1 : e], nil
}

// Scanner provides an interface to read a sequence of messages from an mbox.
// Calling the Next method steps through the messages. The current message can
// then be accessed by calling the Message method.
//
// The Next method returns true while there are messages to skip to and no error
// occurs. When Next returns false, you can call the Err method to check for an
// error.
//
// The Message method returns the current message as *mail.Message, or nil if an
// error occured while calling Next or if you have skipped past the last message
// using Next. If Next returned true, you can expect Message to return a valid
// *mail.Message.
type Scanner struct {
	s       *bufio.Scanner
	m       *mail.Message
	curByte int
	err     error
}

// NewScanner returns a new *Scanner to read messages from mbox file format data
// provided by io.Reader r.
func NewScanner(r io.Reader, headers bool) *Scanner {
	s := bufio.NewScanner(r)
	if headers {
		s.Split(scanHeader)
	} else {
		s.Split(scanMessage)
	}
	return &Scanner{s: s}
}

func (m *Scanner) Location() int {
	return m.curByte
}

// Next skips to the next message and returns true. It will return false if
// there are no messages left or an error occurs. You can call the Err method to
// check if an error occured. If Next returns false and Err returns nil there
// are no messages left.
func (m *Scanner) Next() bool {
	m.m = nil
	if m.err != nil {
		return false
	}

	if !m.s.Scan() {
		m.err = m.s.Err()
		return false
	}
	m.curByte += len(m.s.Bytes())
	m.m, m.err = mail.ReadMessage(bytes.NewReader(m.s.Bytes()))
	if m.err != nil {
		return false
	}
	return true
}

// Err returns the first error that occured while calling Next.
func (m *Scanner) Err() error {
	return m.err
}

// Message returns the current message. It returns nil if you never called Next,
// skipped past the last message or if an error occured during a call to Next.
//
// If Next returned true, you can expect Message to return a valid
// *mail.Message.
func (m *Scanner) Message() *mail.Message {
	if m.err != nil {
		return nil
	}
	return m.m
}

// Buffer sets the initial buffer to use when scanning and the maximum size of
// buffer that may be allocated during scanning.
//
// Buffer panics if it is called after scanning has started.
func (m *Scanner) Buffer(buf []byte, max int) {
	m.s.Buffer(buf, max)
}
