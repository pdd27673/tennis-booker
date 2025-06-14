package retention

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tennis-booker/internal/models"
)

// DoesSlotMatchActivePreferences determines if a court slot matches any of the active user preferences
func DoesSlotMatchActivePreferences(slot models.CourtSlot, activePreferences []models.UserPreferences) (bool, error) {
	// Iterate through all active preferences
	for _, pref := range activePreferences {
		matches, err := doesSlotMatchPreference(slot, pref)
		if err != nil {
			return false, fmt.Errorf("error matching slot against preference %s: %w", pref.ID.Hex(), err)
		}

		// Return true as soon as we find a match
		if matches {
			return true, nil
		}
	}

	// No matches found
	return false, nil
}

// doesSlotMatchPreference checks if a slot matches a single user preference
func doesSlotMatchPreference(slot models.CourtSlot, pref models.UserPreferences) (bool, error) {
	// Check venue preferences (excluded venues take precedence)
	if !matchesVenuePreferences(slot, pref) {
		return false, nil
	}

	// Check day preferences
	if !matchesDayPreferences(slot, pref) {
		return false, nil
	}

	// Check time preferences
	matches, err := matchesTimePreferences(slot, pref)
	if err != nil {
		return false, err
	}
	if !matches {
		return false, nil
	}

	// Check price preferences
	if !matchesPricePreferences(slot, pref) {
		return false, nil
	}

	// All criteria match
	return true, nil
}

// matchesVenuePreferences checks if the slot venue matches the user's venue preferences
func matchesVenuePreferences(slot models.CourtSlot, pref models.UserPreferences) bool {
	venueID := slot.VenueID.Hex()
	venueName := slot.VenueName

	// Check excluded venues first (takes precedence)
	for _, excludedVenue := range pref.ExcludedVenues {
		if excludedVenue == venueID || excludedVenue == venueName {
			return false
		}
	}

	// If no preferred venues specified, any venue is acceptable (as long as not excluded)
	if len(pref.PreferredVenues) == 0 {
		return true
	}

	// Check if venue is in preferred list
	for _, preferredVenue := range pref.PreferredVenues {
		if preferredVenue == venueID || preferredVenue == venueName {
			return true
		}
	}

	// Venue not in preferred list
	return false
}

// matchesDayPreferences checks if the slot day matches the user's day preferences
func matchesDayPreferences(slot models.CourtSlot, pref models.UserPreferences) bool {
	// If no preferred days specified, any day is acceptable
	if len(pref.PreferredDays) == 0 {
		return true
	}

	// Parse the slot date to get the day of week
	slotDay := getWeekdayFromSlot(slot)
	if slotDay == "" {
		// If we can't parse the day, assume it doesn't match
		return false
	}

	// Check if the slot day is in the preferred days
	for _, preferredDay := range pref.PreferredDays {
		if strings.EqualFold(preferredDay, slotDay) {
			return true
		}
	}

	return false
}

// matchesTimePreferences checks if the slot time matches the user's time preferences
func matchesTimePreferences(slot models.CourtSlot, pref models.UserPreferences) (bool, error) {
	// If no time preferences specified, any time is acceptable
	if len(pref.Times) == 0 {
		return true, nil
	}

	// Parse slot start and end times
	slotStart, err := parseTimeString(slot.StartTime)
	if err != nil {
		return false, fmt.Errorf("failed to parse slot start time %s: %w", slot.StartTime, err)
	}

	slotEnd, err := parseTimeString(slot.EndTime)
	if err != nil {
		return false, fmt.Errorf("failed to parse slot end time %s: %w", slot.EndTime, err)
	}

	// Check if slot time overlaps with any preferred time range
	for _, timeRange := range pref.Times {
		prefStart, err := parseTimeString(timeRange.Start)
		if err != nil {
			return false, fmt.Errorf("failed to parse preference start time %s: %w", timeRange.Start, err)
		}

		prefEnd, err := parseTimeString(timeRange.End)
		if err != nil {
			return false, fmt.Errorf("failed to parse preference end time %s: %w", timeRange.End, err)
		}

		// Check for time overlap
		if timesOverlap(slotStart, slotEnd, prefStart, prefEnd) {
			return true, nil
		}
	}

	return false, nil
}

