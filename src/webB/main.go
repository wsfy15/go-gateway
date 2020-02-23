package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
)

type WebBHandler struct {}

func (WebBHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("webB"))
}

func main() {
	c := make(chan os.Signal)
	go func() {
		http.ListenAndServe(":9001", WebBHandler{})
	}()

	signal.Notify(c, os.Interrupt)
	s := <- c
	log.Println(s)
}
