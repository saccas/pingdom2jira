package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

// LoggerMiddleware writes basic request information as JSON to standard output.
func (s Server) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		rec := statusRecorder{w, 200}

		next.ServeHTTP(&rec, r)

		type output struct {
			Time      time.Time `json:"time"`
			Path      string    `json:"path"`
			UserAgent string    `json:"user_agent"`
			Status    int       `json:"status"`
		}

		out := output{
			Time:      t,
			Path:      r.URL.Path,
			UserAgent: r.UserAgent(),
			Status:    rec.status,
		}

		b, _ := json.Marshal(out)
		fmt.Println(string(b))
	})
}
