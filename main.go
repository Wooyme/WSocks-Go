package main

import (
	"WSocks-Go/WsocksServer"
	"os"
)

/*Client Part*/
//func main(){
//	Wsocks.Tray()
//}


/*Server Part*/
func main(){
	cFile:=os.Args[1]
	server:=WsocksServer.ReadServerFromFile(cFile)
	server.Start()
}