package main

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
	"flag"
	//"crypto/tls"
)

type handler http.HandlerFunc

func (f *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello server")
}

func server() {
	h := new(handler)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        h,
		ReadTimeout:    1000 * time.Second,
		WriteTimeout:   1000 * time.Second,
		MaxHeaderBytes: 1 << 32,
	}
	log.Fatal(s.ListenAndServe())
}

func tlsServer(){
	log.Fatal(http.ListenAndServeTLS(":8080", "./cert/cert.pem", "./cert/key.pem", new(handler)))
}

func TestTrunc(t *testing.T) {

	usetls:=flag.Lookup("tls").Value.(flag.Getter).Get().(bool)

	if usetls{
		go tlsServer()
	}else{
		go server()
	}
	
	do("localhost:1234", "/", "", "", 3, "http://")
	t.Log("finished")

}
