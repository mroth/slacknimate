package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestLineScanner(t *testing.T) {
	letters := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	alphabet := strings.Join(letters, "\n")

	t.Run("normal", func(t *testing.T) {
		r := strings.NewReader(alphabet)
		s := NewLineScanner(context.Background(), r)

		// verify each frame
		//
		// channel will close upon completion, if not the range will never
		// conclude and the test will fail via timeout.
		var i int
		for frame := range s.Frames() {
			if want := letters[i]; frame != want {
				t.Errorf("frame %d: want %v got %v", i, want, frame)
			}
			i++
		}

		// check err
		if err := s.Err(); err != nil {
			t.Errorf("Err(): want %v got %v", nil, err)
		}
	})

	t.Run("context expired", func(t *testing.T) {
		ctx, cf := context.WithCancel(context.Background())
		r := strings.NewReader(alphabet)
		s := NewLineScanner(ctx, r)

		// read to roughly the halfway point
		frames := s.Frames()
		for i := 0; i <= len(letters)/2; i++ {
			_, ok := <-frames
			if !ok {
				t.Fatal("channel closed prematurely")
			}
		}
		// give the scanner a bit so we know its next value was queued for the
		// outbound channel, but don't consume it yet
		<-time.After(time.Millisecond)
		// oh no! someone just cancelled our context!
		cf()
		// one remaining produced value to be drained
		if _, ok := <-s.Frames(); !ok {
			t.Fatal("channel closed before drained")
		}
		// and now the channel should be closed
		if _, ok := <-s.Frames(); ok {
			t.Fatal("expected closed channel")
		}
	})
}

func TestLoopingLineScanner(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())

		// create new LLS scanning a source with multiple lines
		chunks := []string{
			"Mary had a litte lamb.",
			"It's fleece was white as snow.",
			"And everywhere that Mary went,",
			"The lamb was sure to go.",
		}
		r := strings.NewReader(strings.Join(chunks, "\n"))
		s := NewLoopingLineScanner(ctx, r, len(chunks)*2)

		// read for 5 full iterations, channel should not be closed
		frames := s.Frames()
		for i := 0; i < 5*len(chunks); i++ {
			got, ok := <-frames
			if !ok {
				t.Fatalf("channel closed before expected on iteration %d", i)
			}
			want := chunks[i%len(chunks)]
			if want != got {
				t.Errorf("unexpected frame contents: want %v got %v", want, got)
			}
		}

		// after which, Err() should still be nil
		if err := s.Err(); err != nil {
			t.Errorf("unexpected Err(): %v", err)
		}

		// cancel the context, make sure channel got closed
		cancelFunc()
		_, ok := <-frames
		if ok {
			t.Fatal("channel not closed after context cancelled")
		}

		// scanner should have received the context cancellation as err
		wantErr := context.Canceled
		if gotErr := s.Err(); gotErr != wantErr {
			t.Errorf("Err(): want %v got %v", wantErr, gotErr)
		}
	})

	t.Run("buffer size exceeded", func(t *testing.T) {
		tenXs := "x\nx\nx\nx\nx\nx\nx\nx\nx\nx"
		r := strings.NewReader(tenXs)
		s := NewLoopingLineScanner(context.Background(), r, 8)
		// frames channel should be closed before output begins
		_, ok := <-s.Frames()
		if ok {
			t.Error("channel not closed after buffer size exceeded")
		}
		wantErr := ErrMaxFramesExceeded
		if gotErr := s.Err(); gotErr != wantErr {
			t.Errorf("Err(): want %v got %v", wantErr, gotErr)
		}
	})
}
