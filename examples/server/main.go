package main

import (
	"context"
	"fmt"
	"net/http"

	ioa "github.com/iotaledger/iota.go/api"
	"github.com/pkartner/tau"
)

const host = "https://node01.iotatoken.nl:443"

type server struct {
	iotaAPI *ioa.API
}

type contextKey string

var tangleIDKey = contextKey("tangle-id-key")

func (s *server) authMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		println("Request Received")
		println("Verifying...")
		tangleID, err := tau.VerifyRequest(s.iotaAPI, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}
		if tangleID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("couldn't authorize")
			return
		}
		println("Request Authorized")
		r = r.WithContext(context.WithValue(r.Context(), tangleIDKey, tangleID))
		h(w, r)
	}
}

func (s *server) handler(w http.ResponseWriter, r *http.Request) {
	// Do something interesting
	contextValue := r.Context().Value(tangleIDKey)
	if contextValue == nil {
		fmt.Println("context value is missing")
		return
	}
	tangleID, ok := contextValue.(*tau.TangleID)
	if !ok {
		fmt.Println("context value is not tangleID")
		return
	}
	w.Write([]byte(tangleID.Email))
}

func main() {
	api, err := ioa.ComposeAPI(ioa.HTTPClientSettings{
		URI: host,
	})
	if err != nil {
		panic(err)
	}
	s := server{
		iotaAPI: api,
	}
	http.Handle("/", s.authMiddleware(s.handler))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
