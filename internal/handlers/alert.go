package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"tennis-booking-bot/internal/models"
)

// AlertHandlers handles alert-related endpoints
type AlertHandlers struct {
	db           *mongo.Database
	alertService *models.AlertHistoryService
}

// NewAlertHandlers creates a new alert handlers instance
func NewAlertHandlers(db *mongo.Database) *AlertHandlers {
	return &AlertHandlers{
		db:           db,
		alertService: models.NewAlertHistoryService(db),
	}
}

// AlertHistoryResponse represents the response for alert history
type AlertHistoryResponse struct {
	Data       []AlertHistoryData `json:"data"`
	Count      int                `json:"count"`
	UserID     string             `json:"user_id"`
	TotalSent  int64              `json:"total_sent"`
	DateRange  DateRange          `json:"date_range"`
	Status     string             `json:"status"`
}

// AlertStatsResponse represents the response for alert statistics
type AlertStatsResponse struct {
	UserID             string              `json:"user_id"`
	TotalAlerts        int64               `json:"total_alerts"`
	RecentAlerts24h    int64               `json:"recent_alerts_24h"`
	RecentAlerts7d     int64               `json:"recent_alerts_7d"`
	EmailStatusCounts  EmailStatusCounts   `json:"email_status_counts"`
	VenueBreakdown     []VenueAlertCount   `json:"venue_breakdown"`
	DailyBreakdown     []DailyAlertCount   `json:"daily_breakdown"`
	LastAlert          *time.Time          `json:"last_alert,omitempty"`
	Status             string              `json:"status"`
}

// AlertHistoryData represents enriched alert history data
type AlertHistoryData struct {
	ID             primitive.ObjectID `json:"id"`
	VenueID        string             `json:"venue_id"`
	VenueName      string             `json:"venue_name"`
	CourtID        string             `json:"court_id"`
	CourtName      string             `json:"court_name"`
	SlotDate       string             `json:"slot_date"`
	SlotTime       string             `json:"slot_time"`       // Combined start-end time
	SlotStartTime  string             `json:"slot_start_time"`
	SlotEndTime    string             `json:"slot_end_time"`
	Duration       string             `json:"duration"`
	Price          float64            `json:"price"`
	FormattedPrice string             `json:"formatted_price"`
	Currency       string             `json:"currency"`
	BookingURL     string             `json:"booking_url"`
	EmailAddress   string             `json:"email_address"`
	AlertSentAt    time.Time          `json:"alert_sent_at"`
	EmailStatus    string             `json:"email_status"`
	DaysAgo        int                `json:"days_ago"`
	IsRecent       bool               `json:"is_recent"` // Within last 24h
}

// EmailStatusCounts represents counts by email delivery status
type EmailStatusCounts struct {
	Sent      int64 `json:"sent"`
	Delivered int64 `json:"delivered"`
	Failed    int64 `json:"failed"`
	Bounced   int64 `json:"bounced"`
	Pending   int64 `json:"pending"`
}

// VenueAlertCount represents alert count per venue
type VenueAlertCount struct {
	VenueID   string `json:"venue_id"`
	VenueName string `json:"venue_name"`
	Count     int64  `json:"count"`
}

// DailyAlertCount represents alert count per day
type DailyAlertCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// DateRange represents a date range
type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// GetAlertHistory handles GET /api/v1/alerts/history
func (h *AlertHandlers) GetAlertHistory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Parse and validate user_id
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_user_id",
			Message: "user_id query parameter is required",
			Status:  "error",
		})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Status:  "error",
		})
		return
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filter
	filter := bson.M{"user_id": userID}

	// Add date range filter if provided
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if dateTo := c.Query("date_to"); dateTo != "" {
			filter["slot_date"] = bson.M{
				"$gte": dateFrom,
				"$lte": dateTo,
			}
		} else {
			filter["slot_date"] = bson.M{"$gte": dateFrom}
		}
	}

	// Add venue filter if provided
	if venueID := c.Query("venue_id"); venueID != "" {
		filter["venue_id"] = venueID
	}

	// Add email status filter if provided
	if emailStatus := c.Query("email_status"); emailStatus != "" {
		filter["email_status"] = emailStatus
	}

	// Query alert history
	collection := h.db.Collection("alert_history")
	
	// Get total count for pagination info
	totalCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to count alerts",
			Status:  "error",
		})
		return
	}

	// Find alerts with pagination
	findOptions := options.Find().
		SetSort(bson.D{{Key: "alert_sent_at", Value: -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve alert history",
			Status:  "error",
		})
		return
	}
	defer cursor.Close(ctx)

	var alerts []models.AlertHistory
	if err = cursor.All(ctx, &alerts); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "decode_error",
			Message: "Failed to decode alert history",
			Status:  "error",
		})
		return
	}

	// Transform to enriched alert history data
	var alertData []AlertHistoryData
	now := time.Now()

	for _, alert := range alerts {
		// Calculate days ago
		daysAgo := int(now.Sub(alert.AlertSentAt).Hours() / 24)
		isRecent := now.Sub(alert.AlertSentAt).Hours() < 24

		// Format price
		formattedPrice := formatPrice(alert.Price, alert.Currency)

		// Calculate duration
		duration := calculateDuration(alert.SlotStartTime, alert.SlotEndTime)

		// Combine slot time
		slotTime := alert.SlotStartTime + "-" + alert.SlotEndTime

		alertData = append(alertData, AlertHistoryData{
			ID:             alert.ID,
			VenueID:        alert.VenueID,
			VenueName:      alert.VenueName,
			CourtID:        alert.CourtID,
			CourtName:      alert.CourtName,
			SlotDate:       alert.SlotDate,
			SlotTime:       slotTime,
			SlotStartTime:  alert.SlotStartTime,
			SlotEndTime:    alert.SlotEndTime,
			Duration:       duration,
			Price:          alert.Price,
			FormattedPrice: formattedPrice,
			Currency:       alert.Currency,
			BookingURL:     alert.BookingURL,
			EmailAddress:   alert.EmailAddress,
			AlertSentAt:    alert.AlertSentAt,
			EmailStatus:    alert.EmailStatus,
			DaysAgo:        daysAgo,
			IsRecent:       isRecent,
		})
	}

	// Ensure we return an empty array instead of null
	if alertData == nil {
		alertData = []AlertHistoryData{}
	}

	// Determine date range
	dateRange := DateRange{}
	if len(alerts) > 0 {
		// Get earliest and latest dates
		earliest := alerts[len(alerts)-1].AlertSentAt
		latest := alerts[0].AlertSentAt
		dateRange.From = earliest.Format("2006-01-02")
		dateRange.To = latest.Format("2006-01-02")
	}

	c.JSON(http.StatusOK, AlertHistoryResponse{
		Data:      alertData,
		Count:     len(alertData),
		UserID:    userIDStr,
		TotalSent: totalCount,
		DateRange: dateRange,
		Status:    "success",
	})
}

