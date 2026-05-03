/*
  Authors:      Zayyan Essani, Syed Zoraiz Asif
  Course:       CSc 372 - Comparative Programming Languages
  Assignment:   Project Part 3 - Creative Program
  Instructor:   Lester McCann
  TAs:          Daniel Reynalod, Muaz Ali
  Due Date:     May 4, 2026
 
  Description:  TCP server that handles client connections for the
                HealthPulse system. Spawns one goroutine per client,
                runs an interactive symptom interview, computes risk
                using concurrent API calls, stores the report, checks
                for clusters, and broadcasts alerts to all connected
                clients when a cluster is detected.
 
  Language:     Go 1.26
  Known Bugs:   None
 */

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)
// symptoms users can select from
var symptomList = []string{
	"fever",
	"cough",
	"fatigue",
	"difficulty_breathing",
	"headache",
	"nausea",
}
/*
StartTCPServer - Listens on the given port and spawns a goroutine
for each new client connection.
Pre - store and wc are initialized.
Post - Runs forever accepting connections.
Params -
port  - port string like ":8080"
store - shared report store
wc - shared weather cache
 */
func StartTCPServer(port string, store *ReportStore, wc *WeatherCache) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Failed to start TCP server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("TCP server listening on %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}
		go handleClient(conn, store, wc)
	}
}
/*
handleClient - Handles a single client. Collects symptoms, zip, travel,
and animal info, computes risk, stores the report, and
checks for clusters. Loops so the user can submit again.
Pre - conn is open, store and wc exist.
Post - Report stored, alert broadcast if cluster found.
Params -
conn  - TCP connection to the client
store - shared report store
wc  - shared weather cache
 */
func handleClient(conn net.Conn, store *ReportStore, wc *WeatherCache) {
	defer conn.Close()
	addClient(store, conn)
	defer removeClient(store, conn)

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	writeLine(rw, "")
	writeLine(rw, "Welcome to HealthPulse Surveillance System")
	writeLine(rw, "==========================================")
	writeLine(rw, "")

	for {
		symptoms, err := promptSymptoms(rw)
		if err != nil {
			return
		}

		zip, err := promptZip(rw)
		if err != nil {
			return
		}

		travel, err := promptYesNo(rw, "Recent travel outside Arizona? (y/n): ")
		if err != nil {
			return
		}

		animal, err := promptYesNo(rw, "Recent animal contact? (y/n): ")
		if err != nil {
			return
		}

		writeLine(rw, "")
		writeLine(rw, "Computing risk assessment...")
		writeLine(rw, "")

		level, score, weather := ComputeRisk(symptoms, zip, travel, animal, wc, store)

		id := generateID()
		report := Report{
			ID:            id,
			ZipCode:       zip,
			Symptoms:      symptoms,
			TravelHistory: travel,
			AnimalContact: animal,
			Timestamp:     time.Now(),
			RiskScore:     score,
			RiskLevel:     level,
			WeatherData:   weather,
		}
		store.AddReport(report)
		// show risk assessment
		writeLine(rw, "==================================")
		writeLine(rw, fmt.Sprintf("  RISK LEVEL:  %s (score: %.2f)", level, score))
		writeLine(rw, fmt.Sprintf("  Report ID:   %s", id))
		writeLine(rw, fmt.Sprintf("  Weather:     %.1f°C, %d%% humidity, %s", weather.Temp, weather.Humidity, weather.Description))
		writeLine(rw, "==================================")
		writeLine(rw, "Thank you. Your report has been recorded.")
		writeLine(rw, "")
		// broadcast if cluster detected
		if alert, found := store.CheckCluster(zip); found {
			broadcastAlert(store, alert)
		}

		writeLine(rw, "Press Enter to submit another report, or type 'quit' to disconnect.")
		write(rw, "> ")
		input, err := rw.ReadString('\n')
		if err != nil {
			return
		}
		if strings.TrimSpace(strings.ToLower(input)) == "quit" {
			writeLine(rw, "Goodbye!")
			return
		}
		writeLine(rw, "")
	}
}
/*
promptSymptoms - Shows symptom menu, reads comma-separated selection.
Re-prompts until valid input is given.
Pre - conn is open.
Post - None
Params - rw - buffered reader/writer
Returns - selected symptoms, or error if connection drops
 */
