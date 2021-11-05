package main

import (
	"context"
	"log"
	"time"

	"github.com/unpoller/unifi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Dispositivo struct {
	CPU       int       `bson:"cpu,omitempty,minsize"`
	Temp      int       `bson:"temp,omitempty,minsize"`
	Uptime    int       `bson:"uptime,omitempty,minsize"`
	FanLevel  int       `bson:"fanlevel,omitempty,minsize"`
	Mem       int       `bson:"mem,omitempty,minsize"`
	Ip        string    `bson:"ip,omitempty,minsize"`
	MAC       string    `bson:"mac,omitempty,minsize"`
	Model     string    `bson:"model,omitempty,minsize"`
	SwMac     string    `bson:"swmac,omitempty,minsize"`
	Name      string    `bson:"name,omitempty,minsize"`
	Version   string    `bson:"version,omitempty,minsize"`
	Type      string    `bson:"type,omitempty,minsize"`
	Timestamp time.Time `bson:"timestamp,omitempty,minsize"`
}

func main() {
	c := unifi.Config{
		User: "wimpuser",
		Pass: "wimp.2021",
		URL:  "https://unifi.unt.edu.ar:8443",
	}
	uni, err := unifi.NewUnifi(&c)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	/* Configuracion para insertar en la BD */
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://juantuc98:juantuc98@db-wimp.yeslm.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	//client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	db := client.Database("wimp")
	col := db.Collection("dispositivos")
	opts := options.Update().SetUpsert(true)

	for {
		sites, err := uni.GetSites()
		if err != nil {
			log.Fatalln("Error:", err)
		}

		devices, err := uni.GetDevices(sites)
		if err != nil {
			log.Fatalln("Error:", err)
		}
		clients, err := uni.GetClients(sites)
		if err != nil {
			log.Fatalln("Error:", err)
		}

		for i := 0; i < len(devices.USWs); i++ {
			var s Dispositivo
			s.CPU = int(devices.USWs[i].SystemStats.CPU.Val)
			s.Temp = int(devices.USWs[i].GeneralTemperature.Val)
			s.Uptime = int(devices.USWs[i].SystemStats.Uptime.Val)
			s.FanLevel = int(devices.USWs[i].FanLevel.Val)
			s.Mem = int(devices.USWs[i].SystemStats.Mem.Val)
			//s.Mem = int((float32(devices.USWs[i].SysStats.MemUsed.Val) / float32(devices.USWs[i].SysStats.MemTotal.Val)) * 100)
			s.Ip = devices.USWs[i].IP
			s.MAC = devices.USWs[i].Mac
			s.Model = devices.USWs[i].Model
			s.Name = devices.USWs[i].Name
			s.Version = devices.USWs[i].Version
			s.Type = "SW"
			s.Timestamp = time.Now()

			filter := bson.M{
				"mac": s.MAC}

			update := bson.M{
				"$set": bson.M{
					"cpu":       s.CPU,
					"temp":      s.Temp,
					"uptime":    s.Uptime,
					"fanlevel":  s.FanLevel,
					"mem":       s.Mem,
					"ip":        s.Ip,
					"mac":       s.MAC,
					"model":     s.Model,
					"name":      s.Name,
					"version":   s.Version,
					"type":      s.Type,
					"timestamp": s.Timestamp}}
			result, err := col.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Fatal(err)
			}

			log.Println(result)
		}

		for j := 0; j < len(devices.UAPs); j++ {
			var ap Dispositivo
			ap.MAC = devices.UAPs[j].Mac
			ap.Ip = devices.UAPs[j].IP
			ap.Name = devices.UAPs[j].Name
			ap.Model = devices.UAPs[j].Model
			ap.Uptime = int(devices.UAPs[j].Uptime.Val)
			ap.CPU = int(devices.UAPs[j].SystemStats.CPU.Val)
			ap.Mem = int((float32(devices.UAPs[j].SysStats.MemUsed.Val) / float32(devices.UAPs[j].SysStats.MemTotal.Val)) * 100)
			ap.Type = "AP"
			ap.Timestamp = time.Now()
			ap.SwMac = devices.UAPs[j].LastUplink.UplinkMac

			filter := bson.M{
				"mac": ap.MAC}
			update := bson.M{
				"$set": bson.M{
					"cpu":       ap.CPU,
					"uptime":    ap.Uptime,
					"mem":       ap.Mem,
					"ip":        ap.Ip,
					"mac":       ap.MAC,
					"model":     ap.Model,
					"name":      ap.Name,
					"swmac":     ap.SwMac,
					"type":      ap.Type,
					"timestamp": ap.Timestamp}}

			result, err := col.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Fatal(err)
			}

			log.Println(result)
		}

		for k := 0; k < len(clients); k++ {
			var cl Dispositivo
			cl.Name = clients[k].Name
			cl.MAC = clients[k].Mac
			cl.Ip = clients[k].IP
			cl.Type = "CL"
			cl.Timestamp = time.Now()

			filter := bson.M{
				"mac": cl.MAC}
			update := bson.M{
				"$set": bson.M{
					"ip":        cl.Ip,
					"mac":       cl.MAC,
					"name":      cl.Name,
					"type":      cl.Type,
					"timestamp": cl.Timestamp}}

			result, err := col.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(result)
		}

		time.Sleep(60 * time.Second)
	}
	client.Disconnect(ctx)
}
