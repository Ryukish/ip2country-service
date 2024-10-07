package database

import (
	"encoding/json"
	"fmt"
	"ip2country-service/internal/models"
	"log"
	"net"
	"os"
	"sort"
)

type JSONDatabase struct {
	DatabaseLocal
}

func NewJSONDatabase(filePath string) (*JSONDatabase, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening JSON file: %v", err)
		return nil, fmt.Errorf("failed to open JSON file")
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var locations []IPLocation

	err = decoder.Decode(&locations)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return nil, fmt.Errorf("failed to decode JSON")
	}

	// Sort the locations by IPFrom for efficient searching
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].IPFrom < locations[j].IPFrom
	})

	return &JSONDatabase{DatabaseLocal{Locations: locations}}, nil
}

func ipToUint32Json(ipStr string) (uint32, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP address")
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("invalid IPv4 address")
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3]), nil
}

func (db *JSONDatabase) Find(ipStr string) (*models.Location, error) {
	ipNum, err := ipToUint32Json(ipStr)
	if err != nil {
		log.Printf("Invalid IP address: %v", err)
		return nil, err
	}

	// Binary search to find the IP range
	index := sort.Search(len(db.Locations), func(i int) bool {
		return db.Locations[i].IPTo >= ipNum
	})

	if index < len(db.Locations) && db.Locations[index].IPFrom <= ipNum && ipNum <= db.Locations[index].IPTo {
		loc := db.Locations[index]
		log.Printf("IP found in range: %d - %d", loc.IPFrom, loc.IPTo)
		return &models.Location{
			Country: loc.Country,
			Region:  loc.Region,
			City:    loc.City,
		}, nil
	}

	log.Printf("IP not found in any range")
	return nil, fmt.Errorf("IP not found")
}
