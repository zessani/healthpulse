/*
  Program: report.go
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
	"crypto/rand"
	"fmt"
	"net"
	"sync"
	"time"
)

// Its a string type made for LOW, MEDIUM, HIGH
type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"    // score under 0.35
	RiskMedium RiskLevel = "MEDIUM" // score between 0.35 and 0.65
	RiskHigh   RiskLevel = "HIGH"   // score over 0.65
)

// holds information about weather for a zip code
type WeatherInfo struct {
	Temp        float64 // temperature in celcius
	Humidity    int     // humidity
	Description string  // a short description
}

// this holds a disease activity within a zip code
type CDCActivity struct {
	ZipCode       string    // zip code
	ActivityLevel string    // level like low, medium, high
	DiseaseTag    string    // disease name
	UpdatedAt     time.Time // time about when got this data
}

// warning about disease in a region
type PathogenAlert struct {
	AlertID       string    // id for alert
	Region        string    // which part of the city has alert
	Pathogen      string    // what the disease is
	SeverityScore float64   // how bad the disease is
	IssuedAt      time.Time // when it was created
}

// report using one health system from user
type Report struct {
	ID            string
	ZipCode       string
	Symptoms      []string
	TravelHistory bool // if they traveled out of state
	AnimalContact bool // did they touch animal recently
	Timestamp     time.Time
	RiskScore     float64
	RiskLevel     RiskLevel
	WeatherData   WeatherInfo
}

// cluster is made if too many people in same zip code have same symptoms
type ClusterAlert struct {
	ZipCode    string
	Count      int
	Symptoms   []string
	RiskLevel  RiskLevel
	DetectedAt time.Time
	Message    string
}

// ReportStore keeps all the reports and connected clients in memory
type ReportStore struct {
	mu      sync.Mutex
	reports []Report
	clients []net.Conn
}

/*
NewReportStore()
Purpose - Makes new empty ReportStore ready to use.
pre condition - none
Post condition - returns a new empty store
Parameters - none
*/
func NewReportStore() *ReportStore {
	return &ReportStore{
		reports: make([]Report, 0),
		clients: make([]net.Conn, 0),
	}
}

/*
AddReport(r)
Purpose - Adds a report to the store and add a look to keep it safe when multiple
clients submit report at same time.
pre condition - store exist
Post condition - report is added
Parameters -
r - the report to add
*/
func (s *ReportStore) AddReport(r Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports = append(s.reports, r)
}

/*
GetReportByZip(zip, window)
Purpose - It get all the reports from a specific zip code witin a time window
pre condition - store
Post condition - returns matching report
Parameters -
zip - the zip code to search for
window - the time frame
*/
func (s *ReportStore) GetReportsByZip(zip string, window time.Duration) []Report {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-window)
	var results []Report
	for _, r := range s.reports {
		if r.ZipCode == zip && r.Timestamp.After(cutoff) {
			results = append(results, r)
		}
	}
	return results
}

/*
CheckCluster(zip)
Purpose - It looks at the recent reports in the zip and check if there is a cluster. cluster is
checked if there is 3+ reports with same symptoms. if yes then generate cluster alert.
pre condition - store haas report in it
Post condition - if there is cluster than generate alert.
Parameters -
zip - the zip code
*/
func (s *ReportStore) CheckCluster(zip string) (ClusterAlert, bool) {
	recent := s.GetReportsByZip(zip, 24*time.Hour)

	// need at least three report to generate cluster
	if len(recent) < 3 {
		return ClusterAlert{}, false
	}

	// count of symptoms
	symptomCount := make(map[string]int)
	for _, r := range recent {
		for _, sym := range r.Symptoms {
			symptomCount[sym]++
		}
	}

	// find symptom that show uo more than 2 times
	var dominant []string
	for sym, count := range symptomCount {
		if count >= 2 {
			dominant = append(dominant, sym)
		}
	}

	// if no symptom is repeated, then no cluster
	if len(dominant) == 0 {
		return ClusterAlert{}, false
	}

	alert := ClusterAlert{
		ZipCode:    zip,
		Count:      len(recent),
		Symptoms:   dominant,
		RiskLevel:  RiskHigh,
		DetectedAt: time.Now(),
		Message:    fmt.Sprintf("CLUSTER ALERT: %d overlapping cases in ZIP %s (%s) — Risk: HIGH", len(recent), zip, joinSymptoms(dominant)),
	}
	return alert, true
}

/*
joinSymptoms( symptoms)
Purpose - It takes a list of symptoms and join them in one string seperated by comma.
pre condition - none
Post condition - none
Parameters -
symptoms - list of symptoms
*/
func joinSymptoms(symptoms []string) string {
	result := ""
	for i, s := range symptoms {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

/*
symptomsOverLap(a, b)
Purpose - It returns how many common symptoms are there in two list.
pre condition - none
Post condition - returns a number of shared symptoms
Parameters -
a - first list
b - second list
*/
func symptomOverlap(a, b []string) int {
	set := make(map[string]bool) // put all the list into set
	for _, s := range a {
		set[s] = true
	}
	count := 0 // how many symptoms from b in a
	for _, s := range b {
		if set[s] {
			count++
		}
	}
	return count
}

/*
addClient(s, conn)
Purpose - It add new connection to clients list so we can send them alerts.
pre condition - Store alerts
Post condition - Connection is added to the list
Parameters -
s - the report store
conn - the tcp connection to add
*/
func addClient(s *ReportStore, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients = append(s.clients, conn)
}

/*
removeClient(s, conn)
Purpose - It remove a connection from client list when someone disconnects.
pre condition - connection was added with addClient
Post condition - connection is removed from the list
Parameters -
s - the report store
conn - the tcp connection to add
*/
func removeClient(s *ReportStore, conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, c := range s.clients {
		if c == conn {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			return
		}
	}
}

/*
generateID()
Purpose - Make a random string to use it as id.
pre condition - none
Post condition - none
Parameters - none
*/
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
