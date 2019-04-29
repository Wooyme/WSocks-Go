package Wsocks

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io/ioutil"
)

type Aes struct {
	iv []byte
	key []byte
	_zlib bool

}

func NewAes(key []byte,_zlib bool) Aes {
	myAes:=Aes{}
	raw:=bytes.NewBuffer(key)
	if len(key)<16 {
		tmp:=make([]byte,16-len(key))
		for i:=range tmp {
			tmp[i] = 0x06
		}
		raw.Write(tmp)
	}
	myAes._zlib = _zlib
	myAes.key = raw.Bytes()
	myAes.iv=[]byte{0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05}
	return  myAes
}

func (myAes *Aes) AESEncrypt(src []byte) []byte {
	var data []byte
	block, err := aes.NewCipher(myAes.key)
	if err != nil {
		fmt.Println("key error1", err)
	}
	ecb := cipher.NewCBCEncrypter(block, myAes.iv)
	if myAes._zlib {
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(src)
		w.Close()
		data = b.Bytes()
	}else{
		data = src
	}
	content := data
	content = PKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	return crypted
}

func (myAes *Aes)AESDecrypt(crypt []byte) []byte {
	block, err := aes.NewCipher(myAes.key)
	if err != nil {
		fmt.Println("key error", err)
	}
	ecb := cipher.NewCBCDecrypter(block, myAes.iv)
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)
	data:=PKCS5Trimming(decrypted)
	if myAes._zlib {
		b := bytes.NewBuffer(data)
		r, err := zlib.NewReader(b)
		if err != nil {
			fmt.Printf("Error %v \n", err)
			return nil
		}
		_bytes, err := ioutil.ReadAll(r)
		if err != nil {
			fmt.Printf("Error %v \n", err)
			return nil
		}
		r.Close()
		return _bytes
	}else{
		return data
	}
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}