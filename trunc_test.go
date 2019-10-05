package main

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
)

type handler http.HandlerFunc

func (f *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello server")
}

func server() {
	h := new(handler)
	s := &http.Server{
		Addr:           ":1234",
		Handler:        h,
		ReadTimeout:    1000 * time.Second,
		WriteTimeout:   1000 * time.Second,
		MaxHeaderBytes: 1 << 32,
	}
	log.Fatal(s.ListenAndServe())
}

func TestTrunc(t *testing.T) {
	go server()
	do("localhost:1234", "/", "", "", 3, "http://")
	t.Log("finished")

}
