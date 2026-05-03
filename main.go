/*
  Program: main.go
  Author: Syed Zoraiz Asif
  Course: CSc 372
  Assignment: part#3
  Instructor: Lester McCann
  TAs: Daniel Reyanalod, Muaz Ali
  Due Date: May 4, 2026

  Description:
  Language: Go
  Known bugs: None
*/

package main

import (
	"fmt"
	"os"
)

func main() {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		fmt.Println("[WARNING] OPENWEATHER_API_KEY not set — using mock weather data")
	}

	store := NewReportStore()
	wc := NewWeatherCache()

	fmt.Println("==========================================")
	fmt.Println("  HealthPulse Surveillance System")
	fmt.Println("  TCP server starting on :8080")
	fmt.Println("==========================================")

	StartTCPServer(":8080", store, wc)
}
