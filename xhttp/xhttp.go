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

const defaultTimeoutResponse = `<html><body>Service Unavailable</body></html>`

func DefaultTimeoutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(503)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(defaultTimeoutResponse)))
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
		opts.Handler = http.HandlerFunc(DefaultTimeoutHandler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancelCause(r.Context())
		defer cancel(nil)

		r = r.WithContext(ctx)

		tw := timeoutWriter{
			TimeoutOptions:     opts,
			cancel:             cancel,
			state:              new(atomic.Uint32),
			Request:            r,
			headers:            make(http.Header),
			ResponseWriter:     w,
			ResponseController: http.NewResponseController(w),
		}

		defer time.AfterFunc(opts.Initial, tw.Timeout).Stop()

		handler.ServeHTTP(&tw, r)
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

	timer *time.Timer

	state   *atomic.Uint32
	Request *http.Request
	headers http.Header
	http.ResponseWriter
	*http.ResponseController
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

	defer func() {
		if w.Rolling <= 0 {
			return
		}

		w.ResponseController.SetWriteDeadline(time.Now().Add(w.Rolling))

		if w.timer != nil {
			w.timer.Reset(w.Rolling)
			return
		}

		w.timer = time.AfterFunc(w.Rolling, func() {
			w.cancel(fmt.Errorf("%w: %w", context.Canceled, ErrTimeoutDuringWrite))
		})
	}()

	return w.ResponseWriter.Write(data)
}
