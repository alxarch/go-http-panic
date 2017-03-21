package httpanic

import (
	"net/http"
)

type Error struct {
	Code    int
	Message string
}

func (p Error) Error() string {
	if "" == p.Message {
		return http.StatusText(p.Code)
	}
	return p.Message
}
func (p Error) HTTPStatus() int {
	return p.Code
}

type HTTPError interface {
	Error() string
	HTTPStatus() int
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(ok bool, code int, message string) {
	if !ok {
		panic(Error{code, message})
	}
}

func AssertError(err error, code int) {
	if err != nil {
		panic(Error{code, err.Error()})
	}
}

func Panic(code int, message string) {
	panic(Error{code, message})
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				if hp, ok := e.(HTTPError); ok {
					// Handle http coded errors gracefully
					http.Error(w, hp.Error(), hp.HTTPStatus())
				} else {
					panic(e)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
