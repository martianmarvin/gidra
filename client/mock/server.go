package mock

import (
	"net/http"
	"net/http/httptest"
	"sync"
)

var (
	responsesMu sync.RWMutex
	responses   = make(map[string][]byte)
)

//NewServer Spins up a test HTTP server that prints the request by default
// If the request has the X-Test-Response header set, the value of the
// registered response with that key is printed instead
func NewServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseKey := r.Header.Get("X-Test-Response")
		if len(responseKey) > 0 {
			responsesMu.RLock()
			defer responsesMu.RUnlock()
			resp, ok := responses[responseKey]
			if !ok {
				w.WriteHeader(500)
				w.Write([]byte("No test response registered for " + responseKey))
			}
			w.Write(resp)
		} else {
			r.Write(w)
		}
	}))
}

// RegisterResponse saves response text that can be later retrieved by
// mkaing a request to the server with the X-Test-Response header set to the
// previously registered key
func RegisterResponse(key string, resp []byte) {
	responsesMu.Lock()
	defer responsesMu.Unlock()
	responses[key] = resp
}
