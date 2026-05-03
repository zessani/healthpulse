/*
  Authors:      Zayyan Essani, Syed Zoraiz Asif
  Course:       CSc 372 - Comparative Programming Languages
  Assignment:   Project Part 3 - Creative Program
  Instructor:   Lester McCann
  TAs:          Daniel Reynalod, Muaz Ali
  Due Date:     May 4, 2026
 
  Description:  Entry point for the HealthPulse surveillance system.
                Creates the shared report store and weather cache,
                then starts the TCP server on port 8080.
 
  Language:     Go 1.26
  Compile/Run:  go run .
                Then connect with: nc localhost 8080
  Known Bugs:   None
 */

package main

import "fmt"

/*
  main - Entry point for the HealthPulse server.
 
  Pre:     None
  Post:    TCP server is running on port 8080.
 */
func main() {
	store := NewReportStore() 
	wc := NewWeatherCache()  

	fmt.Println("==========================================")
	fmt.Println("  HealthPulse Surveillance System")
	fmt.Println("  TCP server starting on :8080")
	fmt.Println("==========================================")

	StartTCPServer(":8080", store, wc) 
}