package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/unpoller/unifi"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type anomalia struct {
	Datetime  time.Time `bson:"timestamp,omitempty,minsize"`
	Anomaly   string    `bson:"anomaly,omitempty,minsize"`
	DeviceMAC string    `bson:"mac,omitempty,minsize"`
}

func main() {
	c := unifi.Config{
		User: "wimpuser",
		Pass: "wimp.2021",
		URL:  "https://unifi.unt.edu.ar:8443/",
		// Log with log.Printf or make your own interface that accepts (msg, fmt)
		//ErrorLog: log.Printf,
		//DebugLog: log.Printf,
	}
	uni, err := unifi.NewUnifi(&c)
	if err != nil {
		log.Fatalln("Error:", err)
	}
	/* Configuracion para insertar en la BD */
	//			client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://juantuc98:juantuc98@db-wimp.yeslm.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	db := client.Database("wimp")
	col := db.Collection("anomalias")
	if err != nil {
		log.Fatal(err)
	}

	for {
		sites, err := uni.GetSites()
		if err != nil {
			log.Fatalln("Error:", err)
		}

		anomalias, err := uni.GetAnomalies(sites)
		if err != nil {
			log.Fatalln("Error:", err)
		}

		tiempoActual := time.Now()
		politicaInsercion := tiempoActual.AddDate(0, -1, 0)

		for i := 0; i < len(anomalias); i++ {
			if anomalias[i].Datetime.After(politicaInsercion) {
				var a anomalia
				a.Datetime = anomalias[i].Datetime
				a.Anomaly = anomalias[i].Anomaly
				a.DeviceMAC = anomalias[i].DeviceMAC
				result, err := col.InsertOne(ctx, a)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(result)
			}
		}
		time.Sleep(60 * time.Second)
	}
	err = client.Disconnect(ctx)

}
