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

func findFroms(data []byte) (int, int) {
	curPos := 0
	for {
		fromPos := bytes.Index(data[curPos:], []byte("\nFrom "))
		if bytes.HasPrefix(data[curPos:], []byte("From ")) {
			fromPos = 0
		}
		if fromPos == -1 {
			return -1, -1
		}
		fromPos += curPos
		nextLine := bytes.IndexByte(data[fromPos+1:], '\n')
		if nextLine == -1 {
			return -1, -1
		}
		nextLine += fromPos + 1
		if data[nextLine-1] <= '9' && data[nextLine-1] >= '0' &&
			data[nextLine-2] <= '9' && data[nextLine-2] >= '0' &&
			data[nextLine-3] <= '9' && data[nextLine-3] >= '0' &&
			(data[nextLine-4] == '1' || data[nextLine-4] == '2') {
			return fromPos, nextLine + 1
		}
		curPos = nextLine
	}
}

// scanMessage is a split function for a bufio.Scanner that returns a message in
// RFC 822 format or an error.
func scanMessage(data []byte, atEOF bool) (int, []byte, error) {
	if len(data) == 0 && atEOF {
		return 0, nil, nil
	}
	start, end := findFroms(data)
	if start == -1 || end == -1 {
		if !atEOF {
			return 0, nil, nil
		}
		// log.Printf("invalid MBOX format, still had data to process as follows:\n*********start*******\n%q\n**********end********", data)
		return len(data), nil, nil
		//return 0, nil, ErrInvalidMboxFormat
	}
	curStart, curEnd := end, end
	for {
		priorStart, priorEnd := findFroms(data[curEnd:])
		//log.Printf("start=%d, end=%d,priorStart=%d,priorEnd=%d,data=%s", start, end, priorStart, priorEnd, data[curEnd:])
		if priorStart == -1 || priorEnd == -1 {
			if atEOF { // have the initial From header, just want to return what we have without finding the next one
				return len(data), data[end:], nil
			}
			return 0, nil, nil
		}
		curStart, curEnd = priorStart+curEnd, priorEnd+curEnd
		if bytes.Index(data[curEnd:], []byte("\n\n")) == -1 {
			// must be a blank after the headers before content
			return 0, nil, nil // get more, end of header hasn't yet come
		}
		tpr := textproto.NewReader(bufio.NewReader(bytes.NewReader(data[curEnd:])))
		header, err := tpr.ReadMIMEHeader()
		if err != nil || len(header) < 2 {
			// error processing header, probably not a valid message!  move on!
			continue
		}
		//if len(header) >= 2 { // found my next proper From!
		return curStart - 1, data[end:curStart], nil
	}
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
