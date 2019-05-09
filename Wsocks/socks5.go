package Wsocks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	uuid2 "github.com/google/uuid"
	"io"
	"net"
	"sync"
)
const (
	ConnectCommand   = uint8(1)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

const (
	socks5Version = uint8(5)
)



type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}
type Request struct {
	Version uint8
	Command uint8
	RemoteAddr *AddrSpec
	DestAddr *AddrSpec
	realDestAddr *AddrSpec
	bufConn      io.Reader
	conn net.Conn
}

type Socks struct {
	address string
	buf chan []byte
	back chan []byte
	connMap map[string]*Request
	myAes Aes
	myData Data
}

var (
	unrecognizedAddrType = fmt.Errorf("Unrecognized address type")
	lock = sync.RWMutex{}
)

func NewSocks(addr string,back chan []byte,buf chan []byte,_aes Aes) Socks{
	return Socks{
		address:addr,
		back:back,
		buf:buf,
		connMap: map[string]*Request{},
		myAes:_aes,
		myData:NewData(_aes),
	}
}

func (s *Socks) Listen(){
	l,err:=net.Listen("tcp",s.address)
	if err!=nil{
		return
	}
	go func() {
		for {
			_bytes:=<-s.back
			flag:=binary.LittleEndian.Uint32(_bytes)
			switch flag {
			case CONNECT_SUCCESS:
				uuid:=s.myData.ParseConnectSuccess(_bytes[4:])
				lock.RLock()
				request := s.connMap[uuid]
				lock.RUnlock()
				if request == nil {
					continue
				}
				success:=[]byte{0x05,0x00,0x00,0x01,0x00,0x00,0x00,0x00,0x00,0x00}
				_, _ = request.conn.Write(success)
				go func() {
					p:=make([]byte,1024)
					for{
						_len,err:=request.bufConn.Read(p)
						if err != nil {
							_ = request.conn.Close()
							//s.buf<-s.myData.CreateException(uuid,"")
							lock.Lock()
							delete(s.connMap,uuid)
							lock.Unlock()
							return
						}
						s.buf<-s.myData.CreateRaw(uuid,p[:_len])
					}
				}()
			case RAW:
				raw:=s.myData.ParseRaw(_bytes[4:])
				lock.RLock()
				req:=s.connMap[raw.Uuid]
				lock.RUnlock()
				if req ==nil {
					continue
				}
				_, _ = req.conn.Write(raw.Data)
			case EXCEPTION:
				exception:=s.myData.ParseException(_bytes[4:])
				if exception == nil {
					continue
				}
				lock.RLock()
				req := s.connMap[exception.Uuid]
				lock.RUnlock()
				if req==nil {
					continue
				}
				_ = req.conn.Close()
				lock.Lock()
				delete(s.connMap,exception.Uuid)
				lock.Unlock()
			}
		}
	}()
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		go s.serveConn(conn)
	}
}

func (s *Socks) serveConn(conn net.Conn){
	bufConn := bufio.NewReader(conn)
	// Read the version byte
	version := []byte{0}
	if _, err := bufConn.Read(version); err != nil {
		fmt.Printf("[ERR] socks: Failed to get version byte: %v", err)
		return
	}

	if version[0] != socks5Version {
		fmt.Printf("[ERR] socks: Unsupported SOCKS version: %v", version)
		return
	}
	_, _ = readMethods(bufConn)
	_, err := conn.Write([]byte{socks5Version, uint8(0)})
	request, err := NewRequest(bufConn)
	if err!=nil {
		fmt.Printf("Error %v \n",err)
		return
	}
	request.conn = conn
	s.handleRequest(request,conn)
}

func readMethods(r io.Reader) ([]byte, error) {
	header := []byte{0}
	if _, err := r.Read(header); err != nil {
		return nil, err
	}

	numMethods := int(header[0])
	methods := make([]byte, numMethods)
	_, err := io.ReadAtLeast(r, methods, numMethods)
	return methods, err
}


// NewRequest creates a new Request from the tcp connection
func NewRequest(bufConn io.Reader) (*Request, error) {
	// Read the version byte
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(bufConn, header, 3); err != nil {
		return nil, fmt.Errorf("Failed to get command version: %v", err)
	}

	// Ensure we are compatible
	if header[0] != socks5Version {
		return nil, fmt.Errorf("Unsupported command version: %v", header[0])
	}

	// Read in the destination address
	dest, err := readAddrSpec(bufConn)
	if err != nil {
		return nil, err
	}

	request := &Request{
		Version:  socks5Version,
		Command:  header[1],
		DestAddr: dest,
		bufConn:  bufConn,
	}

	return request, nil
}

func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	d := &AddrSpec{}

	// Get the address type
	addrType := []byte{0}
	if _, err := r.Read(addrType); err != nil {
		return nil, err
	}

	// Handle on a per type basis
	switch addrType[0] {
	case ipv4Address:
		addr := make([]byte, 4)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case ipv6Address:
		addr := make([]byte, 16)
		if _, err := io.ReadAtLeast(r, addr, len(addr)); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case fqdnAddress:
		if _, err := r.Read(addrType); err != nil {
			return nil, err
		}
		addrLen := int(addrType[0])
		fqdn := make([]byte, addrLen)
		if _, err := io.ReadAtLeast(r, fqdn, addrLen); err != nil {
			return nil, err
		}
		d.FQDN = string(fqdn)

	default:
		return nil, unrecognizedAddrType
	}

	// Read the port
	port := []byte{0, 0}
	if _, err := io.ReadAtLeast(r, port, 2); err != nil {
		return nil, err
	}
	d.Port = (int(port[0]) << 8) | int(port[1])

	return d, nil
}

func (s *Socks) handleRequest(req *Request, conn net.Conn){
	// Switch on the command
	switch req.Command {
	case ConnectCommand:
		s.handleConnect(conn, req)
		return
	default:
		_ = conn.Close()
		fmt.Printf("Unavaliable request command")
		return
	}
}

func (s *Socks) handleConnect(conn net.Conn, req *Request){
	uuid,err:=uuid2.NewRandom()
	if err!=nil {
		fmt.Printf("Failed at generating uuid")
		return
	}
	var host string
	if req.DestAddr.FQDN!= "" {
		host = req.DestAddr.FQDN
	}else {
		host = req.DestAddr.IP.String()
	}
	lock.Lock()
	s.connMap[uuid.String()] = req
	lock.Unlock()
	s.buf<-s.myData.CreateClientConnect(uuid.String(),host,req.DestAddr.Port)
}