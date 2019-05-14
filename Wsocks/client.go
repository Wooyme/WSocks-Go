package Wsocks

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	host string
	user string
	pass string
	State string
	Log string
}

var ws *websocket.Conn= nil
var msg = make(chan []byte)
func NewClient(host,user,pass string) *Client{
	return &Client{
		host:host,
		user:user,
		pass:pass,
		State:"No Action",
		Log:"",
	}
}

func (c *Client) Start(){
	conn:=c.connect()
	if conn==nil {
		return
	}
	ws = conn
	sock:=make(chan []byte)
	go func() {
		defer ws.Close()
		for {
			mst,message,err:=ws.ReadMessage()
			if err!=nil {
				fmt.Printf("Connection closed,Err: %v \n",err)
				c.Log+=fmt.Sprintf("Connection closed,Err: %v \n",err)
				c.State = "Waiting"
				TrayState.SetTitle(c.State)
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
			err:=ws.WriteMessage(websocket.BinaryMessage, buf)
			if err != nil {
				fmt.Printf("Err when writeMessage: %v \n",err)
				c.Log+=fmt.Sprintf("Err when writeMessage: %v \n",err)
			}
		}
	}()
	server:=NewSocks("127.0.0.1:1080", msg, sock,NewAes([]byte(c.pass),true))
	server.Listen()
}

func (c *Client) EditRemote(host,user,pass string){
	c.host = host
	c.user = user
	c.pass = pass
	conn:=c.connect()
	if conn == nil {
		return
	}
	tmp:=ws
	ws = conn
	tmp.Close()
	go func() {
		defer ws.Close()
		for {
			mst,message,err:=ws.ReadMessage()
			if err!=nil {
				fmt.Printf("Connection closed,Err: %v \n",err)
				c.State = "Waiting"
				TrayState.SetTitle(c.State)
				c.Log+=fmt.Sprintf("Connection closed,Err: %v \n",err)
				return
			}
			if mst == websocket.BinaryMessage {
				msg <- message
			}
		}
	}()
}

func (c *Client) connect() *websocket.Conn{
	hash:=md5.New()
	hash.Write([]byte("1"+c.user+c.pass))
	secret:= hex.EncodeToString(hash.Sum(nil))
	u := url.URL{Scheme: "ws", Host: c.host, Path: "/echo"}
	header := http.Header{}
	header.Add("auth",secret)
	conn,_,err := websocket.DefaultDialer.Dial(u.String(),header)
	if err!=nil{
		log.Fatal(err)
		return nil
	}
	c.State = "Connected"
	TrayState.SetTitle(fmt.Sprintf("%v,%v",c.host,c.State))
	fmt.Println("Connected to remote server")
	return conn
}