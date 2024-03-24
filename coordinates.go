package main

import (
	"encoding/csv"
	"log"
	"strconv"
	"strings"
)

type coordinates struct {
	MeshName     string
	MostPositive [3]float64
	MostNegative [3]float64
}

func NewCoordinates(meshName string, mostPositive [3]float64, mostNegative [3]float64) *coordinates {
	return &coordinates{
		MeshName:     meshName,
		MostPositive: mostPositive,
		MostNegative: mostNegative,
	}
}

func GetCoordinatesFromCsv(csvFile []byte) []*coordinates {
	// Parse the CSV
	records, err := csv.NewReader(strings.NewReader(string(csvFile))).ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	// Create the coordinates
	coordinates := make([]*coordinates, 0)
	for i := 1; i < len(records); i++ {
		record := records[i]
		mostPositive := [3]float64{convertToFloat64(record[1]), convertToFloat64(record[2]), convertToFloat64(record[3])}
		mostNegative := [3]float64{convertToFloat64(record[4]), convertToFloat64(record[5]), convertToFloat64(record[6])}
		coordinates = append(coordinates, NewCoordinates(record[0], mostPositive, mostNegative))
	}

	return coordinates
}

// Function to convert string to float64
func convertToFloat64(value string) float64 {
	value = strings.TrimSpace(value) // Trim leading and trailing spaces
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		log.Fatal(err)
	}
	return floatValue
}
