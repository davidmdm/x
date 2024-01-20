package xhttp_test

import (
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
		// {
		// 	Name: "happying rolling response",
		// 	Handler: func(http.ResponseWriter, *http.Request) {
		// 	},
		// 	Opts:           xhttp.TimeoutOptions{},
		// 	ExpectedStatus: 0,
		// 	ExpectedHeader: map[string]string{},
		// 	ExpectedBody:   "",
		// },
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
			require.NoError(t, err)

			require.Equal(t, tc.ExpectedBody, string(body))
		})
	}
}
