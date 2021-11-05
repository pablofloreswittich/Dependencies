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

type AP struct {
	CPU       int       `bson:"cpu,omitempty,minsize"`
	Uptime    int       `bson:"uptime,omitempty,minsize"`
	Mem       int       `bson:"mem,omitempty,minsize"`
	Ip        string    `bson:"ip,omitempty,minsize"`
	MAC       string    `bson:"mac,omitempty,minsize"`
	Model     string    `bson:"model,omitempty,minsize"`
	Name      string    `bson:"name,omitempty,minsize"`
	SwMac	  string    `bson:"swmac,omitempty,minsize"`
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
	col := db.Collection("aps")
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

		for i := 0; i < len(devices.UAPs); i++ {
			var ap AP
			ap.MAC = devices.UAPs[i].Mac
			ap.Ip = devices.UAPs[i].IP
			ap.Name = devices.UAPs[i].Name
			ap.Model = devices.UAPs[i].Model
			ap.Uptime = int(devices.UAPs[i].Uptime.Val)
			ap.CPU = int(devices.UAPs[i].SystemStats.CPU.Val)
			ap.Mem = int((float32(devices.UAPs[i].SysStats.MemUsed.Val) / float32(devices.UAPs[i].SysStats.MemTotal.Val)) * 100)
			ap.Timestamp = time.Now()
			ap.SwMac = devices.UAPs[i].LastUplink.UplinkMac
			filter := bson.D{{"mac", ap.MAC}}
			update := bson.D{
				{"$set",
					bson.D{
						{"cpu", ap.CPU},
						{"uptime", ap.Uptime},
						{"mem", ap.Mem},
						{"ip", ap.Ip},
						{"mac", ap.MAC},
						{"model", ap.Model},
						{"name", ap.Name},
						{"swmac", ap.SwMac},
						{"timestamp", ap.Timestamp},
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
