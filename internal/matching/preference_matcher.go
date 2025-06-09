package matching

import (
	"log"
	"math"
	"strings"
	"time"

	"tennis-booker/internal/models"
)

// PreferenceMatcher handles complex user preference matching logic
type PreferenceMatcher struct {
	logger *log.Logger
}

// NewPreferenceMatcher creates a new preference matcher
func NewPreferenceMatcher(logger *log.Logger) *PreferenceMatcher {
	return &PreferenceMatcher{
		logger: logger,
	}
}

// MatchResult contains the result of a preference matching operation
type MatchResult struct {
	Matches        bool                     `json:"matches"`
	Score          float64                  `json:"score"` // 0-100 match score
	Reasons        []string                 `json:"reasons"`
	MatchDetails   map[string]interface{}   `json:"match_details"`
	FailureReasons []string                 `json:"failure_reasons,omitempty"`
}

// MatchPreferences determines if a court availability event matches user preferences
func (pm *PreferenceMatcher) MatchPreferences(event models.CourtAvailabilityEvent, user models.UserPreferences) *MatchResult {
	result := &MatchResult{
		Matches:      true,
		Score:        100.0,
		Reasons:      []string{},
		MatchDetails: make(map[string]interface{}),
		FailureReasons: []string{},
	}

	// Check each preference category
	pm.checkVenuePreferences(event, user, result)
	pm.checkTimePreferences(event, user, result)
	pm.checkDayPreferences(event, user, result)
	pm.checkPricePreferences(event, user, result)
	pm.checkCourtTypePreferences(event, user, result)
	pm.checkDurationPreferences(event, user, result)
	pm.checkAdvanceBookingPreferences(event, user, result)

	// If any critical check failed, mark as no match
	if len(result.FailureReasons) > 0 {
		result.Matches = false
		result.Score = 0.0
	}

	return result
}

// checkVenuePreferences checks venue-related preferences
func (pm *PreferenceMatcher) checkVenuePreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	// Check if venue is explicitly excluded
	for _, excluded := range user.ExcludedVenues {
		if excluded == event.VenueID {
			result.FailureReasons = append(result.FailureReasons, "Venue is in user's excluded list")
			return
		}
	}

	// Check preferred venues
	if len(user.PreferredVenues) > 0 {
		found := false
		for _, preferred := range user.PreferredVenues {
			if preferred == event.VenueID {
				found = true
				result.Reasons = append(result.Reasons, "Matches preferred venue")
				result.MatchDetails["venue_preference"] = "preferred"
				break
			}
		}
		if !found {
			result.FailureReasons = append(result.FailureReasons, "Venue not in user's preferred list")
			return
		}
	} else {
		result.MatchDetails["venue_preference"] = "no_restriction"
	}

	// Check location-based preferences (if implemented)
	pm.checkLocationPreferences(event, user, result)
}

// checkLocationPreferences checks location-based matching
func (pm *PreferenceMatcher) checkLocationPreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	// Maximum travel distance check (if user has set a preference)
	if user.MaxTravelDistance > 0 {
		// This would require venue coordinates to calculate distance
		// For now, we'll skip this check but the structure is ready
		result.MatchDetails["location_check"] = "not_implemented"
	}
}

// checkTimePreferences checks time-related preferences
func (pm *PreferenceMatcher) checkTimePreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	if len(user.Times) == 0 {
		result.MatchDetails["time_preference"] = "no_restriction"
		return
	}

	// Parse event start time
	eventStartTime, err := time.Parse("15:04", event.StartTime)
	if err != nil {
		result.FailureReasons = append(result.FailureReasons, "Invalid event start time format")
		return
	}

	bestScore := 0.0
	matchedTimeRange := ""

	for _, timeRange := range user.Times {
		score := pm.calculateTimeScore(eventStartTime, timeRange)
		if score > bestScore {
			bestScore = score
			matchedTimeRange = timeRange.Start + "-" + timeRange.End
		}
	}

	if bestScore > 0 {
		result.Reasons = append(result.Reasons, "Matches preferred time range: "+matchedTimeRange)
		result.MatchDetails["time_score"] = bestScore
		result.MatchDetails["matched_time_range"] = matchedTimeRange
		
		// Adjust overall score based on time match quality
		if bestScore < 100 {
			result.Score *= bestScore / 100.0
		}
	} else {
		result.FailureReasons = append(result.FailureReasons, "No time preferences match")
	}
}

