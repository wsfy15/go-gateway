package test_test

import (
	"gopkg.in/ini.v1"
	"log"
	"testing"
)

func TestINI(t *testing.T) {
	cfg, err := ini.Load("env")
	if err != nil {
		t.Fatal(err)
		return
	}
	proxy, err := cfg.GetSection("proxy")
	if err != nil {
		t.Fatal(err)
		return
	}

	if proxy != nil {
		secs := proxy.ChildSections()
		for _, sec := range secs {
			path, _ := sec.GetKey("path")
			pass, _ := sec.GetKey("pass")
			if path != nil && pass != nil {
				log.Println(path, pass)
			}
		}
	}
}
