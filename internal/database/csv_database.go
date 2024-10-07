package database

import (
	"encoding/csv"
	"fmt"
	"ip2country-service/internal/models"
	"ip2country-service/pkg/utils"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
)

type CSVDatabase struct {
	DatabaseLocal
}

func NewCSVDatabase(filePath string) (*CSVDatabase, error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening CSV file: %v", err)
		return nil, fmt.Errorf("%w: %v", utils.ErrDatabaseQuery, err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		log.Printf("Error reading CSV file: %v", err)
		return nil, fmt.Errorf("%w: %v", utils.ErrDatabaseQuery, err)
	}

	var locations []IPLocation
	for _, record := range records {
		if len(record) < 5 {
			log.Printf("Skipping incomplete record: %v", record)
			continue
		}

		ipFrom, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			log.Printf("Error parsing ip_from: %v", err)
			continue
		}

		ipTo, err := strconv.ParseUint(record[1], 10, 32)
		if err != nil {
			log.Printf("Error parsing ip_to: %v", err)
			continue
		}

		location := IPLocation{
			IPFrom:  uint32(ipFrom),
			IPTo:    uint32(ipTo),
			Country: record[2],
			Region:  record[3],
			City:    record[4],
		}

		locations = append(locations, location)
	}

	// Sort the locations by IPFrom for efficient searching
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].IPFrom < locations[j].IPFrom
	})

	log.Printf("Loaded %d locations from CSV", len(locations))
	return &CSVDatabase{DatabaseLocal{Locations: locations}}, nil
}

func ipToUint32CSV(ipStr string) (uint32, error) {
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

func (db *CSVDatabase) Find(ipStr string) (*models.Location, error) {
	const funcName = "CSVDatabase.Find"
	ipNum, err := ipToUint32CSV(ipStr)
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