// calculateTimeScore calculates how well an event time matches a user's time range
func (pm *PreferenceMatcher) calculateTimeScore(eventTime time.Time, userTimeRange models.TimeRange) float64 {
	startTime, err := time.Parse("15:04", userTimeRange.Start)
	if err != nil {
		return 0
	}
	
	endTime, err := time.Parse("15:04", userTimeRange.End)
	if err != nil {
		return 0
	}

	// Perfect match if within range
	if (eventTime.Equal(startTime) || eventTime.After(startTime)) && 
	   (eventTime.Equal(endTime) || eventTime.Before(endTime)) {
		return 100.0
	}

	// Calculate proximity score if outside range
	if eventTime.Before(startTime) {
		minutesBefore := startTime.Sub(eventTime).Minutes()
		if minutesBefore <= 30 {
			return math.Max(0, 80 - (minutesBefore * 2)) // Score decreases as time difference increases
		}
	} else if eventTime.After(endTime) {
		minutesAfter := eventTime.Sub(endTime).Minutes()
		if minutesAfter <= 30 {
			return math.Max(0, 80 - (minutesAfter * 2))
		}
	}

	return 0
}

// checkDayPreferences checks day-of-week preferences
func (pm *PreferenceMatcher) checkDayPreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	if len(user.PreferredDays) == 0 {
		result.MatchDetails["day_preference"] = "no_restriction"
		return
	}

	eventDate, err := time.Parse("2006-01-02", event.Date)
	if err != nil {
		result.FailureReasons = append(result.FailureReasons, "Invalid event date format")
		return
	}

	dayName := strings.ToLower(eventDate.Weekday().String())
	
	for _, preferredDay := range user.PreferredDays {
		if strings.ToLower(preferredDay) == dayName {
			result.Reasons = append(result.Reasons, "Matches preferred day: "+preferredDay)
			result.MatchDetails["matched_day"] = preferredDay
			return
		}
	}

	result.FailureReasons = append(result.FailureReasons, "Day not in user's preferred days")
}

// checkPricePreferences checks price-related preferences
func (pm *PreferenceMatcher) checkPricePreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	if user.MaxPrice <= 0 {
		result.MatchDetails["price_preference"] = "no_restriction"
		return
	}

	if event.Price <= user.MaxPrice {
		// Calculate price attractiveness score
		priceScore := 100.0
		if user.MaxPrice > 0 {
			priceScore = ((user.MaxPrice - event.Price) / user.MaxPrice) * 100
			if priceScore < 0 {
				priceScore = 0
			}
		}
		
		result.Reasons = append(result.Reasons, "Within price budget")
		result.MatchDetails["price_score"] = priceScore
		result.MatchDetails["price_within_budget"] = true
		
		// Boost score for great deals
		if event.Price <= user.MaxPrice * 0.7 { // 30% below max
			result.Score += 10 // Bonus points for good deals
			result.Reasons = append(result.Reasons, "Great price deal")
		}
	} else {
		result.FailureReasons = append(result.FailureReasons, 
			"Price exceeds maximum budget")
		result.MatchDetails["price_within_budget"] = false
	}
}

// checkCourtTypePreferences checks court surface type preferences
func (pm *PreferenceMatcher) checkCourtTypePreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	if len(user.PreferredCourtTypes) == 0 {
		result.MatchDetails["court_type_preference"] = "no_restriction"
		return
	}

	// Extract court type from court name (simple heuristic)
	courtType := pm.extractCourtType(event.CourtName)
	
	if courtType == "" {
		result.MatchDetails["court_type_preference"] = "unknown_type"
		return
	}

	for _, preferredType := range user.PreferredCourtTypes {
		if strings.ToLower(preferredType) == strings.ToLower(courtType) {
			result.Reasons = append(result.Reasons, "Matches preferred court type: "+preferredType)
			result.MatchDetails["matched_court_type"] = preferredType
			return
		}
	}

	// Court type doesn't match but don't fail - just note it
	result.MatchDetails["court_type_mismatch"] = true
	result.Score *= 0.9 // Small penalty for non-preferred court type
}