func promptSymptoms(rw *bufio.ReadWriter) ([]string, error) {
	for {
		writeLine(rw, "[1] Select symptoms (comma-separated numbers):")
		for i, s := range symptomList {
			writeLine(rw, fmt.Sprintf("  %d. %s", i+1, formatSymptom(s)))
		}
		write(rw, "> ")

		input, err := rw.ReadString('\n')
		if err != nil {
			return nil, err
		}

		input = strings.TrimSpace(input)
		parts := strings.Split(input, ",")

		var selected []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			idx := 0
			fmt.Sscanf(p, "%d", &idx)
			if idx >= 1 && idx <= len(symptomList) {
				selected = append(selected, symptomList[idx-1])
			}
		}

		if len(selected) > 0 {
			return selected, nil
		}
		writeLine(rw, "Invalid input. Please enter at least one number (1-6).")
		writeLine(rw, "")
	}
}
/*
promptZip - Asks for a 5-digit zip code. Re-prompts until valid.
Pre - conn is open.
Post - None
Params - rw - buffered reader/writer
Returns - valid 5-digit zip string, or error if connection drops
 */
func promptZip(rw *bufio.ReadWriter) (string, error) {
	for {
		writeLine(rw, "")
		writeLine(rw, "[2] Enter your 5-digit ZIP code (AZ only):")
		write(rw, "> ")

		input, err := rw.ReadString('\n')
		if err != nil {
			return "", err
		}

		zip := strings.TrimSpace(input)
		if len(zip) == 5 {
			allDigits := true
			for _, c := range zip {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return zip, nil
			}
		}
		writeLine(rw, "Invalid ZIP code. Please enter a 5-digit number (e.g. 85719).")
	}
}
/*
promptYesNo - Asks a yes/no question. Re-prompts until valid. 
Pre - conn is open.
Post - None
Params - 
rw  - buffered reader/writer
question - question to display
Returns: true for yes, false for no, or error if connection drops
 */
func promptYesNo(rw *bufio.ReadWriter, question string) (bool, error) {
	for {
		writeLine(rw, "")
		write(rw, question)

		input, err := rw.ReadString('\n')
		if err != nil {
			return false, err
		}

		answer := strings.ToLower(strings.TrimSpace(input))
		if answer == "y" || answer == "yes" {
			return true, nil
		}
		if answer == "n" || answer == "no" {
			return false, nil
		}
		writeLine(rw, "Invalid input. Please enter y or n.")
	}
}
/*
broadcastAlert - Sends a cluster alert to every connected client.
Pre - store has connected clients.
Post - Alert written to all connections.
Params - 
store (in) - report store with client list
alert (in) - cluster alert to send
 */
func broadcastAlert(store *ReportStore, alert ClusterAlert) {
	store.mu.Lock()
	defer store.mu.Unlock()

	msg := fmt.Sprintf("\n*** %s ***\n", alert.Message)

	for _, conn := range store.clients {
		conn.Write([]byte(msg))
	}
}
/*
formatSymptom - Replaces underscores with spaces and capitalizes
the first letter for display.
Pre - None
Post - None
Params -
s 
Returns - formatted string 
 */
func formatSymptom(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	return strings.ToUpper(s[:1]) + s[1:]
}
/*
writeLine - Writes a string with CRLF to the client.
Pre - conn is open.
Post - String written and flushed.
Params -
rw - buffered reader/writer
s - string to write
 */
func writeLine(rw *bufio.ReadWriter, s string) {
	rw.WriteString(s + "\r\n")
	rw.Flush()
}
/*
write - Writes a string without a newline to the client. 
Pre - conn is open.
Post -  String written and flushed.
Params -
rw 
s   
 */
func write(rw *bufio.ReadWriter, s string) {
	rw.WriteString(s)
	rw.Flush()
}
