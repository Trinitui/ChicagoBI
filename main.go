package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/kelvins/geocoder"
	_ "github.com/lib/pq"
)

type CCVIJsonRecords []struct {
	Geography_type        string `json:"geography_type"`
	Community_area_or_zip string `json:"community_area_or_zip"`
	Community_area_name   string `json:"community_area_name"`
	Ccvi_score            string `json:"ccvi_score"`
	Ccvi_category         string `json:"ccvi_category"`
}

type CovidDataJsonRecords []struct {
	Zip_code                       string `json:"zip_code"`
	Week_number                    string `json:"week_number"`
	Week_end                       string `json:"week_end"`
	Cases_weekly                   string `json:"cases_weekly"`
	Tests_weekly                   string `json:"tests_weekly"`
	Deaths_weekly                  string `json:"deaths_weekly"`
	Percent_tested_positive_weekly string `json:"percent_tested_positive_weekly"`
}

type BuildingPermitsJsonRecords []struct {
	Permit_id         string `json:"id"`
	Permit_issue_date string `json:"issue_date"`
	Community_area    string `json:"community_area"`
}

type TaxiTripsJsonRecords []struct {
	Trip_id                    string `json:"trip_id"`
	Trip_start_timestamp       string `json:"trip_start_timestamp"`
	Trip_end_timestamp         string `json:"trip_end_timestamp"`
	Pickup_centroid_latitude   string `json:"pickup_centroid_latitude"`
	Pickup_centroid_longitude  string `json:"pickup_centroid_longitude"`
	Dropoff_centroid_latitude  string `json:"dropoff_centroid_latitude"`
	Dropoff_centroid_longitude string `json:"dropoff_centroid_longitude"`
}

type TransportTripsJsonRecords []struct {
	Trip_id                    string `json:"trip_id"`
	Trip_start_timestamp       string `json:"trip_start_timestamp"`
	Trip_end_timestamp         string `json:"trip_end_timestamp"`
	Pickup_centroid_latitude   string `json:"pickup_centroid_latitude"`
	Pickup_centroid_longitude  string `json:"pickup_centroid_longitude"`
	Dropoff_centroid_latitude  string `json:"dropoff_centroid_latitude"`
	Dropoff_centroid_longitude string `json:"dropoff_centroid_longitude"`
}

type UnemploymentJsonRecords []struct {
	Community_area string `json:"community_area"`
	Unemployment   string `json:"unemployment"`
}

var db *sql.DB

func init() {
	var err error

	fmt.Println("Initializing the DB connection")

	// Establish connection to Postgres Database

	// OPTION 1 - Postgress application running on localhost
	//db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=localhost sslmode=disable port = 5432"

	// OPTION 2
	// Docker container for the Postgres microservice - uncomment when deploy with host.docker.internal
	//db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=host.docker.internal sslmode=disable port = 5433"

	// OPTION 3
	// Docker container for the Postgress microservice - uncomment when deploy with IP address of the container
	// To find your Postgres container IP, use the command with your network name listed in the docker compose file as follows:
	// docker network inspect cbi_backend
	//db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=162.123.0.9 sslmode=disable port = 5433"

	//Option 4
	//Database application running on Google Cloud Platform.
	db_connection := "user=postgres dbname=chicago_business_intelligence password=root host=/cloudsql/finalproject-p3:us-central1:mypostgres sslmode=disable port = 5432"

	db, err = sql.Open("postgres", db_connection)
	if err != nil {
		log.Fatal(fmt.Println("Couldn't Open Connection to database"))
		panic(err)
	}

	// Test the database connection
	//err = db.Ping()
	//if err != nil {
	//	fmt.Println("Couldn't Connect to database")
	//	panic(err)
	//}

}

func main() {
	log.Println("Starting MicroServices!")

	// Spin in a loop and pull data from the city of chicago data portal
	// Once every hour, day, week, etc.
	// Though, please note that Not all datasets need to be pulled on daily basis
	// fine-tune the following code-snippet as you see necessary
	for {
		// build and fine-tune functions to pull data from different data sources
		// This is a code snippet to show you how to pull data from different data sources.

		http.HandleFunc("/", handler)

		// Determine port for HTTP service.
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
			log.Printf("defaulting to port %s", port)
		}

		// Start HTTP server.
		log.Printf("listening on port %s", port)
		log.Print("Navigate to Cloud Run services and find the URL of your service")
		log.Print("Use the browser and navigate to your service URL to to check your service has started")

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}

		go GetTransportTrips(db)
		go GetTaxiTrips(db)
		go GetCCVI(db)
		go GetCovidData(db)
		go GetBuildingPermits(db)
		go GetUnemploymentData(db)

		time.Sleep(24 * time.Hour)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("PROJECT_ID")
	if name == "" {
		name = "CBI-Project"
	}

	fmt.Fprintf(w, "CBI data collection microservices' goroutines have started for %s!\n", name)
}