// extractCourtType attempts to extract court type from court name
func (pm *PreferenceMatcher) extractCourtType(courtName string) string {
	name := strings.ToLower(courtName)
	
	if strings.Contains(name, "hard") {
		return "hard"
	}
	if strings.Contains(name, "clay") {
		return "clay"
	}
	if strings.Contains(name, "grass") {
		return "grass"
	}
	if strings.Contains(name, "carpet") {
		return "carpet"
	}
	if strings.Contains(name, "indoor") {
		return "indoor"
	}
	if strings.Contains(name, "outdoor") {
		return "outdoor"
	}
	
	return ""
}

// checkDurationPreferences checks session duration preferences
func (pm *PreferenceMatcher) checkDurationPreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	if user.PreferredDuration <= 0 {
		result.MatchDetails["duration_preference"] = "no_restriction"
		return
	}

	// Calculate session duration
	startTime, err := time.Parse("15:04", event.StartTime)
	if err != nil {
		return
	}
	
	endTime, err := time.Parse("15:04", event.EndTime)
	if err != nil {
		return
	}

	sessionDuration := endTime.Sub(startTime).Minutes()
	preferredDurationMinutes := float64(user.PreferredDuration)

	// Check if duration matches (within 15 minutes tolerance)
	durationDiff := math.Abs(sessionDuration - preferredDurationMinutes)
	
	if durationDiff <= 15 {
		result.Reasons = append(result.Reasons, "Matches preferred session duration")
		result.MatchDetails["duration_match"] = true
	} else {
		// Don't fail, but adjust score based on duration difference
		durationScore := math.Max(0, 100 - (durationDiff * 2)) // 2 points off per minute difference
		result.Score *= durationScore / 100.0
		result.MatchDetails["duration_score"] = durationScore
	}
	
	result.MatchDetails["session_duration_minutes"] = sessionDuration
}

// checkAdvanceBookingPreferences checks advance booking timing preferences
func (pm *PreferenceMatcher) checkAdvanceBookingPreferences(event models.CourtAvailabilityEvent, user models.UserPreferences, result *MatchResult) {
	// Parse event date
	eventDate, err := time.Parse("2006-01-02", event.Date)
	if err != nil {
		return
	}

	now := time.Now()
	daysInAdvance := eventDate.Sub(now).Hours() / 24

	// Check minimum advance booking time
	if user.MinAdvanceBookingHours > 0 {
		minHours := float64(user.MinAdvanceBookingHours)
		hoursInAdvance := eventDate.Sub(now).Hours()
		
		if hoursInAdvance < minHours {
			result.FailureReasons = append(result.FailureReasons, 
				"Not enough advance booking time")
			return
		}
	}

	// Check maximum advance booking time
	if user.MaxAdvanceBookingDays > 0 {
		maxDays := float64(user.MaxAdvanceBookingDays)
		
		if daysInAdvance > maxDays {
			result.FailureReasons = append(result.FailureReasons, 
				"Too far in advance")
			return
		}
	}

	result.MatchDetails["days_in_advance"] = daysInAdvance
	
	// Prefer slots that are not too last-minute or too far out
	if daysInAdvance >= 1 && daysInAdvance <= 7 {
		result.Reasons = append(result.Reasons, "Good advance booking timing")
	}
}

// GetMatchingSummary returns a human-readable summary of the matching result
func (pm *PreferenceMatcher) GetMatchingSummary(result *MatchResult) string {
	if !result.Matches {
		return "Does not match user preferences: " + strings.Join(result.FailureReasons, ", ")
	}

	summary := "Matches user preferences"
	if len(result.Reasons) > 0 {
		summary += " (" + strings.Join(result.Reasons, ", ") + ")"
	}
	
	if result.Score < 100 {
		summary += " with " + strings.TrimSuffix(strings.TrimSuffix(
			string(rune(int(result.Score))), "0"), ".") + "% compatibility"
	}

	return summary
}

// GetTopMatches sorts and returns the best matching events for a user
func (pm *PreferenceMatcher) GetTopMatches(events []models.CourtAvailabilityEvent, user models.UserPreferences, limit int) []struct {
	Event  models.CourtAvailabilityEvent
	Result *MatchResult
} {
	type EventMatch struct {
		Event  models.CourtAvailabilityEvent
		Result *MatchResult
	}

	var matches []EventMatch
	
	// Check each event
	for _, event := range events {
		result := pm.MatchPreferences(event, user)
		if result.Matches {
			matches = append(matches, EventMatch{
				Event:  event,
				Result: result,
			})
		}
	}

	// Sort by score (descending)
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Result.Score > matches[i].Result.Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Return top matches up to limit
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}

	return matches
} 