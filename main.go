package main

import (
	"Wsocks-Go/Wsocks"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
)


func main(){

	remote:=os.Args[1]
	user:=os.Args[2]
	pass:=os.Args[3]
	_zlib:=os.Args[4]
	hash:=md5.New()
	hash.Write([]byte(_zlib+user+pass))
	secret:= hex.EncodeToString(hash.Sum(nil))
	var addr = flag.String("addr",remote,"http service address")
	flag.Parse()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	header := http.Header{}
	header.Add("auth",secret)
	c,_,err := websocket.DefaultDialer.Dial(u.String(),header)
	if err!=nil{
		log.Fatal(err)
		return
	}
	fmt.Println("Connected to remote server")
	msg := make(chan []byte)
	sock:=make(chan []byte)
	go func() {
		defer c.Close()
		for {
			mst,message,err:=c.ReadMessage()
			if err!=nil {
				return
			}
			if mst == websocket.BinaryMessage {
				msg <- message
			}
		}
	}()
	go func() {
		for{
			buf:=<-sock
			c.WriteMessage(websocket.BinaryMessage,buf)
		}
	}()
	go func(){
		<-interrupt
		os.Exit(0)
	}()
	server:=Wsocks.NewSocks("127.0.0.1:1080", msg, sock,Wsocks.NewAes([]byte(pass),_zlib=="1"))
	server.Listen()
}