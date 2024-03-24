package main

import (
	"encoding/csv"
	"log"
	"strconv"
	"strings"
)

type coordinates struct {
	MeshName     string
	MostPositive [3]float32
	MostNegative [3]float32
}

func NewCoordinates(meshName string, mostPositive [3]float32, mostNegative [3]float32) *coordinates {
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
		mostPositive := [3]float32{convertToFloat32(record[1]), convertToFloat32(record[2]), convertToFloat32(record[3])}
		mostNegative := [3]float32{convertToFloat32(record[4]), convertToFloat32(record[5]), convertToFloat32(record[6])}
		coordinates = append(coordinates, NewCoordinates(record[0], mostPositive, mostNegative))
	}

	return coordinates
}

// Function to convert string to float64
func convertToFloat32(value string) float32 {
	value = strings.TrimSpace(value) // Trim leading and trailing spaces
	floatValue, err := strconv.ParseFloat(value, 32)
	if err != nil {
		log.Fatal(err)
	}
	return float32(floatValue)
}
