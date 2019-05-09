package WsocksServer

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
)

type Data struct {
	_aes Aes
}

const (
	CONNECT = 0
	CONNECT_SUCCESS = 1
	EXCEPTION = 2
	RAW = 3
	DNS = 6
)

type ClientConnect struct {
	Uuid string
	Host string
	Port uint32
}

type Exception struct {
	Uuid string
	Message string
}

type Raw struct {
	Uuid string
	Data []byte
}

type DnsQuery struct {
	Uuid string
	Host string
}

func NewData(_aes Aes) Data{
	return Data{
		_aes:_aes,
	}
}

func (d *Data) CreateClientConnect(uuid string,host string,port int) []byte{
	_json:=[]byte(fmt.Sprintf("{\"host\":\"%v\",\"port\":%v,\"uuid\":\"%v\"}",host,port,uuid))
	encrypted:=d._aes.AESEncrypt(_json)
	flag:=make([]byte,4)
	binary.LittleEndian.PutUint32(flag,CONNECT)
	buf:=bytes.NewBuffer(flag)
	buf.Write(encrypted)
	return buf.Bytes()
}

func (d *Data) ParseClientConnect(data []byte) *ClientConnect {
	decrypted:=d._aes.AESDecrypt(data)
	_dec:=json.NewDecoder(strings.NewReader(string(decrypted)))
	clientConnect:=ClientConnect{}
	if err := _dec.Decode(&clientConnect); err!=nil {
		fmt.Printf("err %v",err)
		return nil
	}else{
		return &clientConnect
	}
}

func (d *Data) CreateException(uuid string,message string) []byte {
	cMsg:=fmt.Sprintf("{\"uuid\":\"%v\",\"message\":\"%v\"}",uuid,message)
	flag:=make([]byte,4)
	binary.LittleEndian.PutUint32(flag,EXCEPTION)
	encrypted:=d._aes.AESEncrypt([]byte(cMsg))
	buf:=bytes.NewBuffer(flag)
	buf.Write(encrypted)
	return buf.Bytes()
}

func (d *Data) CreateRaw(uuid string,data []byte) []byte {
	flag:=make([]byte,4)
	binary.LittleEndian.PutUint32(flag,RAW)
	uuidLen:=make([]byte,4)
	binary.LittleEndian.PutUint32(uuidLen,uint32(len(uuid)))
	buf:=bytes.NewBuffer(uuidLen)
	buf.Write([]byte(uuid))
	buf.Write(data)
	encrypted:=d._aes.AESEncrypt(buf.Bytes())
	final:=bytes.NewBuffer(flag)
	final.Write(encrypted)
	return final.Bytes()
}

func (d *Data) CreateConnectSuccess(uuid string) []byte {
	flag:=make([]byte,4)
	binary.LittleEndian.PutUint32(flag,CONNECT_SUCCESS)
	data:=d._aes.AESEncrypt([]byte(uuid))
	final:=bytes.NewBuffer(flag)
	final.Write(data)
	return final.Bytes()
}

func (d *Data) ParseConnectSuccess(data []byte) string {
	uuid:=string(d._aes.AESDecrypt(data))
	return uuid
}

func (d *Data) ParseException(data []byte) *Exception {
	decrypted:=d._aes.AESDecrypt(data)
	_dec:=json.NewDecoder(strings.NewReader(string(decrypted)))
	exception:=Exception{}
	if err := _dec.Decode(&exception); err!=nil {
		fmt.Printf("err %v",err)
		return nil
	}else{
		return &exception
	}
}

func (d *Data) ParseRaw(data []byte) Raw {
	decrypted:=d._aes.AESDecrypt(data)
	uuidLen:=binary.LittleEndian.Uint32(decrypted)
	uuid:=string(decrypted[4:4+uuidLen])
	_data:=decrypted[4+uuidLen:]
	return Raw{Uuid:uuid,Data:_data}
}

func (d *Data) CreateDnsQuery(uuid string,host string) []byte {
	cMsg:=fmt.Sprintf("{\"uuid\":\"%v\",\"host\":\"%v\"}",uuid,host)
	flag:=make([]byte,4)
	binary.LittleEndian.PutUint32(flag,DNS)
	encrypted:=d._aes.AESEncrypt([]byte(cMsg))
	buf:=bytes.NewBuffer(flag)
	buf.Write(encrypted)
	return buf.Bytes()
}

func (d *Data) ParseDnsQuery(data []byte) *DnsQuery {
	decrypted:=d._aes.AESDecrypt(data)
	_dec:=json.NewDecoder(strings.NewReader(string(decrypted)))
	dns:=DnsQuery{}
	if err := _dec.Decode(&dns); err!=nil {
		fmt.Printf("err %v",err)
		return nil
	}else{
		return &dns
	}
}
