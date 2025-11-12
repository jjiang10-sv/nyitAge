package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	// Define the target
	// targetUrlEnv := "VICTIM_URL"
	// targeturl, _ := os.LookupEnv(targetUrlEnv)

	// // Prepare the attacker with the rate and duration
	// freqEnvVar := "FREQ" // replace with your environment variable name
	// freqStr, _ := os.LookupEnv(freqEnvVar)
	// freq, _ := strconv.Atoi(freqStr)

	// durEnvVar := "DURATION" // replace with your environment variable name
	// durEnvVarStr, _ := os.LookupEnv(durEnvVar)
	// durationInSeconds, _ := strconv.Atoi(durEnvVarStr)
	// duration := time.Duration(durationInSeconds) * time.Second

	freq := 20
	// 5*10*60
	duration := time.Duration(60*10) * time.Second

	rate := vegeta.Rate{Freq: freq, Per: time.Second}

	// docs   health
	//targeturl := "https://dev.intersoul.io/docs"
	targeturl := "http://localhost:8081"

	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    targeturl,
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration, "Big Bang!") {
		// for k, v := range res.Headers {
		// 	fmt.Printf("%s: %s\n", k, v)
		// }
		//fmt.Println(string(res.Body), res.Seq, res.Method, res.Error)

		fmt.Printf("body: %s seq num: %d method: %s code: %d error: %s\n", string(res.Body), res.Seq, res.Method, res.Code, res.Error) // Updated line

		//res.Headers
		metrics.Add(res)
	}
	metrics.Close()
	printMetrics(&metrics)
	//exportMetricsToJSON(&metrics)
}

// printMetrics prints the basic metrics collected during the attack
func printMetrics(metrics *vegeta.Metrics) {
	fmt.Printf("Requests per second: %.2f\n", metrics.Rate)
	fmt.Printf("Total requests latency: %d\n", metrics.Latencies.Total)

	fmt.Printf("Success rate: %.2f%%\n", metrics.Success*100)
	fmt.Printf("Average latency: %.2f ms\n", metrics.Latencies.Mean.Seconds()*1000)

	// Print detailed latencies
	fmt.Println("Detailed latencies (ms):")
	fmt.Printf("P50: %.2f, P95: %.2f, P99: %.2f\n",
		metrics.Latencies.P50.Seconds()*1000,
		metrics.Latencies.P95.Seconds()*1000,
		metrics.Latencies.P99.Seconds()*1000)
}

// exportMetricsToJSON exports metrics to a JSON file for analysis
func exportMetricsToJSON(metrics *vegeta.Metrics) {
	jsonMetrics, err := json.Marshal(metrics)
	if err != nil {
		log.Fatalf("Error serializing metrics to JSON: %v", err)
	}

	// Write to a file
	err = os.WriteFile("metrics.json", jsonMetrics, 0644)
	if err != nil {
		log.Fatalf("Error writing metrics to file: %v", err)
	}

	fmt.Println("Metrics have been exported to metrics.json.")
}
