package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)

type WebAHandler struct {}

func (WebAHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("webA"))
}

func main() {
	c := make(chan os.Signal)
	go func() {
		http.ListenAndServe(":9000", WebAHandler{})
	}()

	signal.Notify(c, os.Interrupt)
	s := <- c
	log.Println(s)
}
