package main

import (
	"bufio"
	"context"
	"errors"
	"io"
)

// LineScanner will scan over an io.Reader line by line, broadcasting each line
// as a string over its Frames() channel. Once the io.Reader reaches EOF, the
// output channel will be closed.
type LineScanner struct {
	out chan string
	ctx context.Context
	err error
}

// Frames returns a channel which will broadcast a string with the contents of
// every line scanned from the underlying io.Reader.
func (s *LineScanner) Frames() <-chan string {
	return s.out
}

// Err returns the underlying error which was the cause of the LineScanner
// closing its Frames channel. If the reason was the underlying io.Reader
// encountered io.EOF, then Err will be nil.
func (s *LineScanner) Err() error {
	return s.err
}

// NewLineScanner starts and returns a new LineScanner for a given io.Reader.
func NewLineScanner(ctx context.Context, in io.Reader) *LineScanner {
	res := LineScanner{
		out: make(chan string),
		ctx: ctx,
	}
	go func() {
		defer close(res.out)
		reader := bufio.NewScanner(in)
		for reader.Scan() {
			select {
			case res.out <- reader.Text():
			case <-res.ctx.Done():
				res.err = res.ctx.Err()
				return
			}
		}
		res.err = reader.Err()
	}()
	return &res
}

// ErrMaxFramesExceeded is returned by (*LoopingLineScanner).Err() if its
// underlying io.Reader provides more lines of input than its specified maximum
// number of frames.
var ErrMaxFramesExceeded = errors.New("maximum number of frames exceeded")

// LoopingLineScanner will first consume an entire underlying io.Reader until
// EOF, and then continuously loop its lines on the Frames channel and never
// close, unless its internal context is cancelled.
//
// A LoopingLineScanner has little practical usage (known to the author anyhow)
// outside of creating animations that loop continulously, e.g. art and memes!
type LoopingLineScanner struct {
	out chan string
	buf []string
	ctx context.Context
	err error
}

// Frames returns a channel which will loop over the scanner's frames forever.
//
// The Frames channel will not begin sending data until the LoopingLineScanner
// has finished consuming the underlying io.Reader to EOF.
func (s *LoopingLineScanner) Frames() <-chan string {
	return s.out
}

// Err returns the underlying error which was the cause of the
// LoopingLineScanner closing its Frames channel.
//
// The likely scenarios where this would occur are either an IO error during the
// initial consumption of the underlying io.Reader (in which case, this error
// will occur prior to any values being sent over the Frames channel), an
// io.Reader that provides more lines than the configured maxFrames for the
// scanner, or the completion of the scanner's context.
func (s *LoopingLineScanner) Err() error {
	return s.err
}

// NewLoopingLineScanner generates a LoopingLineScanner which will first consume
// an entire io.Reader until EOF, and then continuously loop its lines on the
// Frames() channel and never close unless its underlying context is canceled.
//
// As a result, it is only suitable for an input value that will have an EOF, as
// otherwise it will continue consuming memory while never sending anything. You
// can mitigate this risk by providing the required maxFrames parameter: if the
// underlying io.Reader in exceeds this many lines of input, the Scanner will be
// halted with an error and the output channel closed. If maxFrames is 0, no
// checking will occur.
func NewLoopingLineScanner(ctx context.Context, in io.Reader, maxFrames int) *LoopingLineScanner {
	res := LoopingLineScanner{
		out: make(chan string),
		//buf: nil, /* nil is valid zero case for a slice */
		ctx: ctx,
	}

	go func() {
		defer close(res.out)
		// consume all lines into buf slice until EOF
		reader := bufio.NewScanner(in)
		for reader.Scan() {
			if maxFrames > 0 && len(res.buf) >= maxFrames {
				res.err = ErrMaxFramesExceeded
				return
			}
			if ctxDone := res.ctx.Err(); ctxDone != nil {
				res.err = ctxDone
				return
			}
			res.buf = append(res.buf, reader.Text())
		}
		if err := reader.Err(); err != nil {
			res.err = err
			return
		}
		// iterate over buf array as output forever
		for {
			for _, frame := range res.buf {
				select {
				case res.out <- frame:
				case <-res.ctx.Done():
					res.err = res.ctx.Err()
					return
				}
			}
		}
	}()
	return &res
}