func GetCCVI(db *sql.DB) {

	// This function is NOT complete
	// It provides code-snippets for the data source: https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// You need to complete the implmentation and add the data source: https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	// Data Collection needed from two data sources:
	// 1. https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// 2. https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	fmt.Println("GetCCVI: Grabbing CCVI Data")

	// Get your geocoder.ApiKey from here :
	// https://developers.google.com/maps/documentation/geocoding/get-api-key?authuser=2

	geocoder.ApiKey = "AIzaSyB-JwmMaEwb3yEomj66SnNlkA5GyKcRfWU"

	drop_table := `drop table if exists ccvi`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "ccvi" (
						"id"   SERIAL , 
						"geography_type" VARCHAR(255), 
						"community_area_or_zip" VARCHAR(255),
						"community_area_name" VARCHAR(255),
						"ccvi_score" FLOAT, 
						"ccvi_category" VARCHAR(255), 
						PRIMARY KEY ("id") 
					);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	// While doing unit-testing keep the limit value to 500
	// later you could change it to 1000, 2000, 10,000, etc.
	fmt.Println("Grabbing data from Chicago Data...")
	var url = "https://data.cityofchicago.org/resource/xhc6-88s9.json"

	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     300 * time.Second,
		DisableCompression:  true,
		TLSHandshakeTimeout: 1000 * time.Second,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting JSON unmarshalling...")
	body, _ := ioutil.ReadAll(res.Body)
	var ccvi_list CCVIJsonRecords
	json.Unmarshal(body, &ccvi_list)
	fmt.Println("JSON unmarshalling done...")
	fmt.Println("Now unpacking JSON and inserting into db... ")
	for i := 0; i < len(ccvi_list); i++ {

		// We will execute definsive coding to check for messy/dirty/missing data values
		// Any record that has messy/dirty/missing data we don't enter it in the data lake/table

		geo_type := ccvi_list[i].Geography_type
		if geo_type == "" {
			continue
		}

		// if trip start/end timestamp doesn't have the length of 23 chars in the format "0000-00-00T00:00:00.000"
		// skip this record

		// get Trip_start_timestamp
		community_zip_ca := ccvi_list[i].Community_area_or_zip
		if community_zip_ca == "" {
			continue
		}

		// get Trip_end_timestamp
		community_name := ccvi_list[i].Community_area_name
		if community_name == "" {
			continue
		}

		ccvi_score := ccvi_list[i].Ccvi_score
		if ccvi_score == "" {
			continue
		}

		ccvi_category := ccvi_list[i].Ccvi_category
		if ccvi_category == "" {
			continue
		}

		sql := `INSERT INTO ccvi ("geography_type", "community_area_or_zip", "community_area_name", "ccvi_score", "ccvi_category"
			) values($1, $2, $3, $4, $5)`

		_, err = db.Exec(
			sql,
			geo_type,
			community_zip_ca,
			community_name,
			ccvi_score,
			ccvi_category)

		if err != nil {
			panic(err)
		}

	}
	fmt.Println("== Done with CCVI ==")

}

