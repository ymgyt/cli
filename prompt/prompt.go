package prompt

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

const (
	defaultBuffer = 256
)

var (
	defaultTimeout = time.Minute * 30
	ErrTimeout     = errors.New("timeout")
)

func New() *Session {
	return &Session{}
}

type Session struct {
	ctx             context.Context
	in              io.Reader
	out             io.Writer
	displayMessage  string
	requires        []string
	caseInsensitive bool
	timeout         time.Duration
	bufferBytes     int
	err             error
}

func (s *Session) Display(msg string) *Session {
	s.displayMessage = msg
	return s
}

func (s *Session) Require(words ...string) *Session {
	s.requires = words
	return s
}

func (s *Session) CaseInsensitive() *Session {
	s.caseInsensitive = true
	return s
}

func (s *Session) CaseSensitive() *Session {
	s.caseInsensitive = false
	return s
}

func (s *Session) Timeout(d time.Duration) *Session {
	s.timeout = d
	return s
}

func (s *Session) Context(ctx context.Context) *Session {
	s.ctx = ctx
	return s
}

func (s *Session) SetReader(r io.Reader) *Session {
	s.in = r
	return s
}

func (s *Session) SetWriter(w io.Writer) *Session {
	s.out = w
	return s
}

func (s *Session) SetBufferBytes(n int) *Session {
	s.bufferBytes = n
	return s
}

func (s *Session) init() {
	if s.ctx == nil {
		s.ctx = context.Background()
	}
	if s.in == nil {
		s.in = os.Stdin
	}
	if s.out == nil {
		s.out = os.Stdout
	}
	if s.bufferBytes == 0 {
		s.bufferBytes = defaultBuffer
	}
	if s.timeout == 0 {
		s.timeout = defaultTimeout
	}
}

func (s *Session) Prompt() (ok bool, err error) {
	s.init()
	if s.err != nil {
		return false, err
	}

	ch := make(chan result, 1)
	go func() { ch <- s.prompt() }()

	var r result
	select {
	case r = <-ch:
	case <-time.After(s.timeout):
		r = result{err: ErrTimeout}
	case <-s.ctx.Done():
		r = result{err: s.ctx.Err()}
	}
	return r.ok, r.err
}

type result struct {
	ok  bool
	err error
}

func (s *Session) prompt() result {
	fail := func(err error) result { return result{err: err} }

	// write display message
	msg := []byte(s.displayMessage)
	wn := 0
	for {
		n, err := s.out.Write(msg[wn:])
		if err != nil {
			return fail(err)
		}
		wn += n
		if wn >= len(msg) {
			break
		}
	}

	// read user input
	var word string
	buff := make([]byte, s.bufferBytes)
	for {
		n, err := s.in.Read(buff)
		if err == io.EOF {
			break
		} else if err != nil {
			return fail(err)
		}
		word += string(buff[:n])
		if idx := strings.Index(word, "\n"); idx > 0 {
			break
		}
	}
	return s.handleWord(word)
}

func (s *Session) handleWord(word string) result {
	got := strings.TrimSpace(word)
	if s.caseInsensitive {
		got = strings.ToLower(got)
	}
	for _, want := range s.requires {
		if s.caseInsensitive {
			want = strings.ToLower(want)
		}
		if strings.Compare(got, want) == 0 {
			return result{ok: true}
		}
	}
	return result{ok: false}
}

