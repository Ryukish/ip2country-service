package database

import (
	"context"
	"encoding/binary"
	"fmt"
	"ip2country-service/internal/models"
	"log"
	"net"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
	collection *mongo.Collection
}

func NewMongoDatabase(uri, dbName string) (*MongoDatabase, error) {
	log.Printf("Connecting to MongoDB at URI: %s, DB: %s", uri, dbName)
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		return nil, err
	}

	// Ping MongoDB to ensure the connection is successful
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		return nil, err
	}
	log.Println("Successfully connected to MongoDB")

	collection := client.Database(dbName).Collection("ip_locations")
	return &MongoDatabase{collection: collection}, nil
}

func ipToUint32Mongo(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP address")
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4 address")
	}
	return binary.BigEndian.Uint32(ip), nil
}

func (db *MongoDatabase) Find(ipStr string) (*models.Location, error) {
	ipNum, err := ipToUint32Mongo(ipStr)
	if err != nil {
		log.Printf("Error converting IP to uint32: %v", err)
		return nil, err
	}

	filter := bson.M{
		"ip_from": bson.M{"$lte": ipNum},
		"ip_to":   bson.M{"$gte": ipNum},
	}

	var location IPLocation
	err = db.collection.FindOne(context.TODO(), filter).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("IP not found: %v", err)
			return nil, fmt.Errorf("IP not found")
		}
		log.Printf("Error finding IP: %v", err)
		return nil, err
	}

	return &models.Location{
		Country: location.Country,
		Region:  location.Region,
		City:    location.City,
	}, nil
}