// GetAlertStats handles GET /api/v1/alerts/stats
func (h *AlertHandlers) GetAlertStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Parse and validate user_id
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_user_id",
			Message: "user_id query parameter is required",
			Status:  "error",
		})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
			Status:  "error",
		})
		return
	}

	collection := h.db.Collection("alert_history")
	userFilter := bson.M{"user_id": userID}

	// Get total alert count
	totalAlerts, err := collection.CountDocuments(ctx, userFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Failed to count total alerts",
			Status:  "error",
		})
		return
	}

	// Get recent alert counts
	now := time.Now()
	recent24h, _ := collection.CountDocuments(ctx, bson.M{
		"user_id":       userID,
		"alert_sent_at": bson.M{"$gte": now.Add(-24 * time.Hour)},
	})
	
	recent7d, _ := collection.CountDocuments(ctx, bson.M{
		"user_id":       userID,
		"alert_sent_at": bson.M{"$gte": now.Add(-7 * 24 * time.Hour)},
	})

	// Get email status breakdown
	emailStatusCounts := h.getEmailStatusCounts(ctx, collection, userID)

	// Get venue breakdown
	venueBreakdown := h.getVenueBreakdown(ctx, collection, userID)

	// Get daily breakdown (last 30 days)
	dailyBreakdown := h.getDailyBreakdown(ctx, collection, userID, 30)

	// Get last alert timestamp
	var lastAlert *time.Time
	var lastAlertDoc models.AlertHistory
	findOptions := options.FindOne().SetSort(bson.D{{Key: "alert_sent_at", Value: -1}})
	if err := collection.FindOne(ctx, userFilter, findOptions).Decode(&lastAlertDoc); err == nil {
		lastAlert = &lastAlertDoc.AlertSentAt
	}

	c.JSON(http.StatusOK, AlertStatsResponse{
		UserID:            userIDStr,
		TotalAlerts:       totalAlerts,
		RecentAlerts24h:   recent24h,
		RecentAlerts7d:    recent7d,
		EmailStatusCounts: emailStatusCounts,
		VenueBreakdown:    venueBreakdown,
		DailyBreakdown:    dailyBreakdown,
		LastAlert:         lastAlert,
		Status:            "success",
	})
}

// Helper functions

func (h *AlertHandlers) getEmailStatusCounts(ctx context.Context, collection *mongo.Collection, userID primitive.ObjectID) EmailStatusCounts {
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":   "$email_status",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return EmailStatusCounts{}
	}
	defer cursor.Close(ctx)

	counts := EmailStatusCounts{}
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			switch result.ID {
			case "sent":
				counts.Sent = result.Count
			case "delivered":
				counts.Delivered = result.Count
			case "failed":
				counts.Failed = result.Count
			case "bounced":
				counts.Bounced = result.Count
			case "pending":
				counts.Pending = result.Count
			}
		}
	}

	return counts
}

func (h *AlertHandlers) getVenueBreakdown(ctx context.Context, collection *mongo.Collection, userID primitive.ObjectID) []VenueAlertCount {
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":        "$venue_id",
			"venue_name": bson.M{"$first": "$venue_name"},
			"count":      bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": 10}, // Top 10 venues
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return []VenueAlertCount{}
	}
	defer cursor.Close(ctx)

	var breakdown []VenueAlertCount
	for cursor.Next(ctx) {
		var result struct {
			ID        string `bson:"_id"`
			VenueName string `bson:"venue_name"`
			Count     int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			breakdown = append(breakdown, VenueAlertCount{
				VenueID:   result.ID,
				VenueName: result.VenueName,
				Count:     result.Count,
			})
		}
	}

	return breakdown
}

func (h *AlertHandlers) getDailyBreakdown(ctx context.Context, collection *mongo.Collection, userID primitive.ObjectID, days int) []DailyAlertCount {
	since := time.Now().AddDate(0, 0, -days)
	
	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":       userID,
			"alert_sent_at": bson.M{"$gte": since},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{
					"format": "%Y-%m-%d",
					"date":   "$alert_sent_at",
				},
			},
			"count": bson.M{"$sum": 1},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return []DailyAlertCount{}
	}
	defer cursor.Close(ctx)

	var breakdown []DailyAlertCount
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			breakdown = append(breakdown, DailyAlertCount{
				Date:  result.ID,
				Count: result.Count,
			})
		}
	}

	return breakdown
}



 