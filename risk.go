/*
Program: risk.go
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

// weatherResult is used to send weather data through a channel
type weatherResult struct {
	info WeatherInfo // weather info
	err  error       // any error that happened
}

// this is used to send CDC data through channel
type cdcResult struct {
	data CDCActivity
}

// pathogenResults is used to send pathogen data throuh channel
type pathogenResult struct {
	alerts []PathogenAlert
}

/*
ComputerRisk(symptoms, zip, travel, animal, wc, store)
Purpose - This function is the main risk calculator. It sends three goroutines to grab weather, CDC and pathogen data
at the same timeusing channels. Then it adds up score from symptoms, weather, CDC, pathogen, animal and wheather
there are more reports from the same zip code.
Pre condition - cache and store are setup
Post condition - returns risk level, the score and weather info.
Parameters -
symptoms - the list of symptoms we picked
zip - the zip code
travel - if there travel any where out of state
animal - if they had any animal contact
wc - weather cache
store - to check for cluster
*/
func ComputeRisk(symptoms []string, zip string, travel, animal bool, wc *WeatherCache, store *ReportStore) (RiskLevel, float64, WeatherInfo) {
	// make channels to get data back from goroutines
	wCh := make(chan weatherResult, 1)  // weather chaannel
	cCh := make(chan cdcResult, 1)      // CDC channel
	pCh := make(chan pathogenResult, 1) // pathogen channel

	// send 3 goroutines to fetch data at the same time
	go func() {
		w, e := fetchWeather(zip, wc)
		wCh <- weatherResult{w, e}
	}()
	go func() {
		cCh <- cdcResult{fetchCDCActivity(zip)}
	}()
	go func() {
		pCh <- pathogenResult{fetchPathogenAlerts()}
	}()

	// wait for all three to finish
	wRes := <-wCh
	cRes := <-cCh
	pRes := <-pCh

	// add up the total risk from everything
	score := 0.0
	score += scoreSymptoms(symptoms)
	score += scoreWeather(wRes.info)
	score += scoreCDC(cRes.data)
	score += scorePathogens(pRes.alerts)

	if travel {
		score += 0.05
	}
	if animal {
		score += 0.05
	}

	// if 2+ other reports from same zip code then add more risk
	clusterReports := store.GetReportsByZip(zip, 24*60*60*1e9)
	if len(clusterReports) >= 2 {
		score += 0.10
	}

	// score not to go over 1.0
	if score > 1.0 {
		score = 1.0
	}

	return levelFromScore(score), score, wRes.info
}

/*
scoreSymptoms(symptom)
Purpose - Gives weight to each symptoms based upon how serious they are.
Pre condition - none
Post condition - none
Parameters -
symptom - a list of symptoms
*/
func scoreSymptoms(symptoms []string) float64 {
	weights := map[string]float64{
		"fever":                0.30,
		"difficulty_breathing": 0.35,
		"cough":                0.15,
		"fatigue":              0.10,
		"headache":             0.05,
		"nausea":               0.05,
	}

	score := 0.0
	for _, s := range symptoms {
		if w, ok := weights[s]; ok {
			score += w
		}
	}
	if score > 1.0 {
		score = 1.0
	}
	return score
}

/*
scoreWeather(w)
Purpose - Add risk if the weather is extreme.
Pre condition - none
Post condition - none
Parameters -
w - data about weather
*/
func scoreWeather(w WeatherInfo) float64 {
	score := 0.0
	if w.Temp > 38 {
		score += 0.10
	}
	if w.Humidity > 85 {
		score += 0.10
	}
	return score
}

/*
scoreCDC(cdc)
Purpose - This function turns the cdc level in to number.
Pre condition - none
Post condition - none
Parameters -
cdc - the data about cdc
*/
func scoreCDC(cdc CDCActivity) float64 {
	levels := map[string]float64{
		"minimal":  0.0,
		"low":      0.05,
		"moderate": 0.10,
		"high":     0.15,
	}
	if s, ok := levels[cdc.ActivityLevel]; ok {
		return s
	}
	return 0.0
}

/*
scorePathogens(alerts)
Purpose - It add the severity score from all the pathogen alert
Pre condition - none
Post condition - none
Parameters -
alert - list of pathogen alert
*/
func scorePathogens(alerts []PathogenAlert) float64 {
	score := 0.0
	for _, a := range alerts {
		score += a.SeverityScore
	}
	if score > 0.15 {
		score = 0.15
	}
	return score
}

/*
levelFromScore(score)
Purpose - Turns the score in to LOW, MEDIUM or HIGH based upon score.
Pre condition - score is between 0to 1.0
Post condition - none
Parameters - none
*/
func levelFromScore(score float64) RiskLevel {
	if score < 0.35 {
		return RiskLow
	}
	if score <= 0.65 {
		return RiskMedium
	}
	return RiskHigh
}
