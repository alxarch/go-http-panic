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

type PanicHandler func(id int64, err interface{}, stacktrace []byte, r *http.Request)

var logmu sync.Mutex

func defaultPanicHandler(id int64, err interface{}, stacktrace []byte, r *http.Request) {
	logmu.Lock()
	defer logmu.Unlock()
	log.Printf("panic=%016x message = %v\n", id, err)
	// for _, line := range stacktrace {
	// 	log.Printf("panic=%016x %s", id, line)
	// }
}

var Middleware = PanicHandler(defaultPanicHandler).Middleware

const StackSize = 8192

func (h PanicHandler) handle(e interface{}, r *http.Request) int64 {
	stack := make([]byte, StackSize)
	stack = stack[:runtime.Stack(stack, false)]
	id := rand.Int63()
	h(id, e, stack, r)
	return id

}
func (h PanicHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				defer func() {
					if ee := recover(); ee != nil {
						log.Printf("Panic handler for %v panicked: %v", e, ee)
					}
				}()
				if hp, ok := e.(httpanic); ok {
					// Handle http coded errors gracefully
					http.Error(w, hp.Error(), hp.Code)
				} else {
					id := h.handle(e, r)
					body := fmt.Sprintf(
						"%s\n%016x",
						http.StatusText(http.StatusInternalServerError),
						id)
					http.Error(w, body, http.StatusInternalServerError)
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
