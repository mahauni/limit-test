package middlewares

import "net/http"

type Middleware func(http.Handler) http.Handler

func CreateStack(xs ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			h = x(h)
		}

		return h
	}
}
