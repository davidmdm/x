package xhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

type TimeoutOptions struct {
	Initial time.Duration
	Rolling time.Duration
	Handler http.Handler
}

func TimeoutHandler(handler http.Handler, opts TimeoutOptions) http.Handler {
	if opts.Initial <= 0 && opts.Rolling <= 0 {
		return handler
	}
	if opts.Initial <= 0 {
		opts.Initial = opts.Rolling
	}
	if opts.Handler == nil {
		opts.Handler = http.HandlerFunc(defaultTimeoutHandler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancelCause(r.Context())
		defer cancel(nil)

		r = r.WithContext(ctx)

		done := make(chan struct{}, 2)

		tw := timeoutWriter{
			TimeoutOptions:  opts,
			done:            done,
			cancel:          cancel,
			rollingDeadline: time.Time{},
			state:           new(atomic.Uint32),
			headers:         make(http.Header),
			ResponseWriter:  w,
			Request:         r,
			Controller:      http.NewResponseController(w),
		}

		defer time.AfterFunc(opts.Initial, tw.Timeout).Stop()

		go func() {
			handler.ServeHTTP(&tw, r)
			done <- struct{}{}
		}()

		<-done
	})
}

var (
	ErrTimeoutBeforeWrite = errors.New("request timeout reached before write")
	ErrTimeoutDuringWrite = errors.New("request rolling timeout reached during write")
)

func defaultTimeoutHandler(w http.ResponseWriter, r *http.Request) {
	const defaultTimeoutResponse = `<html><body>Service Unavailable</body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(defaultTimeoutResponse)))
	w.WriteHeader(503)
	io.WriteString(w, defaultTimeoutResponse)
}

const (
	pending = iota
	timeout
	writing
)

type timeoutWriter struct {
	TimeoutOptions

	done   chan<- struct{}
	cancel context.CancelCauseFunc

	rollingDeadline time.Time

	state   *atomic.Uint32
	headers http.Header
	http.ResponseWriter
	Request    *http.Request
	Controller *http.ResponseController
}

func (w timeoutWriter) Timeout() {
	if !w.state.CompareAndSwap(pending, timeout) {
		return
	}
	w.cancel(fmt.Errorf("%w: %w", context.Canceled, ErrTimeoutBeforeWrite))
	w.Handler.ServeHTTP(w.ResponseWriter, w.Request)
	w.done <- struct{}{}
}

func (w timeoutWriter) Header() http.Header {
	if w.state.Load() == writing {
		return w.ResponseWriter.Header()
	}
	return w.headers
}

func (w timeoutWriter) tryWriting() bool {
	if w.state.CompareAndSwap(pending, writing) {
		for key := range w.headers {
			w.ResponseWriter.Header().Set(key, w.headers.Get(key))
		}
	}
	return w.state.Load() == writing
}

func (w timeoutWriter) WriteHeader(status int) {
	if !w.tryWriting() {
		return
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *timeoutWriter) Write(data []byte) (n int, err error) {
	if !w.tryWriting() {
		return 0, ErrTimeoutBeforeWrite
	}

	if !w.rollingDeadline.IsZero() && time.Now().After(w.rollingDeadline) {
		w.Controller.SetWriteDeadline(w.rollingDeadline.Add(-time.Hour))
		if _, deadlineErr := w.ResponseWriter.Write(data); deadlineErr != nil {
			err = fmt.Errorf("%w: %w", ErrTimeoutDuringWrite, deadlineErr)
		} else {
			err = ErrTimeoutDuringWrite
		}
		return
	}

	n, err = w.ResponseWriter.Write(data)
	if err != nil {
		return
	}

	if w.Rolling > 0 {
		w.rollingDeadline = time.Now().Add(w.Rolling)
	}

	return
}

// Unwrap satisfies the implicit http.rwUnwrapper interface.
func (w timeoutWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
