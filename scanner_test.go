package mbox

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"
)

const mboxWithOneMessage = `From herp.derp at example.com  Thu Jan  1 00:00:01 2015
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxWithOneMessageMissingHeaders = `From herp.derp at example.com  Thu Jan  1 00:00:01 2015
This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxWithThreeMessages = `From herp.derp at example.com  Thu Jan  1 00:00:01 2015
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.

From derp.herp at example.com  Thu Jan  1 00:00:01 2015
From: derp.herp at example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Another line.

Bye.

From bernd.lauert at example.com  Thu Jan  3 00:00:01 2015
From: bernd.lauert at example.com (Bernd Lauert)
Date: Thu, 03 Jan 2015 00:00:01 +0100
Subject: A last test

This is the last simple test.

Bye.
`

const mboxWithStartingLF = `
From herp.derp at example.com  Thu Jan  1 00:00:01 2015
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.

From derp.herp at example.com  Thu Jan  1 00:00:01 2015
From: derp.herp at example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Another line.

Bye.

From bernd.lauert at example.com  Thu Jan  3 00:00:01 2015
From: bernd.lauert at example.com (Bernd Lauert)
Date: Thu, 03 Jan 2015 00:00:01 +0100
Subject: A last test

This is the last simple test.

Bye.
`

const mboxWithThreeMessagesMalformedButValid = `From herp.derp at example.com  Thu Jan  1 00:00:01 2015
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
From derp.herp at example.com  Thu Jan  1 00:00:01 2015
From: derp.herp at example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Another line.

Bye.

From bernd.lauert at example.com  Thu Jan  3 00:00:01 2015
From: bernd.lauert at example.com (Bernd Lauert)
Date: Thu, 03 Jan 2015 00:00:01 +0100
Subject: A last test

This is the last simple test.

Bye.
`

const mboxWithOneMessageMissingSeparator = `From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxFirstMessage = `From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxFirstMessageBody = `This is a simple test.

And, by the way, this is how a "From" line is escaped in mboxo format:

>From Herp Derp with love.

Bye.
`

const mboxSecondMessageSubjectHeader = "Another test"

type tsmInput struct {
	data  string
	atEOF bool
}

type tsmExpected struct {
	advance     int
	token       string
	yieldsError bool
}

func testScanMessage(t *testing.T, input *tsmInput, expected *tsmExpected) {
	advance, token, err := scanMessage([]byte(input.data), input.atEOF)

	if err == nil && expected.yieldsError {
		t.Errorf("unexpected success")
	}
	if err != nil && !expected.yieldsError {
		t.Errorf("unexpected error: %v", err)
	}
	if advance != expected.advance {
		t.Errorf("unexpected advance: %d", advance)
	}
	if string(token) != expected.token {
		t.Errorf("unexpected token: %q", token)
	}
}

func TestScanMessageMboxEmptyAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  "",
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithOneMessageAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessage,
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     281,
		token:       mboxFirstMessage,
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithOneMessageMissingSeparatorAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessageMissingSeparator,
	}

	expected := &tsmExpected{
		yieldsError: true,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithThreeMessages(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  mboxWithThreeMessages,
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     282,
		token:       mboxFirstMessage,
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageWithThreeMessagesMalformedButValid(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  mboxWithThreeMessagesMalformedButValid,
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     281,
		token:       mboxFirstMessage,
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageIncompleteRecord(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  mboxWithOneMessage[:100],
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageVeryShortIncompleteRecord(t *testing.T) {
	input := &tsmInput{
		atEOF: false,
		data:  "From",
	}

	expected := &tsmExpected{
		yieldsError: false,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageOnlySeperatorAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessage[:55],
	}

	expected := &tsmExpected{
		yieldsError: true,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func TestScanMessageMboxWithOneMessageWithoutNewlineAtEOF(t *testing.T) {
	input := &tsmInput{
		atEOF: true,
		data:  mboxWithOneMessage[:len(mboxWithOneMessage)-1],
	}

	expected := &tsmExpected{
		yieldsError: true,
		advance:     0,
		token:       "",
	}

	testScanMessage(t, input, expected)
}

func testMboxMessage(t *testing.T, mbox string, count int) {
	b := bytes.NewBufferString(mbox)
	m := NewScanner(b, false)

	for i := 0; i < count; i++ {
		if !m.Next() {
			t.Errorf("Next() failed; pass %d", i)
		}
		if m.Err() != nil {
			t.Errorf("Unexpected error after Next(): %v", m.Err())
		}

		msg := m.Message()
		if msg == nil {
			t.Errorf("message is nil; pass %d", i)
			continue
		}
		body := new(bytes.Buffer)
		_, err := body.ReadFrom(msg.Body)
		if err != nil {
			t.Errorf("Unexpected error reading message body: %v", err)
		}
		if i == 0 && body.String() != mboxFirstMessageBody {
			t.Errorf("Expected:\n %q\ngot\n%q", mboxFirstMessageBody, body.String())
		}
		if i == 1 && msg.Header.Get("Subject") != mboxSecondMessageSubjectHeader {
			t.Errorf("Unexpected subject header: %q", msg.Header.Get("Subject"))
		}
		if m.Err() != nil {
			t.Errorf("Unexpected error after Message(): %v", m.Err())
		}
	}

	if m.Next() {
		t.Errorf("Next() succeeded")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Next(): %v", m.Err())
	}
	if msg := m.Message(); msg != nil {
		t.Errorf("message is not nil")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Message(): %v", m.Err())
	}
}

func TestMboxMessageWithOneMessage(t *testing.T) {
	testMboxMessage(t, mboxWithOneMessage, 1)
}

func TestMboxMessageWithThreeMessages(t *testing.T) {
	testMboxMessage(t, mboxWithThreeMessages, 3)
}

func TestMboxMessageWithStartingLF(t *testing.T) {
	testMboxMessage(t, mboxWithStartingLF, 3)
}

func TestMboxMessageWithThreeMessagesMalformedButValid(t *testing.T) {
	testMboxMessage(t, mboxWithThreeMessagesMalformedButValid, 3)
}

func testMboxMessageInvalid(t *testing.T, mbox string) {
	b := bytes.NewBufferString(mbox)
	m := NewScanner(b, false)

	if m.Next() {
		t.Errorf("Next() succeeded")
	}
	if m.Err() == nil {
		t.Errorf("Missing error after Next(): %v", m.Err())
	}
	if msg := m.Message(); msg != nil {
		t.Errorf("message is not nil")
	}
	if m.Err() == nil {
		t.Errorf("Missing error after Message(): %v", m.Err())
	}
	if m.Next() {
		t.Errorf("Next() after error succeeded")
	}
}

func TestMboxMessageWithOneMessageMissingSeparator(t *testing.T) {
	testMboxMessageInvalid(t, mboxWithOneMessageMissingSeparator)
}

func TestMboxMessageWithOneMessageMissingHeaders(t *testing.T) {
	testMboxMessageInvalid(t, mboxWithOneMessageMissingHeaders)
}

func TestScanMessageWithBoundaries(t *testing.T) {
	sourceData := `
From one place.  2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test
Content-Type: multipart/alternative;
        boundary=Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0

This is a test of boundaries.  Don't accept a new email via \nFrom until the boundary is done!'

And, by the way, this is how a "From" line is escaped in mboxo format:
From Herp Derp with love.

From Herp Derp with love.

Bye.
--Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0--

From another!  2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is the second email in a test of boundaries.
`
	expected := []string{
		"This is a test of boundaries.  Don't accept a new email via \\nFrom until the boundary is done!'\n\nAnd, by the way, this is how a \"From\" line is escaped in mboxo format:\nFrom Herp Derp with love.\n\nFrom Herp Derp with love.\n\nBye.\n--Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0--\n",
		"This is the second email in a test of boundaries.\n",
	}
	b := bytes.NewBufferString(sourceData)
	m := NewScanner(b, false)

	for i := range expected {
		if !m.Next() {
			t.Errorf("Next() failed; pass %d", i)
		}
		if m.Err() != nil {
			t.Errorf("Unexpected error after Next(): %v", m.Err())
		}

		msg := m.Message()
		if msg == nil {
			t.Errorf("message is nil; pass %d", i)
			continue
		}
		body := new(bytes.Buffer)
		_, err := body.ReadFrom(msg.Body)
		if err != nil {
			t.Errorf("%d - Unexpected error reading message body: %v", i, err)
			continue
		}
		if body.String() != expected[i] {
			t.Errorf("%d - Expected:\n %q\ngot\n%q", i, expected[i], body.String())
		}
		if m.Err() != nil {
			t.Errorf("%d - Unexpected error after Message(): %v", i, m.Err())
		}
	}

	if m.Next() {
		t.Errorf("Next() succeeded")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Next(): %v", m.Err())
	}
	if msg := m.Message(); msg != nil {
		t.Errorf("message is not nil")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Message(): %v", m.Err())
	}
}

func TestScanMessageWithOpenBoundaries(t *testing.T) {
	sourceData := `
From one place.  2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test
Content-Type: multipart/alternative;
        boundary=Apple-Mail-D55D9B1A-A379-4D5C-BDA9-00D35DF424A0

This is a test of boundaries.  Accept new boundaries if a new multipart Content-Type is found

From two place.  2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test
Content-Type: multipart/alternative;
        boundary=newboundary

From Herp Derp with love two.
--newboundary--

From another! 2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is the third email in a test of boundaries.
`
	expected := []string{
		"This is a test of boundaries.  Accept new boundaries if a new multipart Content-Type is found\n",
		"From Herp Derp with love two.\n--newboundary--\n",
		"This is the third email in a test of boundaries.\n",
	}
	b := bytes.NewBufferString(sourceData)
	m := NewScanner(b, false)

	for i := range expected {
		if !m.Next() {
			t.Errorf("Next() failed; pass %d", i)
		}
		if m.Err() != nil {
			t.Errorf("Unexpected error after Next(): %v", m.Err())
		}

		msg := m.Message()
		if msg == nil {
			t.Errorf("message is nil; pass %d", i)
			continue
		}
		body := new(bytes.Buffer)
		_, err := body.ReadFrom(msg.Body)
		if err != nil {
			t.Errorf("%d - Unexpected error reading message body: %v", i, err)
			continue
		}
		if body.String() != expected[i] {
			t.Errorf("%d - Expected:\n %q\ngot\n%q", i, expected[i], body.String())
		}
		if m.Err() != nil {
			t.Errorf("%d - Unexpected error after Message(): %v", i, m.Err())
		}
	}

	if m.Next() {
		t.Errorf("Next() succeeded")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Next(): %v", m.Err())
	}
	if msg := m.Message(); msg != nil {
		t.Errorf("message is not nil")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Message(): %v", m.Err())
	}
}

func TestScanMessageWithTextBoundary(t *testing.T) {
	sourceData := `
From one place.  2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test
Content-Type: text/html; charset="utf-8";
 boundary="monkey_d3df4dc8-da5e-47dd-be15-f19c5ed55194"

This is a test of boundaries.  Don't accept a new email via \nFrom until the boundary is done!'

And, by the way, this is how a "From" line is escaped in mboxo format:

Bye.

From another!  2014
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is the second email in a test of boundaries.
`
	expected := []string{
		"This is a test of boundaries.  Don't accept a new email via \\nFrom until the boundary is done!'\n\nAnd, by the way, this is how a \"From\" line is escaped in mboxo format:\n\nBye.\n",
		"This is the second email in a test of boundaries.\n",
	}
	b := bytes.NewBufferString(sourceData)
	m := NewScanner(b, false)
	for i := range expected {
		if !m.Next() {
			t.Errorf("Next() failed; pass %d", i)
		}
		if m.Err() != nil {
			t.Errorf("Unexpected error after Next(): %v", m.Err())
		}

		msg := m.Message()
		if msg == nil {
			t.Errorf("message is nil; pass %d", i)
			continue
		}
		body := new(bytes.Buffer)
		_, err := body.ReadFrom(msg.Body)
		if err != nil {
			t.Errorf("%d - Unexpected error reading message body: %v", i, err)
			continue
		}
		if body.String() != expected[i] {
			t.Errorf("%d - Expected:\n %q\ngot\n%q", i, expected[i], body.String())
		}
		if m.Err() != nil {
			t.Errorf("%d - Unexpected error after Message(): %v", i, m.Err())
		}
	}
	if m.Next() {
		t.Errorf("Next() succeeded")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Next(): %v", m.Err())
	}
	if msg := m.Message(); msg != nil {
		t.Errorf("message is not nil")
	}
	if m.Err() != nil {
		t.Errorf("Unexpected error after Message(): %v", m.Err())
	}
}

func TestHeaders(t *testing.T) {
	tests := []struct {
		name          string
		expectedFound int
		expectedError error
		buffer        string
	}{
		{
			name:          "one message",
			expectedFound: 1,
			expectedError: nil,
			buffer: `Delivered-To: test@host.com
From: test0@host.com
To: test@host.com
Date: 14 Oct 2013 09:08:42 +0200
Message-ID: <messageid-is-unique@host.com>


`,
		},
		{
			name:          "four messages",
			expectedFound: 4,
			expectedError: nil,
			buffer: `Delivered-To: test@host.com
From: test0@host.com
To: test@host.com
Date: 14 Oct 2013 09:08:42 +0200
Message-ID: <messageid-is-unique@host.com>


Delivered-To: test@host.com
From: test1@host.com
To: test@host.com
Date: 14 Oct 2013 09:08:42 +0200
Message-ID: <messageid-is-unique@host.com>


Delivered-To: test@host.com
From: test2@host.com
To: test@host.com
Date: 14 Oct 2013 09:08:42 +0200
Message-ID: <messageid-is-unique@host.com>


Delivered-To: test@host.com
From: test3@host.com
To: test@host.com
Date: 14 Oct 2013 09:08:42 +0200
Message-ID: <messageid-is-unique@host.com>


`,
		},
	}
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
	for i, test := range tests {
		b := strings.NewReader(test.buffer)
		m := NewScanner(b, true)
		for j := 0; j < test.expectedFound; j++ {
			next := m.Next()
			err := m.Err()
			if !next {
				if err != test.expectedError {
					t.Errorf("%s - Expected error %v, got %v", test.name, test.expectedError, err)
				} else {
					t.Errorf("%s - Next, on %d, returned false before it should!", test.name, j)
				}
			}
			if msg := m.Message(); msg == nil {
				t.Errorf("message is nil; pass %d", i)
			} else {
				fromHdr := msg.Header["From"]
				expected := fmt.Sprintf("test%d@host.com", j)
				if len(fromHdr) < 1 {
					t.Errorf("%s-%d Expected from address of %s, got %v", test.name, j, expected, fromHdr)
				} else if got := fromHdr[0]; got != expected {
					t.Errorf("%s-%d Expected from address of %s, got %s", test.name, j, expected, got)
				}
			}
		}
		if m.Next() {
			t.Errorf("%s - Next() succeeded", test.name)
		}
		if m.Err() != nil {
			t.Errorf("%s - Unexpected error after Next(): %v", test.name, m.Err())
		}
		if msg := m.Message(); msg != nil {
			t.Errorf("%s - message is not nil, got %#v", test.name, msg)
		}
		if m.Err() != nil {
			t.Errorf("%s - Unexpected error after Message(): %v", test.name, m.Err())
		}
	}
}

func ExampleScanner() {
	r := strings.NewReader(`From herp.derp at example.com  Thu Jan  1 00:00:01 2015
From: herp.derp at example.com (Herp Derp)
Date: Thu, 01 Jan 2015 00:00:01 +0100
Subject: Test

This is a simple test.

CU.

From derp.herp at example.com  Thu Jan  1 00:00:01 2015
From: derp.herp at example.com (Derp Herp)
Date: Thu, 02 Jan 2015 00:00:01 +0100
Subject: Another test

This is another simple test.

Bye.
`)

	mbox := NewScanner(r, false)
	for mbox.Next() {
		// If Next() returns true, you can expect Message() to return a
		// valid *mail.Message.
		fmt.Printf("Message from %v\n", mbox.Message().Header.Get("from"))
	}
	if mbox.Err() != nil {
		fmt.Print("Oops, something went wrong!")
	}
	// Output:
	// Message from herp.derp at example.com (Herp Derp)
	// Message from derp.herp at example.com (Derp Herp)
}
