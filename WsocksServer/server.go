package WsocksServer

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)
type Server struct {
	info ServerInfo
	authMap map[string]*User
}

type ServerInfo struct {
	Port int
	Users []User
}

type User struct {
	User string
	Pass string
	Multiple int
	Limit int
}
var (
	upgrader = websocket.Upgrader{} // use default options
	connMap = make(map[string]net.Conn)
	lock = sync.RWMutex{}
)
func ReadServerFromFile(filename string) *Server{
	f,err:=os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	bytes:=make([]byte,1024)
	ln,err:=f.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	decoder:=json.NewDecoder(strings.NewReader(string(bytes[:ln])))
	serverInfo:=ServerInfo{}

	if err:=decoder.Decode(&serverInfo); err != nil{
		log.Fatal(err)
	}
	authMap:=make(map[string]*User)

	for _,user:=range serverInfo.Users {
		hash:=md5.New()
		hash.Write([]byte("1"+user.User+user.Pass))
		secret:= hex.EncodeToString(hash.Sum(nil))
		authMap[secret] = &user
	}
	server:=Server{
		info:serverInfo,
		authMap:authMap,
	}
	return &server
}

func (s *Server) Start(){
	addr := flag.String("addr", "0.0.0.0:"+strconv.Itoa(s.info.Port), "http service address")
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", s.echo)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func (s *Server) echo(w http.ResponseWriter, r *http.Request) {
	var user *User = nil
	for k:=range r.Header {
		if user=s.authMap[r.Header.Get(k)]; user!=nil {
			goto SUCCESS
		}
	}
	http.Error(w, http.StatusText(300), 300)
	return
	SUCCESS:

	fmt.Printf("Login successfully,%v \n",user.User)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	parser:=NewData(NewAes([]byte(user.Pass),true))
	wsChan:=make(chan []byte)
	defer c.Close()
	go func() {
		for{
			_bytes:=<-wsChan
			if err:=c.WriteMessage(websocket.BinaryMessage,_bytes);err != nil {
				return
			}
		}
	}()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		if mt != websocket.BinaryMessage {
			continue
		}
		_flag:=binary.LittleEndian.Uint32(message)
		switch _flag {
		case CONNECT:
			go handleConnect(wsChan,parser,parser.ParseClientConnect(message[4:]))
		case RAW:
			handleRaw(wsChan,parser,parser.ParseRaw(message[4:]))
		case EXCEPTION:
			handleException(wsChan,parser,parser.ParseException(message[4:]))
		case DNS:
			handleDnsQuery(wsChan,parser,parser.ParseDnsQuery(message[4:]))
		default:
			fmt.Printf("Unknown flag %v \n",_flag)
		}
	}
}

func handleConnect(c chan []byte,parser Data,data *ClientConnect){
	if data == nil { return }
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v",data.Host,data.Port))
	if err != nil {
		fmt.Printf("Connect failed, Err: %v \n",err)
		c<-parser.CreateException(data.Uuid,"Connection failed")
		return
	}
	go func() {
		defer func(){
			lock.Lock()
			delete(connMap,data.Uuid)
			lock.Unlock()
			conn.Close()
		}()
		for {
			_bytes:=make([]byte,4096)
			ln,err:=conn.Read(_bytes)
			if err != nil {
				return
			}
			c<-parser.CreateRaw(data.Uuid,_bytes[:ln])
		}
	}()
	lock.RLock()
	connMap[data.Uuid] = conn
	lock.RUnlock()
	c<-parser.CreateConnectSuccess(data.Uuid)
}

func handleRaw(c chan []byte,parser Data,data Raw){
	lock.RLock()
	conn:=connMap[data.Uuid]
	lock.RUnlock()
	if conn == nil {
		c<-parser.CreateException(data.Uuid,"Connection closed")
	}else{
		if _,err:=conn.Write(data.Data);err != nil {
			c<-parser.CreateException(data.Uuid,"Connection closed")
		}
	}
}

func handleException(c chan []byte,parser Data,data *Exception){
	if data == nil { return }
	lock.RLock()
	if conn:= connMap[data.Uuid];conn != nil {
		conn.Close()
	}
	lock.RUnlock()
}

func handleDnsQuery(c chan []byte,parser Data,data *DnsQuery) {
	if data == nil { return }
	ips, err := net.LookupIP(data.Host)
	if err != nil {
		_, _ = fmt.Printf("Could not get IPs: %v\n", err)
		c<-parser.CreateDnsQuery(data.Uuid,"0.0.0.0")
		return
	}
	if len(ips) > 0 {
		c<-parser.CreateDnsQuery(data.Uuid,ips[0].String())
	}else{
		c<-parser.CreateDnsQuery(data.Uuid,"0.0.0.0")
	}
}
