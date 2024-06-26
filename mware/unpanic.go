package mware

import (
	"fmt"
	"net/http"

	"zgo.at/zhttp"
	"zgo.at/zstd/zdebug"
)

// Unpanic recovers from panics in handlers and calls ErrPage().
func Unpanic(filterStack ...string) zhttp.Middleware {
	return func(next zhttp.HandlerFunc) zhttp.HandlerFunc {
		return zhttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}

				err, ok := rec.(error)
				if !ok {
					err = fmt.Errorf("panic at %s %s%s: %+v\n\nForm: %#v\nHeaders: %#v",
						r.Method, r.Host, r.RequestURI, rec, r.Form, r.Header)
				}

				err = fmt.Errorf("%w\n%s", err, zdebug.Stack(append(filterStack,
					"net/http", "zgo.at/zhttp", "github.com/go-chi/chi")...))
				zhttp.ErrPage(w, r, err)
			}()
			return next(w, r)
		})
	}
}
