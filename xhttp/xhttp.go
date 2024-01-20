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
			cancel:          cancel,
			rollingDeadline: time.Time{},
			state:           new(atomic.Uint32),
			Request:         r,
			headers:         make(http.Header),
			ResponseWriter:  w,
		}

		timout := func() {
			tw.Timeout()
			done <- struct{}{}
		}

		defer time.AfterFunc(opts.Initial, timout).Stop()

		go func() {
			handler.ServeHTTP(&tw, r)
			done <- struct{}{}
		}()

		<-done
	})
}

const (
	pending = iota
	timeout
	writing
)

type timeoutWriter struct {
	TimeoutOptions

	cancel context.CancelCauseFunc

	rollingDeadline time.Time

	state   *atomic.Uint32
	Request *http.Request
	headers http.Header
	http.ResponseWriter
}

func (w timeoutWriter) Timeout() {
	if !w.state.CompareAndSwap(pending, timeout) {
		return
	}
	w.cancel(fmt.Errorf("%w: %w", context.Canceled, ErrTimeoutBeforeWrite))
	w.Handler.ServeHTTP(w.ResponseWriter, w.Request)
}

func (w timeoutWriter) Header() http.Header {
	if w.tryWriting() {
		return w.ResponseWriter.Header()
	}
	return w.headers
}

func (w timeoutWriter) tryWriting() bool {
	return w.state.CompareAndSwap(pending, writing) || w.state.Load() == writing
}

func (w timeoutWriter) WriteHeader(status int) {
	if !w.tryWriting() {
		return
	}
	for name := range w.headers {
		w.ResponseWriter.Header().Set(name, w.headers.Get(name))
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *timeoutWriter) Write(data []byte) (int, error) {
	if !w.tryWriting() {
		return 0, ErrTimeoutBeforeWrite
	}

	if !w.rollingDeadline.IsZero() && time.Now().After(w.rollingDeadline) {
		return 0, ErrTimeoutDuringWrite
	}

	if w.Rolling > 0 {
		defer func() { w.rollingDeadline = time.Now().Add(w.Rolling) }()
	}

	return w.ResponseWriter.Write(data)
}
