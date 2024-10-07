package database

import (
	"context"
	"encoding/binary"
	"fmt"
	"ip2country-service/internal/models"
	"ip2country-service/pkg/utils"
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
		return nil, fmt.Errorf("%w: %v", utils.ErrDatabaseQuery, err)
	}

	// Ping MongoDB to ensure the connection is successful
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		return nil, fmt.Errorf("%w: %v", utils.ErrDatabaseQuery, err)
	}
	log.Println("Successfully connected to MongoDB")

	collection := client.Database(dbName).Collection("ip_locations")
	return &MongoDatabase{collection: collection}, nil
}

func ipToUint32Mongo(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("%w: %s", utils.ErrInvalidIP, ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("%w: %s", utils.ErrInvalidIP, ipStr)
	}
	return binary.BigEndian.Uint32(ip), nil
}

func (db *MongoDatabase) Find(ipStr string) (*models.Location, error) {
	const funcName = "MongoDatabase.Find"
	ipNum, err := ipToUint32Mongo(ipStr)
	if err != nil {
		log.Printf("[%s] Error converting IP '%s' to uint32: %v", funcName, ipStr, err)
		return nil, err
	}

	filter := bson.M{
		"ip_from": bson.M{"$lte": ipNum},
		"ip_to":   bson.M{"$gte": ipNum},
	}

	// Log the filter for debugging
	log.Printf("[%s] MongoDB filter: %v", funcName, filter)

	var location IPLocation
	err = db.collection.FindOne(context.TODO(), filter).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[%s] IP '%s' not found in MongoDB", funcName, ipStr)
			return nil, utils.ErrDatabaseQuery
		}
		log.Printf("[%s] Error finding IP '%s': %v", funcName, ipStr, err)
		return nil, fmt.Errorf("%w: %v", utils.ErrDatabaseQuery, err)
	}

	log.Printf("[%s] IP '%s' found in MongoDB", funcName, ipStr)
	return &models.Location{
		Country: location.Country,
		Region:  location.Region,
		City:    location.City,
	}, nil
}
