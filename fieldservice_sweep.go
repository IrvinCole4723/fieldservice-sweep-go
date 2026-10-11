package main

import "time"

type WorkOrder struct {
	ID             string
	PhotoArchived  bool
	DispatchStatus string
	FollowUpDue    time.Time
}

func ShouldClean(w WorkOrder, now time.Time) bool {
	return w.PhotoArchived && w.DispatchStatus == "completed" && !w.FollowUpDue.After(now)
}

func staleOrders(orders []WorkOrder, now time.Time) []WorkOrder {
	var result []WorkOrder
	for _, order := range orders {
		if ShouldClean(order, now) {
			result = append(result, order)
		}
	}
	return result
}
