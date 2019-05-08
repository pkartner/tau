package main

import (
	"context"
	"fmt"
	"net/http"

	ioa "github.com/iotaledger/iota.go/api"
	"github.com/pkartner/tau"
)

const nodeURL = "https://node01.iotatoken.nl:443"

type server struct {
	iotaAPI *ioa.API
}

type contextKey string

var tangleIDKey = contextKey("tangle-id-key")

func (s *server) authMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Request received")
		fmt.Println("Verifying...")
		tangleID, err := tau.VerifyRequest(s.iotaAPI, r)
		if err == tau.ErrEmptyAuthorization {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("Could not authorize")
			return
		}
		if err == tau.ErrCallingTangle {
			w.WriteHeader(http.StatusServiceUnavailable)
			// Tangle node is not available right now normally we should have a backup node to try again
			fmt.Println("Tangle node is not available right now. You could try to change the nodeURL constant with a healthy node found here https://www.iotatoken.nl/")
			return
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("We received the following error when trying to authenticate ", err)
			return
		}
		if tangleID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("Could not authorize")
			return
		}
		fmt.Println("Request authorized")
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
	w.Write([]byte(tangleID.Name))
}

func main() {
	fmt.Println("Starting server...")
	api, err := ioa.ComposeAPI(ioa.HTTPClientSettings{
		URI: nodeURL,
	})
	if err != nil {
		panic(err)
	}
	s := server{
		iotaAPI: api,
	}
	fmt.Println("Ready! Please make a tangle authentication request.")
	http.Handle("/", s.authMiddleware(s.handler))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
