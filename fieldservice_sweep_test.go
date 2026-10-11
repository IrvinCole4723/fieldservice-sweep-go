package main

import (
	"testing"
	"time"
)

func TestStaleOrders(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	orders := []WorkOrder{
		{ID: "WO-101", PhotoArchived: true, DispatchStatus: "completed", FollowUpDue: now.Add(-time.Hour)},
		{ID: "WO-102", PhotoArchived: false, DispatchStatus: "completed", FollowUpDue: now.Add(-time.Hour)},
		{ID: "WO-103", PhotoArchived: true, DispatchStatus: "en_route", FollowUpDue: now.Add(-time.Hour)},
	}
	got := staleOrders(orders, now)
	if len(got) != 1 || got[0].ID != "WO-101" {
		t.Fatalf("cleaned orders = %#v", got)
	}
}
