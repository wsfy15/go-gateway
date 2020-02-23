package main

import (
	"gopkg.in/ini.v1"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

var ProxyConfigs map[string]string

func init() {
	ProxyConfigs = make(map[string]string)
	cfg, err := ini.Load("env")
	if err != nil {
		log.Fatal(err)
		return
	}

	proxy, err := cfg.GetSection("proxy")
	if err != nil {
		log.Fatal(err)
		return
	}

	if proxy != nil {
		secs := proxy.ChildSections()
		for _, sec := range secs {
			path, _ := sec.GetKey("path")
			pass, _ := sec.GetKey("pass")
			if path != nil && pass != nil {
				ProxyConfigs[path.Value()] = pass.Value()
			}
		}
	}
}

type ProxyHandler struct {}

func (ProxyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
	}()

	for k, v := range ProxyConfigs {
		if matched, _ := regexp.MatchString(k, request.URL.Path); matched == true {
			// go 内置的代理，会处理cookie、重定向、timeout等
			target, _ := url.Parse(v)
			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ServeHTTP(writer, request)

			// 自己实现的代理
			//utils.Request(writer, request, v)
			return
		}
	}

	writer.Write([]byte("default index"))
}

func main() {
	http.ListenAndServe(":8000", ProxyHandler{})
}
