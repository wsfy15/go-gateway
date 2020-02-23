package utils

import (
	"io/ioutil"
	"net/http"
)

func CloneHeader(src, dst http.Header) {
	for k, v := range src {
		dst.Set(k, v[0])
	}
}

func Request(writer http.ResponseWriter, request *http.Request, url string) {
	newRequest, _ := http.NewRequest(request.Method, url, request.Body)
	CloneHeader(request.Header, newRequest.Header)	// 将真实请求的头拷贝到代理请求中，例如 Basic Auth头
	// Header.Add会扩展(key, value)对，而Header.Set会replace
	newRequest.Header.Add("x-forwarded-for", request.RemoteAddr)
	newResponse, _ := http.DefaultClient.Do(newRequest)

	header := writer.Header()
	CloneHeader(newResponse.Header, header)	// 将服务器响应的头 拷贝 到发往用户的响应中
	writer.WriteHeader(newResponse.StatusCode)

	defer newResponse.Body.Close()
	bytes, _ := ioutil.ReadAll(newResponse.Body)
	writer.Write(bytes)
}