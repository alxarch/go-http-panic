package httpanic

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
)

type httpanic struct {
	Code    int
	Message string
}

func (p httpanic) Error() string {
	if "" == p.Message {
		return http.StatusText(p.Code)
	}
	return p.Message
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Assert(ok bool, code int, message string) {
	if !ok {
		panic(httpanic{code, message})
	}
}

func Error(err error, code int) {
	if err != nil {
		panic(httpanic{code, err.Error()})
	}
}

func Panic(code int, message string) {
	panic(httpanic{code, message})
}

type PanicHandler func(id int64, err interface{}, stacktrace []string, r *http.Request)

var logmu sync.Mutex

func defaultPanicHandler(id int64, err interface{}, stacktrace []string, r *http.Request) {
	logmu.Lock()
	defer logmu.Unlock()
	log.Printf("panic=%016x message = %v\n", id, err)
	for _, line := range stacktrace {
		log.Printf("panic=%016x %s", id, line)
	}
}

var Middleware = PanicHandler(defaultPanicHandler).Middleware

func (h PanicHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			if hp, ok := e.(httpanic); ok {
				http.Error(w, hp.Error(), hp.Code)
				return
			}
			id := rand.Int63()
			var lines []string
			for skip := 1; ; skip++ {
				pc, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				if file[len(file)-1] == 'c' {
					continue
				}
				f := runtime.FuncForPC(pc)
				s := fmt.Sprintf("%s:%d %s()\n", file, line, f.Name())
				lines = append(lines, s)
			}
			h(id, e, lines, r)
			body := fmt.Sprintf(
				"%s\n%016x",
				http.StatusText(http.StatusInternalServerError),
				id,
			)
			http.Error(w, body, http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}
