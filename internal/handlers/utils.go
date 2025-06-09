package handlers

import (
	"fmt"
	"strconv"
	"time"
)

// calculateDuration calculates duration between start and end times
func calculateDuration(start, end string) string {
	if start == "" || end == "" {
		return ""
	}

	startTime, err := time.Parse("15:04", start)
	if err != nil {
		return ""
	}

	endTime, err := time.Parse("15:04", end)
	if err != nil {
		return ""
	}

	duration := endTime.Sub(startTime)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// formatPrice formats a price with currency symbol
func formatPrice(price float64, currency string) string {
	if currency == "" {
		currency = "GBP"
	}
	
	switch currency {
	case "GBP":
		return "£" + strconv.FormatFloat(price, 'f', 2, 64)
	case "USD":
		return "$" + strconv.FormatFloat(price, 'f', 2, 64)
	case "EUR":
		return "€" + strconv.FormatFloat(price, 'f', 2, 64)
	default:
		return strconv.FormatFloat(price, 'f', 2, 64) + " " + currency
	}
} 