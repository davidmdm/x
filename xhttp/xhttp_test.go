package xhttp_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidmdm/x/xhttp"
	"github.com/stretchr/testify/require"
)

func TestTimeoutHandler(t *testing.T) {
	cases := []struct {
		Name    string
		Handler http.HandlerFunc

		Opts xhttp.TimeoutOptions

		ExpectedReadError func(*testing.T, error)

		ExpectedStatus int
		ExpectedHeader map[string]string
		ExpectedBody   string
	}{
		{
			Name: "happy",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Test-Dirty-Write", "true")
				io.WriteString(w, "success!")
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
			Handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(10 * time.Millisecond)
				w.Header().Set("Test-Dirty-Write", "true")
				io.WriteString(w, "success!")
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
		},
		{
			Name: "happying rolling response",
			Handler: func(w http.ResponseWriter, r *http.Request) {
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
					io.WriteString(w, value.data)
				}
			},
			Opts:           xhttp.TimeoutOptions{Rolling: 5 * time.Millisecond},
			ExpectedStatus: 200,
			ExpectedHeader: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedBody: `{"hello":"world"}`,
		},
		{
			Name: "failed rolling response",
			Handler: func(w http.ResponseWriter, r *http.Request) {
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
						return
					}
				}
			},
			Opts: xhttp.TimeoutOptions{Rolling: 5 * time.Millisecond},
			ExpectedReadError: func(t *testing.T, err error) {
				require.True(t, errors.Is(err, io.EOF))
			},
			ExpectedStatus: 200,
			ExpectedHeader: map[string]string{
				"Content-Type": "application/json",
			},
			ExpectedBody: `{"hello":"wor`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			handler := xhttp.TimeoutHandler(tc.Handler, tc.Opts)

			server := httptest.NewServer(handler)
			defer server.Close()

			resp, err := http.Get(server.URL)
			require.NoError(t, err)

			defer resp.Body.Close()

			require.Equal(t, tc.ExpectedStatus, resp.StatusCode)

			for key, value := range tc.ExpectedHeader {
				require.Equal(t, value, resp.Header.Get(key), "unexpected value for header %s", key)
			}

			body, err := io.ReadAll(resp.Body)
			if tc.ExpectedReadError != nil {
				tc.ExpectedReadError(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.ExpectedBody, string(body))
		})
	}
}
