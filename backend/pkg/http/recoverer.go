package http

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recoverer captura panics em qualquer handler, loga o stack trace e devolve
// 500 ao client — sem isso um panic derrubaria o processo inteiro em vez de
// falhar só a request atual.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
