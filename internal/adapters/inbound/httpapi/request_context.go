package httpapi

import (
	"context"
	"net/http"
)

type requestContextKey struct{}

type RequestContext struct {
	BasicUsername   string
	BasicPassword   string
	BasicAuthSet    bool
	InstanceURL     string
	UserInfoBaseURL string
}

func RequestContextFrom(ctx context.Context) RequestContext {
	reqCtx, _ := ctx.Value(requestContextKey{}).(RequestContext)
	return reqCtx
}

func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCtx := RequestContext{}
		if username, password, ok := r.BasicAuth(); ok {
			reqCtx.BasicUsername = username
			reqCtx.BasicPassword = password
			reqCtx.BasicAuthSet = true
		}
		if r.Host != "" {
			scheme := "http"
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				scheme = "https"
			}
			reqCtx.InstanceURL = scheme + "://" + r.Host
			reqCtx.UserInfoBaseURL = "http://" + r.Host
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestContextKey{}, reqCtx)))
	})
}
