package database

import (
	"encoding/json"
	"fmt"
	"ip2country-service/internal/models"
	"ip2country-service/pkg/utils"
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
		return nil, fmt.Errorf("%w: %v", utils.ErrDatabaseQuery, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var locations []IPLocation

	err = decoder.Decode(&locations)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return nil, fmt.Errorf("%w: %v", utils.ErrJSONUnmarshal, err)
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
		return 0, fmt.Errorf("%w: %s", utils.ErrInvalidIP, ipStr)
	}
	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("%w: %s", utils.ErrInvalidIP, ipStr)
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3]), nil
}

func (db *JSONDatabase) Find(ipStr string) (*models.Location, error) {
	const funcName = "JSONDatabase.Find"
	ipNum, err := ipToUint32Json(ipStr)
	if err != nil {
		log.Printf("[%s] Error converting IP '%s' to uint32: %v", funcName, ipStr, err)
		return nil, fmt.Errorf("%w: %s", utils.ErrInvalidIP, ipStr)
	}

	// Binary search to find the IP range
	index := sort.Search(len(db.Locations), func(i int) bool {
		return db.Locations[i].IPTo >= ipNum
	})

	if index < len(db.Locations) && db.Locations[index].IPFrom <= ipNum {
		loc := db.Locations[index]
		log.Printf("[%s] IP '%s' found in range %d - %d", funcName, ipStr, loc.IPFrom, loc.IPTo)
		return &models.Location{
			Country: loc.Country,
			Region:  loc.Region,
			City:    loc.City,
		}, nil
	}

	log.Printf("[%s] IP '%s' not found in any range", funcName, ipStr)
	return nil, fmt.Errorf("%w: %s", utils.ErrDatabaseQuery, ipStr)
}
