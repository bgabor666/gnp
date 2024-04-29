package main

import (
	"flag"
	"io/ioutil"
	"log"

	tftp "github.com/bgabor666/gnp/ch06"
)


var (
    address = flag.String("a", "127.0.0.1:69", "listen address")
    payload = flag.String("p", "payload.svg", "file to serve to clients")
)


func main() {
    flag.Parse()

    payload, err := ioutil.ReadFile(*payload)
    if err != nil {
	log.Fatal(err)
    }

    server := tftp.Server{Payload: payload}

    log.Fatal(server.ListenAndServe(*address))
}
