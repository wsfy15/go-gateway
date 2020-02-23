package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

type Web1Handler struct {}

func (this Web1Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	auth := request.Header.Get("Authorization")
	if auth == "" {
		writer.Header().Set("WWW-Authenticate", `Basic realm="必须输入用户名和密码"`)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	auth_list := strings.Split(auth, " ")
	if len(auth_list) == 2 && auth_list[0] == "Basic" {
		decodeString, err := base64.StdEncoding.DecodeString(auth_list[1])
		if err != nil {
			writer.Write([]byte("用户名或密码错误"))
		}
		if string(decodeString) == "sf:123" {
			writer.Write([]byte(fmt.Sprintf("welcome from %s", this.GetIP(request))))
			return
		}

	}
	writer.Write([]byte("用户名或密码错误"))
}

func (Web1Handler)GetIP(r *http.Request) string {
	ips := r.Header.Get("x-forwarded-for")
	if ips != "" {
		ips_list := strings.Split(ips, ",")
		if len(ips_list) > 0 && ips_list[0] != "" {
			return ips_list[0]
		}
	}
	return r.RemoteAddr
}

type Web2Handler struct {}

func (Web2Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Write([]byte("web2"))
}

func main() {
	c := make(chan os.Signal)
	go func() {
		http.ListenAndServe(":9000", Web1Handler{})
	}()

	go func() {
		http.ListenAndServe(":9001", Web2Handler{})
	}()

	signal.Notify(c, os.Interrupt)
	s := <- c
	log.Println(s)
}