func GetCovidData(db *sql.DB) {

	// This function is NOT complete
	// It provides code-snippets for the data source: https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// You need to complete the implmentation and add the data source: https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	// Data Collection needed from two data sources:
	// 1. https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// 2. https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	fmt.Println("GetCovidData: Grabbing COVID Data")

	// Get your geocoder.ApiKey from here :
	// https://developers.google.com/maps/documentation/geocoding/get-api-key?authuser=2

	geocoder.ApiKey = "AIzaSyB-JwmMaEwb3yEomj66SnNlkA5GyKcRfWU"

	drop_table := `drop table if exists covid_data`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS covid_data (
						"id"   SERIAL , 
						"zip_code" VARCHAR(255), 
						"week_number" INT,
						"week_end" DATE,
						"cases_weekly" FLOAT,
						"tests_weekly" FLOAT, 
						"deaths_weekly" FLOAT, 
						"percent_tested_positive_weekly" FLOAT,
						PRIMARY KEY ("id") 
					);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	// While doing unit-testing keep the limit value to 500
	// later you could change it to 1000, 2000, 10,000, etc.
	fmt.Println("Grabbing data from Chicago Data...")
	var url = "https://data.cityofchicago.org/resource/yhhz-zm2v.json"

	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     300 * time.Second,
		DisableCompression:  true,
		TLSHandshakeTimeout: 1000 * time.Second,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting JSON unmarshalling...")
	body, _ := ioutil.ReadAll(res.Body)
	var covid_list CovidDataJsonRecords
	json.Unmarshal(body, &covid_list)
	fmt.Println("JSON unmarshalling done...")
	fmt.Println("Now unpacking JSON and inserting into db... ")
	for i := 0; i < len(covid_list); i++ {

		// We will execute definsive coding to check for messy/dirty/missing data values
		// Any record that has messy/dirty/missing data we don't enter it in the data lake/table

		zip_code := covid_list[i].Zip_code
		if zip_code == "" {
			continue
		}

		week_number := covid_list[i].Week_number
		if week_number == "" {
			continue
		}

		week_end := covid_list[i].Week_end
		if week_end == "" {
			continue
		}

		cases_weekly := covid_list[i].Cases_weekly
		if cases_weekly == "" {
			continue
		}

		tests_weekly := covid_list[i].Tests_weekly
		if tests_weekly == "" {
			continue
		}

		deaths_weekly := covid_list[i].Deaths_weekly
		if deaths_weekly == "" {
			continue
		}

		percent_tested_positive_weekly := covid_list[i].Percent_tested_positive_weekly
		if percent_tested_positive_weekly == "" {
			continue
		}

		sql := `INSERT INTO covid_data ("zip_code", "week_number", "week_end", "cases_weekly", "tests_weekly", "deaths_weekly", "percent_tested_positive_weekly"
			) values($1, $2, $3, $4, $5, $6, $7)`

		_, err = db.Exec(
			sql,
			zip_code,
			week_number,
			week_end,
			cases_weekly,
			tests_weekly,
			deaths_weekly,
			percent_tested_positive_weekly)

		if err != nil {
			panic(err)
		}

	}
	fmt.Println("== Done with Covid Data ==")

}

func GetBuildingPermits(db *sql.DB) {

	// This function is NOT complete
	// It provides code-snippets for the data source: https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// You need to complete the implmentation and add the data source: https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	// Data Collection needed from two data sources:
	// 1. https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// 2. https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	fmt.Println("GetBuildingPermits: Collecting Building Permits Data")

	// Get your geocoder.ApiKey from here :
	// https://developers.google.com/maps/documentation/geocoding/get-api-key?authuser=2

	geocoder.ApiKey = "AIzaSyB-JwmMaEwb3yEomj66SnNlkA5GyKcRfWU"

	drop_table := `drop table if exists building_permits`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "building_permits" (
						"id"   SERIAL , 
						"permit_id" VARCHAR(255), 
						"permit_issue_date" DATE, 
						"community_area" INT, 
						PRIMARY KEY ("id") 
					);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	// While doing unit-testing keep the limit value to 500
	// later you could change it to 1000, 2000, 10,000, etc.
	fmt.Println("Grabbing data from Chicago Data...")
	var url = "https://data.cityofchicago.org/resource/ydr8-5enu.json"

	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     300 * time.Second,
		DisableCompression:  true,
		TLSHandshakeTimeout: 1000 * time.Second,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting JSON unmarshalling...")
	body, _ := ioutil.ReadAll(res.Body)
	var building_permits_list BuildingPermitsJsonRecords
	json.Unmarshal(body, &building_permits_list)
	fmt.Println("JSON unmarshalling done...")
	fmt.Println("Now unpacking JSON and inserting into db... ")
	for i := 0; i < len(building_permits_list); i++ {

		// We will execute definsive coding to check for messy/dirty/missing data values
		// Any record that has messy/dirty/missing data we don't enter it in the data lake/table

		permit_id := building_permits_list[i].Permit_id
		if permit_id == "" {
			continue
		}

		// if trip start/end timestamp doesn't have the length of 23 chars in the format "0000-00-00T00:00:00.000"
		// skip this record

		// get Trip_start_timestamp
		permit_issue_date := building_permits_list[i].Permit_issue_date
		if permit_issue_date == "" {
			continue
		}

		// get Trip_end_timestamp
		community_area := building_permits_list[i].Community_area
		if community_area == "" {
			continue
		}

		sql := `INSERT INTO building_permits ("permit_id", "permit_issue_date", "community_area") values($1, $2, $3)`

		_, err = db.Exec(
			sql,
			permit_id,
			permit_issue_date,
			community_area,
		)

		if err != nil {
			panic(err)
		}

	}
	fmt.Println("== Done with inserting Building Permits ==")

}