// matchesPricePreferences checks if the slot price matches the user's price preferences
func matchesPricePreferences(slot models.CourtSlot, pref models.UserPreferences) bool {
	// If no max price specified, any price is acceptable
	if pref.MaxPrice <= 0 {
		return true
	}

	// Check if slot price is within the user's budget
	return slot.Price <= pref.MaxPrice
}

// getWeekdayFromSlot extracts the weekday name from a court slot
func getWeekdayFromSlot(slot models.CourtSlot) string {
	// Try to use SlotDate first (time.Time field)
	if !slot.SlotDate.IsZero() {
		return strings.ToLower(slot.SlotDate.Weekday().String())
	}

	// Fallback to parsing the Date string field
	if slot.Date != "" {
		if date, err := time.Parse("2006-01-02", slot.Date); err == nil {
			return strings.ToLower(date.Weekday().String())
		}
	}

	return ""
}

// parseTimeString parses a time string in "HH:MM" format to minutes since midnight
func parseTimeString(timeStr string) (int, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid hours: %s", parts[0])
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid minutes: %s", parts[1])
	}

	if hours < 0 || hours > 23 || minutes < 0 || minutes > 59 {
		return 0, fmt.Errorf("time out of range: %s", timeStr)
	}

	return hours*60 + minutes, nil
}

// timesOverlap checks if two time ranges overlap
func timesOverlap(start1, end1, start2, end2 int) bool {
	// Two ranges overlap if one starts before the other ends
	return start1 < end2 && start2 < end1
}

// MatchingResult represents the result of matching a slot against preferences
type MatchingResult struct {
	Matches       bool
	MatchedUserID string
	MatchReason   string
	Error         error
}

// DoesSlotMatchActivePreferencesDetailed provides detailed matching information
func DoesSlotMatchActivePreferencesDetailed(slot models.CourtSlot, activePreferences []models.UserPreferences) MatchingResult {
	for _, pref := range activePreferences {
		matches, err := doesSlotMatchPreference(slot, pref)
		if err != nil {
			return MatchingResult{
				Matches: false,
				Error:   fmt.Errorf("error matching slot against preference %s: %w", pref.ID.Hex(), err),
			}
		}

		if matches {
			return MatchingResult{
				Matches:       true,
				MatchedUserID: pref.UserID.Hex(),
				MatchReason:   buildMatchReason(slot, pref),
			}
		}
	}

	return MatchingResult{
		Matches:     false,
		MatchReason: "No matching preferences found",
	}
}

// buildMatchReason creates a human-readable explanation of why a slot matched a preference
func buildMatchReason(slot models.CourtSlot, pref models.UserPreferences) string {
	reasons := []string{}

	// Venue matching
	if len(pref.PreferredVenues) > 0 {
		for _, venue := range pref.PreferredVenues {
			if venue == slot.VenueID.Hex() || venue == slot.VenueName {
				reasons = append(reasons, fmt.Sprintf("preferred venue: %s", slot.VenueName))
				break
			}
		}
	}

	// Day matching
	if len(pref.PreferredDays) > 0 {
		slotDay := getWeekdayFromSlot(slot)
		for _, day := range pref.PreferredDays {
			if strings.EqualFold(day, slotDay) {
				reasons = append(reasons, fmt.Sprintf("preferred day: %s", slotDay))
				break
			}
		}
	}

	// Time matching
	if len(pref.Times) > 0 {
		reasons = append(reasons, fmt.Sprintf("preferred time: %s-%s", slot.StartTime, slot.EndTime))
	}

	// Price matching
	if pref.MaxPrice > 0 && slot.Price <= pref.MaxPrice {
		reasons = append(reasons, fmt.Sprintf("within budget: £%.2f <= £%.2f", slot.Price, pref.MaxPrice))
	}

	if len(reasons) == 0 {
		return "matches default preferences"
	}

	return strings.Join(reasons, ", ")
}
