package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"

	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	device         string        = "enp0s8" //me estoy conectando por la interfaz del server
	snapshot_len   int32         = 1024
	promiscuous    bool          = false
	timeout        time.Duration = -1 * time.Second
	pcapFile       string        = "test.pcap"
	handle         *pcap.Handle
	err            error
	snapshotLenuuu uint32 = 1024
	//le puse otra variable por las compatibilidades que presenta con las funciones de abajo
	packetCount int = 0
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

func main() {

	/* Abrimos el dispositivo */
	handle, err = pcap.OpenLive(device, snapshot_len, promiscuous, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	/* Seteamos el filtro de captura */
	var filter string = "ether dst 00:1b:24:3e:0b:d3"
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Only capturing MAC dst 00:1b:24:3e:0b:d3 packets.")

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
			p.Length = packet.Metadata().CaptureLength
			p.Timestamp = packet.Metadata().Timestamp
		}

		/* Internet */
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			p.SrcIp = ip.SrcIP.String()
			p.DstIp = ip.DstIP.String()
			p.ProtoTp = ip.Protocol.String()
		}

		/* Transporte */
		transportLayer := packet.TransportLayer()
		if transportLayer.LayerType() == layers.LayerTypeTCP {
			tcp, _ := transportLayer.(*layers.TCP)
			p.SrcTp = tcp.SrcPort.String()
			p.DstTp = tcp.DstPort.String()
			p.ProtoApp = tcp.NextLayerType().String()
		}

		if transportLayer.LayerType() == layers.LayerTypeUDP {
			udp, _ := transportLayer.(*layers.UDP)
			p.SrcTp = udp.SrcPort.String()
			p.DstTp = udp.DstPort.String()
			p.ProtoApp = udp.NextLayerType().String()
		}

		/* Configuracion para insertar en la BD */
		client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
		//	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://juantuc98:juantuc98@db-wimp.yeslm.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))

		if err != nil {
			log.Fatal(err)
		}
		/* 		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) */
		ctx := context.Background()
		err = client.Connect(ctx)
		db := client.Database("wimp")
		col := db.Collection("paquetes")
		if err != nil {
			log.Fatal(err)
		}
		result, err := col.InsertOne(ctx, p)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(result)
	}
}
