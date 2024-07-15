package main

import (
	influxdb2 "github.com/influxdata/influxdb-client-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nidotls/meineschufa-exporter/collector"
	"log"
	"os"
	"time"
)

func main() {
	app, err := collector.NewSmsApp()

	if err != nil {
		log.Fatal(err)
	}

	schufaApp, err := collector.NewSchufaApp(app)

	if err != nil {
		log.Fatal(err)
	}

	client := influxdb2.NewClient(os.Getenv("INFLUXDB_URL"), os.Getenv("INFLUXDB_TOKEN"))
	writeApi := client.WriteApi(os.Getenv("INFLUXDB_ORG"), os.Getenv("INFLUXDB_BUCKET"))

	for {
		log.Println("Getting score...")
		score, err := schufaApp.GetScore()

		if err != nil {
			log.Fatal(err)
		}

		types := make(map[string]int)
		categories := make(map[string]int)

		log.Println("Score:", score.Score)

		for _, entity := range score.Datalist {
			if _, ok := types[entity.Type]; ok {
				types[entity.Type]++
			} else {
				types[entity.Type] = 1
			}

			if _, ok := categories[entity.Category]; ok {
				categories[entity.Category]++
			} else {
				categories[entity.Category] = 1
			}
		}

		writeApi.WritePoint(influxdb2.NewPointWithMeasurement("score").
			AddField("value", score.Score).
			AddField("data", len(score.Datalist)).
			SetTime(time.Now()))

		for key, value := range types {
			writeApi.WritePoint(influxdb2.NewPointWithMeasurement("types").
				AddTag("type", key).
				AddField("value", value).
				SetTime(time.Now()))
		}
		for key, value := range categories {
			writeApi.WritePoint(influxdb2.NewPointWithMeasurement("categories").
				AddTag("type", key).
				AddField("value", value).
				SetTime(time.Now()))
		}

		time.Sleep(1 * time.Hour)
	}
}
