package prompt_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ymgyt/cli/prompt"
)

func TestSession_Prompt(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		in := bytes.NewBufferString("yes\n")
		out := &custom{buff: make([]byte, 0)}
		msg := "are you sure?[yes/no]"
		p := prompt.New().SetReader(in).SetWriter(out).Display(msg).Require("yes")
		ok, err := p.Prompt()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("want ok, but got not ok")
		}
		got, want := out.String(), msg
		if got != want {
			t.Errorf("displayed prompt message does not match. got %q, want %q", got, want)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		in := bytes.NewBufferString("YES\n")
		msg := "type yes!"
		p := prompt.New().SetReader(in).SetWriter(ioutil.Discard).Display(msg).Require("yes").CaseInsensitive()
		ok, err := p.Prompt()
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("want ok, but got not ok")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		in := &custom{}
		p := prompt.New().SetReader(in).SetWriter(ioutil.Discard).Require("yes").Timeout(100 * time.Millisecond)
		_, err := p.Prompt()
		if err != prompt.ErrTimeout {
			t.Fatalf("want timeout error, but got %v", err)
		}
	})

	t.Run("respect context", func(t *testing.T) {
		in := &custom{}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		p := prompt.New().SetReader(in).SetWriter(ioutil.Discard).Context(ctx).Timeout(1 * time.Second)
		_, err := p.Prompt()
		if err != context.DeadlineExceeded {
			t.Fatalf("want context.DeadlineExceeded, but got %v", err)
		}
	})
}

type custom struct {
	i    int
	buff []byte
}

func (c *custom) Write(p []byte) (int, error) {
	c.buff = append(c.buff, p[0])
	c.i++
	return 1, nil
}

func (c *custom) Read(p []byte) (int, error) {
	select {}
	return 0, errors.New("not sleeping") // nolint
}

func (c *custom) String() string { return string(c.buff) }

