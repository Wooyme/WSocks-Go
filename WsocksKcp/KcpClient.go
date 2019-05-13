package WsocksKcp

import (
	"github.com/xtaci/kcp-go"
	"time"
)

func Test(){
	kcpClient:=kcp.NewKCP(100000, func(buf []byte, size int) {

	})
	kcpClient.NoDelay(1,20,2,1)
	go func() {
		time.Sleep(10)
		kcpClient.Update()
	}()
}