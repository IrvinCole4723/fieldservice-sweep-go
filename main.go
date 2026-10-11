package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	client, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	workerURL := os.Getenv("SWEEP_WORKER_URL")
	if workerURL == "" {
		workerURL = "https://example.com/fieldservice/sweep"
	}
	var job struct {
		JobID string `json:"job_id"`
	}
	if err := client.CronCreate("0 * * * *", workerURL, &job); err != nil {
		log.Fatal(err)
	}
	now := time.Now().UTC()
	orders := []WorkOrder{{ID: "WO-101", PhotoArchived: true, DispatchStatus: "completed", FollowUpDue: now.Add(-24 * time.Hour)}, {ID: "WO-102", PhotoArchived: false, DispatchStatus: "completed", FollowUpDue: now.Add(-48 * time.Hour)}}
	for _, order := range staleOrders(orders, now) {
		payload, _ := json.Marshal(order)
		var message struct {
			MessageID string `json:"message_id"`
		}
		if err := client.QueuePublish("fieldservice-cleanup", string(payload), &message); err != nil {
			log.Fatal(err)
		}
		fmt.Println("queued cleanup", order.ID, message.MessageID)
	}
	fmt.Println("scheduled sweep", job.JobID)
}
