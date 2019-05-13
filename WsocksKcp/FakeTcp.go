package WsocksKcp

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"time"
)

func SendFakeTcp(srcIp,dstIp string,srcPort,dstPort int,data []byte){
	ip := &layers.IPv4{
		SrcIP:    net.ParseIP(srcIp),
		DstIP:    net.ParseIP(dstIp),
		Protocol: layers.IPProtocolTCP,
	}
	tcp := &layers.TCP{
		SrcPort: layers.TCPPort(srcPort),
		DstPort: layers.TCPPort(dstPort),
		Seq:     1105024978,
		SYN:     true,
		Window:  14600,
	}
	_ = tcp.SetNetworkLayerForChecksum(ip)
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	if err := gopacket.SerializeLayers(buf, opts, tcp,gopacket.Payload(data)); err != nil {
		log.Fatal(err)
	}
	handle, err := pcap.OpenLive(
		"en0",	// device
		int32(65535),
		false,
		100 * time.Millisecond,
	)
	if err != nil {
		fmt.Println("Open handle error", err.Error())
	}
	defer handle.Close()
	//send
	if err := handle.WritePacketData(buf.Bytes()); err != nil {
		fmt.Println("Send error", err.Error())
	}

}