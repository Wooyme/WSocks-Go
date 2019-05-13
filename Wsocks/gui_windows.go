// +build windows

package Wsocks

import (
	"golang.org/x/sys/windows/registry"
	"log"
)

func winAutoProxy(pac bool){
	if pac {
		k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
		if err != nil {
			log.Fatal(err)
		}
		defer k.Close()
		_ = k.SetStringValue("AutoConfigURL", "https://blackwhite.txthinking.com/white.pac")
		_ = k.SetDWordValue("ProxyEnable", 0)
	}else{
		k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.SET_VALUE)
		if err != nil {
			log.Fatal(err)
		}
		defer k.Close()
		_ = k.SetStringValue("AutoConfigURL", "http://localhost:1082/pac")
		_ = k.SetDWordValue("ProxyEnable", 0)
	}
}
