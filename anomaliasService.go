package main

import (
	"fmt"
	"log"
	"time"

	"github.com/unpoller/unifi"
)

type anomalia struct {
	Datetime   time.Time
	SourceName string
	SiteName   string
	Anomaly    string
	DeviceMAC  string
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

	// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://juantuc98:juantuc98@db-wimp.yeslm.mongodb.net/myFirstDatabase?retryWrites=true&w=majority"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// ctx := context.Background()
	// err = client.Connect(ctx)
	// db := client.Database("wimp")
	// col := db.Collection("alertas")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	for {

		sites, err := uni.GetSites()
		if err != nil {
			log.Fatalln("Error:", err)
		}

		anomalias, err := uni.GetAnomalies(sites)
		if err != nil {
			log.Fatalln("Error:", err)
		}

		/* Configuracion para insertar en la BD */
		//			client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))

		for i := 0; i < len(anomalias); i++ {
			var a anomalia
			a.Datetime = anomalias[i].Datetime
			a.SourceName = anomalias[i].SourceName
			a.SiteName = anomalias[i].SiteName
			a.Anomaly = anomalias[i].Anomaly
			a.DeviceMAC = anomalias[i].DeviceMAC

			// result, err := col.InsertOne(ctx, a)
			// if err != nil {
			// 	fmt.Println(err)
			// }
			// fmt.Println(result)

			fmt.Println(a)
		}
		time.Sleep(60 * time.Second)
	}
	//Cerrar coneccion mongo
	//err = client.Disconnect(ctx)

}
