package xhttp_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/davidmdm/x/xhttp"
	"github.com/stretchr/testify/require"
)

func TestTimeoutHandler(t *testing.T) {
	cases := []struct {
		Name    string
		Handler func(http.ResponseWriter, *http.Request) error

		Opts xhttp.TimeoutOptions

		ExpectedResponseError func(*testing.T, error)
		ExpectedReadError     func(*testing.T, error)
		ExpectedServeError    func(*testing.T, error)

		ExpectedStatus int
		ExpectedHeader map[string]string
		ExpectedBody   string
	}{
		{
			Name: "happy",
			Handler: func(w http.ResponseWriter, r *http.Request) error {
				w.Header().Set("Test-Dirty-Write", "true")
				_, err := io.WriteString(w, "success!")
				return err
			},
			Opts: xhttp.TimeoutOptions{
				Initial: 50 * time.Millisecond,
			},
			ExpectedStatus: 200,
			ExpectedHeader: map[string]string{},
			ExpectedBody:   "success!",
		},
		{
			Name: "basic initial timeout",
			Handler: func(w http.ResponseWriter, r *http.Request) error {
				time.Sleep(10 * time.Millisecond)
				w.Header().Set("Test-Dirty-Write", "true")
				_, err := io.WriteString(w, "success!")
				return err
			}, Opts: xhttp.TimeoutOptions{
				Initial: 1 * time.Millisecond,
			},
			ExpectedStatus: 503,
			ExpectedHeader: map[string]string{
				"Content-Type":     "text/html; charset=utf-8",
				"Content-Length":   "45",
				"Test-Dirty-Write": "",
			},
			ExpectedBody: "<html><body>Service Unavailable</body></html>",
			ExpectedServeError: func(t *testing.T, err error) {
				require.EqualError(t, err, "request timeout reached before write")
			},
		},
		{
			Name: "basic initial timeout with custom handler",
			Handler: func(w http.ResponseWriter, r *http.Request) error {
				time.Sleep(10 * time.Millisecond)
				w.Header().Set("Test-Dirty-Write", "true")
				_, err := io.WriteString(w, "success!")
				return err
			},
			Opts: xhttp.TimeoutOptions{
				Initial: 1 * time.Millisecond,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(500)
					json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
				}),
			},
			ExpectedStatus: 500,
			ExpectedHeader: map[string]string{
				"Content-Type":     "application/json",
				"Content-Length":   "32",
				"Test-Dirty-Write": "",
			},
			ExpectedBody: `{"error":"service unavailable"}` + "\n",
			ExpectedServeError: func(t *testing.T, err error) {
				require.EqualError(t, err, "request timeout reached before write")
			},
		},
		{
			Name: "happying rolling response",
			Handler: func(w http.ResponseWriter, r *http.Request) error {
				stream := []struct {
					data  string
					delay time.Duration
				}{
					{data: `{"hel`, delay: 2 * time.Millisecond},
					{data: `lo":`, delay: 2 * time.Millisecond},
					{data: `"wor`, delay: 2 * time.Millisecond},
					{data: `ld"`, delay: 2 * time.Millisecond},
					{data: `}`, delay: 2 * time.Millisecond},
				}

				w.Header().Set("Content-Type", "application/json")

				for _, value := range stream {
					time.Sleep(value.delay)
					if _, err := io.WriteString(w, value.data); err != nil {
						return err
					}
				}

				return nil
			},
			Opts:           xhttp.TimeoutOptions{Rolling: 5 * time.Millisecond},
			ExpectedStatus: 200,
			ExpectedHeader: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedBody: `{"hello":"world"}`,
		},
		{
			Name: "quick failed rolling response",
			Handler: func(w http.ResponseWriter, r *http.Request) error {
				stream := []struct {
					data  string
					delay time.Duration
				}{
					{data: `{"hel`, delay: 0 * time.Millisecond},
					{data: `lo":`, delay: 2 * time.Millisecond},
					{data: `"wor`, delay: 4 * time.Millisecond},
					{data: `ld"`, delay: 8 * time.Millisecond},
					{data: `}`, delay: 160 * time.Millisecond},
				}

				w.Header().Set("Content-Type", "application/json")

				for _, value := range stream {
					time.Sleep(value.delay)
					if _, err := io.WriteString(w, value.data); err != nil {
						return err
					}
				}
				return nil
			},
			Opts: xhttp.TimeoutOptions{Rolling: 5 * time.Millisecond},
			ExpectedResponseError: func(t *testing.T, err error) {
				require.True(t, errors.Is(err, io.EOF), "expected error to be %v but got: %v", io.EOF, err)
			},
			ExpectedServeError: func(t *testing.T, err error) {
				require.EqualError(t, err, "request rolling timeout reached during write")
			},
		},
		{
			Name: "long failed rolling response",
			Handler: func(w http.ResponseWriter, r *http.Request) error {
				stream := []struct {
					data  string
					delay time.Duration
				}{
					{data: strings.Repeat("a", 4*1024*1024), delay: 0 * time.Millisecond},
					{data: strings.Repeat("b", 4*1024*1024), delay: 30 * time.Millisecond},
				}

				w.Header().Set("Content-Type", "application/json")

				for _, value := range stream {
					time.Sleep(value.delay)
					if _, err := io.WriteString(w, value.data); err != nil {
						return err
					}
				}

				return nil
			},
			Opts: xhttp.TimeoutOptions{Rolling: 20 * time.Millisecond},
			ExpectedReadError: func(t *testing.T, err error) {
				require.EqualError(t, err, "unexpected EOF")
			},
			ExpectedServeError: func(t *testing.T, err error) {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), "request rolling timeout reached during write")
				require.Contains(t, err.Error(), "i/o timeout")
			},

			ExpectedStatus: 200,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			serveErr := make(chan error, 1)
			var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serveErr <- tc.Handler(w, r)
				close(serveErr)
			})

			handler = xhttp.TimeoutHandler(handler, tc.Opts)

			server := httptest.NewServer(handler)
			defer server.Close()

			defer func() {
				if tc.ExpectedServeError != nil {
					tc.ExpectedServeError(t, <-serveErr)
					return
				}
				require.NoError(t, <-serveErr)
			}()

			resp, err := http.Get(server.URL)
			if tc.ExpectedResponseError != nil {
				tc.ExpectedResponseError(t, err)
				return
			}
			require.NoError(t, err)

			defer resp.Body.Close()

			require.Equal(t, tc.ExpectedStatus, resp.StatusCode)

			for key, value := range tc.ExpectedHeader {
				require.Equal(t, value, resp.Header.Get(key), "unexpected value for header %s", key)
			}

			body, err := io.ReadAll(resp.Body)
			if tc.ExpectedReadError != nil {
				tc.ExpectedReadError(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.ExpectedBody, string(body))
		})
	}
}
