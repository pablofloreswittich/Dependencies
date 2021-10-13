package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/unpoller/unifi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Switch struct {
	CPU       int       `bson:"cpu,omitempty,minsize"`
	Temp      int       `bson:"temp,omitempty,minsize"`
	Uptime    int       `bson:"uptime,omitempty,minsize"`
	FanLevel  int       `bson:"fanlevel,omitempty,minsize"`
	Mem       int       `bson:"mem,omitempty,minsize"`
	Ip        string    `bson:"ip,omitempty,minsize"`
	MAC       string    `bson:"mac,omitempty,minsize"`
	Model     string    `bson:"model,omitempty,minsize"`
	Name      string    `bson:"name,omitempty,minsize"`
	Version   string    `bson:"version,omitempty,minsize"`
	Timestamp time.Time `bson:"timestamp,omitempty,minsize"`
}

func main() {
	c := unifi.Config{
		User: "wimpuser",
		Pass: "wimp.2021",
		URL:  "https://unifi.unt.edu.ar:8443",
		// Log with log.Printf or make your own interface that accepts (msg, fmt)
		/* ErrorLog: log.Printf, */
		/* DebugLog: log.Printf, */
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
	/* ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) */
	ctx := context.Background()
	err = client.Connect(ctx)
	db := client.Database("wimp")
	col := db.Collection("switches")
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

		for i := 0; i < len(devices.USWs); i++ {
			var s Switch
			s.CPU = int(devices.USWs[i].SystemStats.CPU.Val)
			s.Temp = int(devices.USWs[i].GeneralTemperature.Val)
			s.Uptime = int(devices.USWs[i].SystemStats.Uptime.Val)
			s.FanLevel = int(devices.USWs[i].FanLevel.Val)
			s.Mem = int((float32(devices.USWs[0].SysStats.MemUsed.Val) / float32(devices.USWs[0].SysStats.MemTotal.Val)) * 100)
			s.Ip = devices.USWs[i].IP
			s.MAC = devices.USWs[i].Mac
			s.Model = devices.USWs[i].Model
			s.Name = devices.USWs[i].Name
			s.Version = devices.USWs[i].Version
			s.Timestamp = time.Now()
			fmt.Println(s)
			filter := bson.D{{"mac", s.MAC}}
			update := bson.D{
				{"$set",
					bson.D{
						{"cpu", s.CPU},
						{"temp", s.Temp},
						{"uptime", s.Uptime},
						{"fanlevel", s.FanLevel},
						{"mem", s.Mem},
						{"ip", s.Ip},
						{"mac", s.MAC},
						{"model", s.Model},
						{"name", s.Name},
						{"version", s.Version},
						{"timestamp", s.Timestamp},
					},
				},
			}
			result, err := col.UpdateOne(ctx, filter, update, opts)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(result)
		}

		time.Sleep(60 * time.Second)
	}
	client.Disconnect(ctx)

}
