package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IPLocation struct {
	IPFrom  uint32 `json:"ip_from" bson:"ip_from"`
	IPTo    uint32 `json:"ip_to" bson:"ip_to"`
	Country string `json:"country" bson:"country"`
	Region  string `json:"region" bson:"region"`
	City    string `json:"city" bson:"city"`
}

func main() {
	// Command-line flags
	var (
		filePath string
		fileType string
	)
	flag.StringVar(&filePath, "file", "data/ip_database.json", "Path to the data file (CSV or JSON)")
	flag.StringVar(&fileType, "type", "", "Type of the data file: csv or json (optional, inferred from extension if not provided)")
	flag.Parse()

	// Determine file type based on extension if not provided
	if fileType == "" {
		if strings.HasSuffix(strings.ToLower(filePath), ".csv") {
			fileType = "csv"
		} else if strings.HasSuffix(strings.ToLower(filePath), ".json") {
			fileType = "json"
		} else {
			log.Fatal("Could not determine file type. Please specify using the -type flag.")
		}
	}

	// MongoDB connection settings
	mongoURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGODB_NAME")

	if mongoURI == "" || dbName == "" {
		log.Fatal("MONGODB_URI and MONGODB_NAME environment variables are required")
	}

	clientOpts := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database(dbName).Collection("ip_locations")

	// Check if data already exists in the collection
	count, err := collection.CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		log.Fatalf("Failed to count documents: %v", err)
	}

	if count > 0 {
		log.Println("Data already exists in the collection. Skipping migration.")
		return
	}

	log.Println("No existing data found. Proceeding with migration.")

	var documents []interface{}

	switch fileType {
	case "csv":
		documents, err = parseCSV(filePath)
	case "json":
		documents, err = parseJSON(filePath)
	default:
		log.Fatalf("Unsupported file type: %s", fileType)
	}

	if err != nil {
		log.Fatalf("Failed to parse data file: %v", err)
	}

	// Insert documents into MongoDB in batches
	batchSize := 1000
	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		_, err := collection.InsertMany(context.TODO(), documents[i:end])
		if err != nil {
			log.Fatalf("Failed to insert documents: %v", err)
		}
		fmt.Printf("Inserted %d records\n", end-i)
	}

	fmt.Println("Data migration completed successfully!")
}

func parseCSV(filePath string) ([]interface{}, error) {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	var documents []interface{}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %v", err)
		}

		if len(record) < 5 {
			continue // Skip incomplete records
		}

		ipFrom, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			log.Printf("Invalid IPFrom value: %v", record[0])
			continue
		}

		ipTo, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			log.Printf("Invalid IPTo value: %v", record[1])
			continue
		}

		location := IPLocation{
			IPFrom:  uint32(ipFrom),
			IPTo:    uint32(ipTo),
			Country: record[2],
			Region:  record[3],
			City:    record[4],
		}

		documents = append(documents, location)
	}

	return documents, nil
}

func parseJSON(filePath string) ([]interface{}, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open JSON file: %v", err)
	}
	defer jsonFile.Close()

	var locations []IPLocation
	decoder := json.NewDecoder(jsonFile)
	err = decoder.Decode(&locations)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	documents := make([]interface{}, len(locations))
	for i, location := range locations {
		documents[i] = location
	}

	return documents, nil
}
