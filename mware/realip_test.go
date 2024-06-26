package mware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"zgo.at/zhttp"
)

func TestRealIP(t *testing.T) {
	tests := []struct {
		remoteAddr string
		header     http.Header
		want       string
	}{
		// Remote addr
		{"1.1.1.1:42", nil, "1.1.1.1"},
		{"1.1.1.1", nil, "1.1.1.1"},

		// CF-Connecting-IP
		{"1.2.3.4", http.Header{"Cf-Connecting-Ip": {"4.4.4.4"}}, "4.4.4.4"},

		// X-Real-IP
		{"1.1.1.1", http.Header{"X-Real-Ip": {"101.100.100.100"}}, "101.100.100.100"},
		{"1.1.1.1:42", http.Header{"X-Real-Ip": {"101.100.100.100"}}, "101.100.100.100"},
		{"4006:beef::0", http.Header{"X-Real-Ip": {"4006:dead::0"}}, "4006:dead::0"},

		// Fly-Client-Ip
		{"1.1.1.1", http.Header{"Fly-Client-Ip": {"101.100.100.100"}}, "101.100.100.100"},
		{"1.1.1.1:42", http.Header{"Fly-Client-Ip": {"101.100.100.100"}}, "101.100.100.100"},
		{"4006:beef::0", http.Header{"Fly-Client-Ip": {"4006:dead::0"}}, "4006:dead::0"},

		// X-Forwarded-For
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"101.100.100.100"}}, "101.100.100.100"},
		{"4006:beef::0", http.Header{"X-Forwarded-For": {"4006:dead::0"}}, "4006:dead::0"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"2.2.2.2, 101.100.100.100"}}, "101.100.100.100"},

		// Filter local
		{"1.1.1.1", http.Header{"X-Real-Ip": {"192.168.5.5"}}, "1.1.1.1"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"127.0.0.1"}}, "1.1.1.1"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"2.2.2.2, 127.0.0.1"}}, "2.2.2.2"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"2.2.2.2, fc00:123::1"}}, "2.2.2.2"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"9999::1, fc00:123::1"}}, "9999::1"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"2.2.2.2, 127.0.0.1"}}, "2.2.2.2"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"2.2.2.2, 127.0.0.1, 192.168.1.1"}}, "2.2.2.2"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"127.0.0.1, 2.2.2.2, 192.168.1.1"}}, "2.2.2.2"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"127.0.0.1, 2.2.2.2, localhost"}}, "2.2.2.2"},
		{"1.1.1.1", http.Header{"X-Forwarded-For": {"127.0.0.1, 2.2.2.2, example.com, localhost"}}, "2.2.2.2"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			realIP := ""
			handler := RealIP()(zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				realIP = r.RemoteAddr
				w.Write([]byte("Hello World"))
				return nil
			}))

			r, _ := http.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			r.Header = tt.header
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, r)

			if rr.Code != 200 {
				t.Fatalf("wrong response code: %d; wanted 200", rr.Code)
			}

			if realIP != tt.want {
				t.Fatalf("wrong IP: %q; wanted %q", realIP, tt.want)
			}
		})
	}
}