func GetTaxiTrips(db *sql.DB) int {

	// This function is NOT complete
	// It provides code-snippets for the data source: https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// You need to complete the implmentation and add the data source: https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	// Data Collection needed from two data sources:
	// 1. https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// 2. https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	fmt.Println("GetTaxiTrips: Collecting Taxi Trips Data")

	// Get your geocoder.ApiKey from here :
	// https://developers.google.com/maps/documentation/geocoding/get-api-key?authuser=2

	geocoder.ApiKey = "AIzaSyB-JwmMaEwb3yEomj66SnNlkA5GyKcRfWU"

	drop_table := `drop table if exists taxi_trips`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "taxi_trips" (
						"id"   SERIAL , 
						"trip_id" VARCHAR(255), 
						"trip_start_timestamp" TIMESTAMP WITH TIME ZONE, 
						"trip_end_timestamp" TIMESTAMP WITH TIME ZONE, 
						"pickup_centroid_latitude" DOUBLE PRECISION, 
						"pickup_centroid_longitude" DOUBLE PRECISION, 
						"dropoff_centroid_latitude" DOUBLE PRECISION, 
						"dropoff_centroid_longitude" DOUBLE PRECISION, 
						"pickup_zip_code" VARCHAR(255), 
						"dropoff_zip_code" VARCHAR(255), 
						PRIMARY KEY ("id") 
					);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	// While doing unit-testing keep the limit value to 500
	// later you could change it to 1000, 2000, 10,000, etc.
	fmt.Println("Grabbing data from Chicago Data...")
	var url = "https://data.cityofchicago.org/resource/wrvz-psew.json"

	tr := &http.Transport{
		MaxIdleConns:          10,
		IdleConnTimeout:       1000 * time.Second,
		TLSHandshakeTimeout:   1000 * time.Second,
		ExpectContinueTimeout: 1000 * time.Second,
		DisableCompression:    true,
		Dial: (&net.Dialer{
			Timeout:   1000 * time.Second,
			KeepAlive: 1000 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: 1000 * time.Second,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting JSON unmarshalling...")
	body, _ := ioutil.ReadAll(res.Body)
	var taxi_trips_list TaxiTripsJsonRecords
	json.Unmarshal(body, &taxi_trips_list)
	fmt.Println("JSON unmarshalling done...")
	fmt.Println("Now unpacking JSON and inserting into db... ")
	if len(taxi_trips_list) == 0 {
		fmt.Println("No Data in Taxi Trips")
		return 0
	}
	for i := 0; i < len(taxi_trips_list); i++ {

		//fmt.Println(taxi_trips_list[i])

		// We will execute definsive coding to check for messy/dirty/missing data values
		// Any record that has messy/dirty/missing data we don't enter it in the data lake/table

		trip_id := taxi_trips_list[i].Trip_id
		if trip_id == "" {
			continue
		}

		// if trip start/end timestamp doesn't have the length of 23 chars in the format "0000-00-00T00:00:00.000"
		// skip this record

		// get Trip_start_timestamp
		trip_start_timestamp := taxi_trips_list[i].Trip_start_timestamp
		if len(trip_start_timestamp) < 23 {
			continue
		}

		// get Trip_end_timestamp
		trip_end_timestamp := taxi_trips_list[i].Trip_end_timestamp
		if len(trip_end_timestamp) < 23 {
			continue
		}

		pickup_centroid_latitude := taxi_trips_list[i].Pickup_centroid_latitude

		if pickup_centroid_latitude == "" {
			continue
		}

		pickup_centroid_longitude := taxi_trips_list[i].Pickup_centroid_longitude
		//pickup_centroid_longitude := taxi_trips_list[i].PICKUP_LONG

		if pickup_centroid_longitude == "" {
			continue
		}

		dropoff_centroid_latitude := taxi_trips_list[i].Dropoff_centroid_latitude
		//dropoff_centroid_latitude := taxi_trips_list[i].DROPOFF_LAT

		if dropoff_centroid_latitude == "" {
			continue
		}

		dropoff_centroid_longitude := taxi_trips_list[i].Dropoff_centroid_longitude
		//dropoff_centroid_longitude := taxi_trips_list[i].DROPOFF_LONG

		if dropoff_centroid_longitude == "" {
			continue
		}

		// Using pickup_centroid_latitude and pickup_centroid_longitude in geocoder.GeocodingReverse
		// we could find the pickup zip-code

		pickup_centroid_latitude_float, _ := strconv.ParseFloat(pickup_centroid_latitude, 64)
		pickup_centroid_longitude_float, _ := strconv.ParseFloat(pickup_centroid_longitude, 64)
		pickup_location := geocoder.Location{
			Latitude:  pickup_centroid_latitude_float,
			Longitude: pickup_centroid_longitude_float,
		}

		pickup_address_list, _ := geocoder.GeocodingReverse(pickup_location)

		if len(pickup_address_list) == 0 {
			fmt.Printf("No results found for crash at latitude : %f and Longitude : %f \n", pickup_centroid_latitude_float, pickup_centroid_longitude_float)
			continue
		}

		pickup_address := pickup_address_list[0]
		pickup_zip_code := pickup_address.PostalCode

		// Using dropoff_centroid_latitude and dropoff_centroid_longitude in geocoder.GeocodingReverse
		// we could find the dropoff zip-code

		dropoff_centroid_latitude_float, _ := strconv.ParseFloat(dropoff_centroid_latitude, 64)
		dropoff_centroid_longitude_float, _ := strconv.ParseFloat(dropoff_centroid_longitude, 64)

		dropoff_location := geocoder.Location{
			Latitude:  dropoff_centroid_latitude_float,
			Longitude: dropoff_centroid_longitude_float,
		}

		dropoff_address_list, _ := geocoder.GeocodingReverse(dropoff_location)

		if len(dropoff_address_list) == 0 {
			fmt.Printf("No results found for crash at latitude : %f and Longitude : %f \n", dropoff_centroid_latitude_float, dropoff_centroid_longitude_float)
			continue
		}

		dropoff_address := dropoff_address_list[0]
		dropoff_zip_code := dropoff_address.PostalCode

		sql := `INSERT INTO taxi_trips ("trip_id", "trip_start_timestamp", "trip_end_timestamp", "pickup_centroid_latitude", "pickup_centroid_longitude", "dropoff_centroid_latitude", "dropoff_centroid_longitude", "pickup_zip_code", 
			"dropoff_zip_code") values($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = db.Exec(
			sql,
			trip_id,
			trip_start_timestamp,
			trip_end_timestamp,
			pickup_centroid_latitude,
			pickup_centroid_longitude,
			dropoff_centroid_latitude,
			dropoff_centroid_longitude,
			pickup_zip_code,
			dropoff_zip_code)

		if err != nil {
			panic(err)
		}

	}
	fmt.Println("== Done with Taxi Trips ==")
	return 1
}

func GetTransportTrips(db *sql.DB) {

	// This function is NOT complete
	// It provides code-snippets for the data source: https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// You need to complete the implmentation and add the data source: https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	// Data Collection needed from two data sources:
	// 1. https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// 2. https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	fmt.Println("GetTransportProviders: Collecting Transport Trips Data")

	// Get your geocoder.ApiKey from here :
	// https://developers.google.com/maps/documentation/geocoding/get-api-key?authuser=2

	geocoder.ApiKey = "AIzaSyB-JwmMaEwb3yEomj66SnNlkA5GyKcRfWU"

	drop_table := `drop table if exists transport_trips`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS "transport_trips" (
						"id"   SERIAL , 
						"trip_id" VARCHAR(255), 
						"trip_start_timestamp" TIMESTAMP WITH TIME ZONE, 
						"trip_end_timestamp" TIMESTAMP WITH TIME ZONE, 
						"pickup_centroid_latitude" DOUBLE PRECISION, 
						"pickup_centroid_longitude" DOUBLE PRECISION, 
						"dropoff_centroid_latitude" DOUBLE PRECISION, 
						"dropoff_centroid_longitude" DOUBLE PRECISION, 
						"pickup_zip_code" VARCHAR(255), 
						"dropoff_zip_code" VARCHAR(255), 
						PRIMARY KEY ("id") 
					);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	// While doing unit-testing keep the limit value to 500
	// later you could change it to 1000, 2000, 10,000, etc.
	fmt.Println("Grabbing data from Chicago Data...")
	var url = "https://data.cityofchicago.org/resource/m6dm-c72p.json"

	tr := &http.Transport{
		MaxIdleConns:          10,
		IdleConnTimeout:       1000 * time.Second,
		TLSHandshakeTimeout:   1000 * time.Second,
		ExpectContinueTimeout: 1000 * time.Second,
		DisableCompression:    true,
		Dial: (&net.Dialer{
			Timeout:   1000 * time.Second,
			KeepAlive: 1000 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: 1000 * time.Second,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Println("TRANSPORTS: Starting JSON unmarshalling...")
	body, _ := ioutil.ReadAll(res.Body)
	var transport_trips_list TransportTripsJsonRecords
	json.Unmarshal(body, &transport_trips_list)
	fmt.Println("TRANSPORTS: JSON unmarshalling done...")
	fmt.Println("TRANSPORTS: Now unpacking JSON and inserting into db... ")
	for i := 0; i < len(transport_trips_list); i++ {
		//fmt.Println(transport_trips_list[i])

		// We will execute definsive coding to check for messy/dirty/missing data values
		// Any record that has messy/dirty/missing data we don't enter it in the data lake/table

		trip_id := transport_trips_list[i].Trip_id
		if trip_id == "" {
			continue
		}

		// if trip start/end timestamp doesn't have the length of 23 chars in the format "0000-00-00T00:00:00.000"
		// skip this record

		// get Trip_start_timestamp
		trip_start_timestamp := transport_trips_list[i].Trip_start_timestamp
		if len(trip_start_timestamp) < 23 {
			continue
		}

		// get Trip_end_timestamp
		trip_end_timestamp := transport_trips_list[i].Trip_end_timestamp
		if len(trip_end_timestamp) < 23 {
			continue
		}

		pickup_centroid_latitude := transport_trips_list[i].Pickup_centroid_latitude

		if pickup_centroid_latitude == "" {
			continue
		}

		pickup_centroid_longitude := transport_trips_list[i].Pickup_centroid_longitude
		//pickup_centroid_longitude := taxi_trips_list[i].PICKUP_LONG

		if pickup_centroid_longitude == "" {
			continue
		}

		dropoff_centroid_latitude := transport_trips_list[i].Dropoff_centroid_latitude
		//dropoff_centroid_latitude := taxi_trips_list[i].DROPOFF_LAT

		if dropoff_centroid_latitude == "" {
			continue
		}

		dropoff_centroid_longitude := transport_trips_list[i].Dropoff_centroid_longitude
		//dropoff_centroid_longitude := taxi_trips_list[i].DROPOFF_LONG

		if dropoff_centroid_longitude == "" {
			continue
		}

		// Using pickup_centroid_latitude and pickup_centroid_longitude in geocoder.GeocodingReverse
		// we could find the pickup zip-code

		pickup_centroid_latitude_float, _ := strconv.ParseFloat(pickup_centroid_latitude, 64)
		pickup_centroid_longitude_float, _ := strconv.ParseFloat(pickup_centroid_longitude, 64)
		pickup_location := geocoder.Location{
			Latitude:  pickup_centroid_latitude_float,
			Longitude: pickup_centroid_longitude_float,
		}

		pickup_address_list, _ := geocoder.GeocodingReverse(pickup_location)

		if len(pickup_address_list) == 0 {
			fmt.Printf("No results found for crash at latitude : %f and Longitude : %f \n", pickup_centroid_latitude_float, pickup_centroid_longitude_float)
			continue
		}

		pickup_address := pickup_address_list[0]
		pickup_zip_code := pickup_address.PostalCode

		// Using dropoff_centroid_latitude and dropoff_centroid_longitude in geocoder.GeocodingReverse
		// we could find the dropoff zip-code

		dropoff_centroid_latitude_float, _ := strconv.ParseFloat(dropoff_centroid_latitude, 64)
		dropoff_centroid_longitude_float, _ := strconv.ParseFloat(dropoff_centroid_longitude, 64)

		dropoff_location := geocoder.Location{
			Latitude:  dropoff_centroid_latitude_float,
			Longitude: dropoff_centroid_longitude_float,
		}

		dropoff_address_list, _ := geocoder.GeocodingReverse(dropoff_location)

		if len(dropoff_address_list) == 0 {
			fmt.Printf("No results found for crash at latitude : %f and Longitude : %f \n", dropoff_centroid_latitude_float, dropoff_centroid_longitude_float)
			continue
		}

		dropoff_address := dropoff_address_list[0]
		dropoff_zip_code := dropoff_address.PostalCode

		sql := `INSERT INTO transport_trips ("trip_id", "trip_start_timestamp", "trip_end_timestamp", "pickup_centroid_latitude", "pickup_centroid_longitude", "dropoff_centroid_latitude", "dropoff_centroid_longitude", "pickup_zip_code", 
			"dropoff_zip_code") values($1, $2, $3, $4, $5, $6, $7, $8, $9)`

		_, err = db.Exec(
			sql,
			trip_id,
			trip_start_timestamp,
			trip_end_timestamp,
			pickup_centroid_latitude,
			pickup_centroid_longitude,
			dropoff_centroid_latitude,
			dropoff_centroid_longitude,
			pickup_zip_code,
			dropoff_zip_code)

		if err != nil {
			panic(err)
		}

	}
	fmt.Println("== Done with Transport Trips ==")

}

func GetUnemploymentData(db *sql.DB) {

	// This function is NOT complete
	// It provides code-snippets for the data source: https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// You need to complete the implmentation and add the data source: https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	// Data Collection needed from two data sources:
	// 1. https://data.cityofchicago.org/Transportation/Taxi-Trips/wrvz-psew
	// 2. https://data.cityofchicago.org/Transportation/Transportation-Network-Providers-Trips/m6dm-c72p

	fmt.Println("GetUnemploymentData: Grabbing Unemployment Data")

	// Get your geocoder.ApiKey from here :
	// https://developers.google.com/maps/documentation/geocoding/get-api-key?authuser=2

	geocoder.ApiKey = "AIzaSyB-JwmMaEwb3yEomj66SnNlkA5GyKcRfWU"

	drop_table := `drop table if exists unemployment_data`
	_, err := db.Exec(drop_table)
	if err != nil {
		panic(err)
	}

	create_table := `CREATE TABLE IF NOT EXISTS unemployment_data (
						"id"   SERIAL , 
						"community_area" VARCHAR(255), 
						"unemployment_rate" FLOAT,
						PRIMARY KEY ("id") 
					);`

	_, _err := db.Exec(create_table)
	if _err != nil {
		panic(_err)
	}

	// While doing unit-testing keep the limit value to 500
	// later you could change it to 1000, 2000, 10,000, etc.
	fmt.Println("Grabbing data from Chicago Data...")
	var url = "https://data.cityofchicago.org/resource/iqnk-2tcu.json"

	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     300 * time.Second,
		DisableCompression:  true,
		TLSHandshakeTimeout: 1000 * time.Second,
	}

	client := &http.Client{Transport: tr}

	res, err := client.Get(url)

	if err != nil {
		panic(err)
	}

	fmt.Println("Starting JSON unmarshalling...")
	body, _ := ioutil.ReadAll(res.Body)
	var unemployment_list UnemploymentJsonRecords
	json.Unmarshal(body, &unemployment_list)
	fmt.Println("JSON unmarshalling done...")
	fmt.Println("Now unpacking JSON and inserting into db... ")
	for i := 0; i < len(unemployment_list); i++ {

		// We will execute definsive coding to check for messy/dirty/missing data values
		// Any record that has messy/dirty/missing data we don't enter it in the data lake/table

		community_area := unemployment_list[i].Community_area
		if community_area == "" {
			continue
		}

		unemployment_rate := unemployment_list[i].Unemployment
		if unemployment_rate == "" {
			continue
		}

		sql := `INSERT INTO unemployment_data ("community_area", "unemployment_rate"
			) values($1, $2)`

		_, err = db.Exec(
			sql,
			community_area,
			unemployment_rate)

		if err != nil {
			panic(err)
		}

	}
	fmt.Println("== Done with Unemployment Data ==")

}
