package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/gopacket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	device         string        = "enp0s8" //me estoy conectando por la interfaz del server
	snapshot_len   int32         = 1024
	promiscuous    bool          = true
	timeout        time.Duration = 10 * time.Second
	handle         *pcap.Handle
	err            error
	snapshotLenuuu uint32 = 1024
	//le puse otra variable por las compatibilidades que presenta con las funciones de abajo
)

type paquete struct {
	SrcMac    string    `bson:"srcmac,omitempty,minsize"`
	DstMac    string    `bson:"dstmac,omitempty,minsize"`
	ProtoIp   string    `bson:"protoip,omitempty,minsize"`
	SrcIp     string    `bson:"srcip,omitempty,minsize"`
	DstIp     string    `bson:"dstip,omitempty,minsize"`
	ProtoTp   string    `bson:"prototp,omitempty,minsize"`
	SrcTp     string    `bson:"srctp,omitempty,minsize"`
	DstTp     string    `bson:"dsttp,omitempty,minsize"`
	ProtoApp  string    `bson:"protoapp,omitempty,minsize"`
	Length    int       `bson:"length,omitempty,minsize"`
	Timestamp time.Time `bson:"timestamp,omitempty,minsize"`
}

func match(s string) string {
	i := strings.Index(s, "(")
	if i >= 0 {
		j := strings.Index(s[i:], ")")
		if j >= 0 {
			return s[i+1 : j+i]
		}
	}
	return ""
}

func main() {

	/* Abrimos el dispositivo */
	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	/* Seteamos el filtro de captura */
	var filter string = "ether dst 78:8a:20:47:8c:62"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Capturando paquetes con MAC dst 78:8a:20:47:8c:62")

	/* Configuracion para insertar en la BD */
	//client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://juantuc98:juantuc98@db-wimp.yeslm.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))

	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	db := client.Database("wimp")
	col := db.Collection("paquetes")
	if err != nil {
		log.Fatal(err)
	}

	/* Iteramos en las capas de los paquetes para sacar informacion. */
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		var p paquete
		/* Acceso a la red */
		ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethernetLayer != nil {
			ethernetPacket, _ := ethernetLayer.(*layers.Ethernet)
			p.SrcMac = ethernetPacket.SrcMAC.String()
			p.DstMac = ethernetPacket.DstMAC.String()
			p.ProtoIp = ethernetPacket.EthernetType.String()
			if p.ProtoIp != "IPv4" && p.ProtoIp != "IPv6" {
				/* Por la VLAN que usa Dot1Q */
				dotlayer := packet.Layer(layers.LayerTypeDot1Q)
				if dotlayer != nil {
					dot, _ := dotlayer.(*layers.Dot1Q)
					if dot.Type.String() == "IPv6" || dot.Type.String() == "IPv4" {
						p.ProtoIp = dot.Type.String()
					}
				}
			}
			p.Length = packet.Metadata().CaptureLength
			p.Timestamp = packet.Metadata().Timestamp
		}

		if p.ProtoIp == "IPv4" {
			/* Internet */
			ipLayer := packet.Layer(layers.LayerTypeIPv4)
			if ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv4)
				p.SrcIp = ip.SrcIP.String()
				p.DstIp = ip.DstIP.String()
				if ip.NextLayerType().String() == "TCP" || ip.NextLayerType().String() == "UDP" {
					p.ProtoTp = ip.NextLayerType().String()
				}

			}
		}

		if p.ProtoIp == "IPv6" {
			ipLayer := packet.Layer(layers.LayerTypeIPv6)
			if ipLayer != nil {
				ip, _ := ipLayer.(*layers.IPv6)
				p.SrcIp = ip.SrcIP.String()
				p.DstIp = ip.DstIP.String()
				if ip.NextLayerType().String() == "TCP" || ip.NextLayerType().String() == "UDP" {
					p.ProtoTp = ip.NextLayerType().String()
				}
			}
		}

		/* Transporte */
		transportLayer := packet.TransportLayer()
		if transportLayer != nil {
			if transportLayer.LayerType() == layers.LayerTypeTCP {
				tcp, _ := transportLayer.(*layers.TCP)
				p.SrcTp = tcp.SrcPort.String()
				p.DstTp = tcp.DstPort.String()
				if tcp.DstPort.String() == "443(https)" || tcp.SrcPort.String() == "443(https)" {
					p.ProtoApp = "https"
				} else if tcp.NextLayerType().String() == "Payload" || tcp.NextLayerType().String() == "" {
					p.ProtoApp = match(p.DstTp)
					if p.ProtoApp == "Payload" || p.ProtoApp == "" {
						p.ProtoApp = match(p.SrcTp)
					}
				} else {
					p.ProtoApp = tcp.NextLayerType().String()
				}
			}

			if transportLayer.LayerType() == layers.LayerTypeUDP {
				udp, _ := transportLayer.(*layers.UDP)
				if udp.NextLayerType().String() == "Payload" || udp.NextLayerType().String() == "" {
					p.ProtoApp = match(p.DstTp)
					if p.ProtoApp == "Payload" || p.ProtoApp == "" {
						p.ProtoApp = match(p.SrcTp)
					}
				} else {
					p.ProtoApp = udp.NextLayerType().String()
				}
			}
		}

		result, err := col.InsertOne(ctx, p)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(result)
	}
}
