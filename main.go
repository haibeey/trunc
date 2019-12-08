package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"time"

	conn "github.com/haibeey/trunc/connection"
)

var (
	service   = flag.String("service", "localhost:8080", "The server url  to test")
	resource  = flag.String("resource", "/", "The server resource to get")
	username  = flag.String("username", "", "username for authetication")
	password  = flag.String("password", "", "")
	protocol  = flag.String("protocol", "http://", "Protocol scheme to request url")
	concLevel = flag.Uint64("concrs", 3, "The number of concurrent connection to test")
	startreq  = flag.Int("startreq", 10, "The number of start request to make for average calculation")
	usetls       = flag.Bool("tls", false, "Use tls for connection")
	certpath  = flag.String("certpath", "./cert/cert.pem", "absolute path to tls certificate ")
)

func request(svc string, rsr string, prt string) time.Duration {

	startTime := time.Now()
	client := &http.Client{}

	req, err := http.NewRequest("GET", prt+svc+rsr, nil)
	if err != nil {
		fmt.Println("Error while proccessing this request ", err)
		panic(err)
	}

	if *username == "" {
		req.SetBasicAuth(*username, *password)
	}

	resp, err := client.Do(req)

	seeResponse(resp)

	return time.Now().Sub(startTime)
}

func seeResponse(resp *http.Response) {

}

func do(service string, resource string, username string, password string, concLevel uint64, protocol string) {
	fmt.Println(protocol + service + resource)

	//make first request when no connections to server
	avgLatency := request(service, resource, protocol).Seconds()
	for i := 0; i < 50; i++ {
		avgLatency = (avgLatency + request(service, resource, protocol).Seconds()) / 2
	}

	ticker := time.NewTicker(5000 * time.Millisecond)
	done := make(chan bool)
	//https://en.wikipedia.org/wiki/Basic_access_authentication
	//Authorization: Basic username:password
	var encdusrpwd string
	if username != "" {
		encdusrpwd = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}

	cg := conn.NewConnGroup(resource, encdusrpwd)

	var level uint64
	for level = 1; level <= concLevel; level++ {
		cg.AddConn(conn.NewConn(level))
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("receive ticker at ", t)
				cg.Read()
			}
		}
	}()

	// make a request while server has multiple connections
	avgLatencyConn := request(service, resource, protocol).Seconds()
	//loop till all connection is closed
	for !cg.Done() {
		avgLatencyConn = (avgLatencyConn + request(service, resource, protocol).Seconds()) / 2
	}
	cg.ReleaseAll()

	//wait till all read completes
	done <- true

	fmt.Println(avgLatency, avgLatencyConn)
	fmt.Println("finishing")
}

func main() {
	flag.Parse()
	//make comparison
	do(*service, *resource, *username, *password, *concLevel, *protocol)
}
