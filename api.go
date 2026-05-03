/*
Program: api.go
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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// WeatherCache stores weather so we dont have to fetch every time
type WeatherCache struct {
	mu   sync.Mutex
	data map[string]weatherCacheEntry
}

type weatherCacheEntry struct {
	info      WeatherInfo
	fetchedAt time.Time
}

/*
NewWeatherCache()
Purpose - Makes a new empty weather cache.
Pre condition - none
Post condition - returns a new empty cache ready to use
Parameters - none
*/
func NewWeatherCache() *WeatherCache {
	return &WeatherCache{
		data: make(map[string]weatherCacheEntry),
	}
}

// owmResponse is the JSON format that OpenWeatherMap will use if we have API in future
type owmResponse struct {
	Main struct {
		Temp      float64 `json:"temp"`     // temp in celsius
		Humidity  int     `json:"humidity"` // humidity %
		FeelsLike float64 `json:"feels_like"`
	} `json:"main"`
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
}

/*
fetchWeather(zip, wc)
Purpose - It gets weather data from zip code and checks the cache first and returns the weather data.
Pre condition - cache has been created with NewWeatherCache
Post condition - cache might get updated with new data
Parameters -
zip  - zip code
wc - the weather cache
*/
func fetchWeather(zip string, wc *WeatherCache) (WeatherInfo, error) {
	wc.mu.Lock()
	if entry, ok := wc.data[zip]; ok {
		if time.Since(entry.fetchedAt) < 10*time.Minute {
			wc.mu.Unlock()
			return entry.info, nil
		}
	}
	wc.mu.Unlock()

	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return mockWeather(), nil
	}

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?zip=%s,us&appid=%s&units=metric", zip, apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return mockWeather(), err
	}
	defer resp.Body.Close()

	var owm owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&owm); err != nil {
		return mockWeather(), err
	}

	desc := "clear sky"
	if len(owm.Weather) > 0 {
		desc = owm.Weather[0].Description
	}

	info := WeatherInfo{
		Temp:        owm.Main.Temp,
		Humidity:    owm.Main.Humidity,
		Description: desc,
	}

	wc.mu.Lock()
	wc.data[zip] = weatherCacheEntry{info: info, fetchedAt: time.Now()}
	wc.mu.Unlock()

	return info, nil
}

/*
mockWeather()
Purpose - returns a hardcoded weather. as we are not using any api.
Pre condition - none
Post condition - none.
Parameters - none
*/
func mockWeather() WeatherInfo {
	return WeatherInfo{
		Temp:        35.0,
		Humidity:    40,
		Description: "clear sky",
	}
}

/*
fetchCDCAActivity(zip)
Purpose - It returns a disease activity for a ZIP code. differnt zip get different results according to thier data.
Pre condition - none
Post condition - none
Parameters -
zip - the zip code
*/
func fetchCDCActivity(zip string) CDCActivity {
	// different levels and disease to pick from
	levels := []string{"minimal", "low", "low", "moderate", "high"}
	diseases := []string{"influenza-a", "rsv", "covid19", "influenza-b", "norovirus"}

	zipInt, err := strconv.Atoi(zip) // turns zip in to number
	if err != nil {
		zipInt = 0
	}
	idx := zipInt % 5 // pick an index based on the zip

	return CDCActivity{
		ZipCode:       zip,
		ActivityLevel: levels[idx],
		DiseaseTag:    diseases[idx],
		UpdatedAt:     time.Now(),
	}
}

/*
fetchPathogenAlerts()
Purpose - Returns a list of pathagons alerts. right now only valley fever alert.
Pre condition - none
Post condition - none
Parameters - none
*/
func fetchPathogenAlerts() []PathogenAlert {
	return []PathogenAlert{
		{
			AlertID:       "PA-001",
			Region:        "Southern Arizona",
			Pathogen:      "Valley Fever (Coccidioides)",
			SeverityScore: 0.3,
			IssuedAt:      time.Now(),
		},
	}
}
