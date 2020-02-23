package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)

type WebCHandler struct {}

func (WebCHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("webC"))
}

func main() {
	c := make(chan os.Signal)
	go func() {
		http.ListenAndServe(":9002", WebCHandler{})
	}()

	signal.Notify(c, os.Interrupt)
	s := <- c
	log.Println(s)
}
