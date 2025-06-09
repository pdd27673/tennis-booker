package email

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Client wraps SendGrid functionality for tennis court notifications
type Client struct {
	sgClient   *sendgrid.Client
	fromEmail  string
	fromName   string
	templateID string // SendGrid template ID for court availability
}

// Config holds email client configuration
type Config struct {
	APIKey     string
	FromEmail  string
	FromName   string
	TemplateID string
}

// NewClient creates a new email client with SendGrid
func NewClient(config *Config) *Client {
	if config.APIKey == "" {
		log.Printf("Warning: SendGrid API key not set, email notifications will be disabled")
		return nil
	}

	return &Client{
		sgClient:   sendgrid.NewSendClient(config.APIKey),
		fromEmail:  config.FromEmail,
		fromName:   config.FromName,
		templateID: config.TemplateID,
	}
}

// NewClientFromEnv creates a new email client using environment variables
func NewClientFromEnv() *Client {
	config := &Config{
		APIKey:     os.Getenv("SENDGRID_API_KEY"),
		FromEmail:  getEnvWithDefault("SENDGRID_FROM_EMAIL", "alerts@tennisbooker.com"),
		FromName:   getEnvWithDefault("SENDGRID_FROM_NAME", "Tennis Court Alerts"),
		TemplateID: os.Getenv("SENDGRID_TEMPLATE_ID"),
	}

	return NewClient(config)
}

// CourtAvailabilityData represents data for court availability email template
type CourtAvailabilityData struct {
	UserName          string    `json:"user_name"`
	VenueName         string    `json:"venue_name"`
	VenueLocation     string    `json:"venue_location"`
	CourtType         string    `json:"court_type"`
	AvailableDate     string    `json:"available_date"`
	AvailableTime     string    `json:"available_time"`
	Duration          string    `json:"duration"`
	Price             string    `json:"price,omitempty"`
	BookingURL        string    `json:"booking_url"`
	UnsubscribeURL    string    `json:"unsubscribe_url"`
	AlertType         string    `json:"alert_type"` // "new_slot", "cancellation", "price_drop"
	NotificationTime  string    `json:"notification_time"`
}

// SendCourtAvailabilityAlert sends a court availability notification
func (c *Client) SendCourtAvailabilityAlert(ctx context.Context, toEmail string, data *CourtAvailabilityData) error {
	if c == nil {
		return fmt.Errorf("email client not initialized")
	}

	from := mail.NewEmail(c.fromName, c.fromEmail)
	to := mail.NewEmail("", toEmail)
	
	var message *mail.SGMailV3

	if c.templateID != "" {
		// Use SendGrid template
		message = mail.NewV3Mail()
		message.SetFrom(from)
		message.SetTemplateID(c.templateID)

		personalization := mail.NewPersonalization()
		personalization.AddTos(to)
		
		// Add template data
		personalization.SetDynamicTemplateData("user_name", data.UserName)
		personalization.SetDynamicTemplateData("venue_name", data.VenueName)
		personalization.SetDynamicTemplateData("venue_location", data.VenueLocation)
		personalization.SetDynamicTemplateData("court_type", data.CourtType)
		personalization.SetDynamicTemplateData("available_date", data.AvailableDate)
		personalization.SetDynamicTemplateData("available_time", data.AvailableTime)
		personalization.SetDynamicTemplateData("duration", data.Duration)
		personalization.SetDynamicTemplateData("price", data.Price)
		personalization.SetDynamicTemplateData("booking_url", data.BookingURL)
		personalization.SetDynamicTemplateData("unsubscribe_url", data.UnsubscribeURL)
		personalization.SetDynamicTemplateData("alert_type", data.AlertType)
		personalization.SetDynamicTemplateData("notification_time", data.NotificationTime)

		message.AddPersonalizations(personalization)
	} else {
		// Fallback to plain text email
		subject := fmt.Sprintf("🎾 Court Available: %s at %s", data.AvailableTime, data.VenueName)
		content := c.generatePlainTextContent(data)
		message = mail.NewSingleEmail(from, subject, to, content, "")
	}

	// Send with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	response, err := c.sgClient.SendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid API error: status %d, body: %s", response.StatusCode, response.Body)
	}

	log.Printf("Court availability alert sent successfully to %s (status: %d)", toEmail, response.StatusCode)
	return nil
}

// generatePlainTextContent creates fallback plain text email content
func (c *Client) generatePlainTextContent(data *CourtAvailabilityData) string {
	alertTypeText := map[string]string{
		"new_slot":     "New court slot available",
		"cancellation": "Court slot freed up due to cancellation",
		"price_drop":   "Price drop on court booking",
	}

	alertText, exists := alertTypeText[data.AlertType]
	if !exists {
		alertText = "Court availability update"
	}

	content := fmt.Sprintf(`🎾 %s

Hi %s,

Great news! A tennis court that matches your preferences is now available:

📍 Venue: %s (%s)
🏟️  Court Type: %s
📅 Date: %s
⏰ Time: %s
⏱️  Duration: %s`, 
		alertText, data.UserName, data.VenueName, data.VenueLocation,
		data.CourtType, data.AvailableDate, data.AvailableTime, data.Duration)

	if data.Price != "" {
		content += fmt.Sprintf("\n💰 Price: %s", data.Price)
	}

	content += fmt.Sprintf(`

🔗 Book now: %s

This alert was sent at %s. Don't wait too long - popular slots book quickly!

---
To stop receiving these alerts, click here: %s
Tennis Court Availability Alert System`, 
		data.BookingURL, data.NotificationTime, data.UnsubscribeURL)

	return content
}

// SendTestEmail sends a test email to verify configuration
func (c *Client) SendTestEmail(ctx context.Context, toEmail string) error {
	if c == nil {
		return fmt.Errorf("email client not initialized")
	}

	testData := &CourtAvailabilityData{
		UserName:         "Test User",
		VenueName:        "Test Tennis Club",
		VenueLocation:    "Test Location",
		CourtType:        "Hard Court",
		AvailableDate:    time.Now().Format("Monday, January 2, 2006"),
		AvailableTime:    "3:00 PM - 4:00 PM",
		Duration:         "1 hour",
		Price:            "£15.00",
		BookingURL:       "https://example.com/book/test",
		UnsubscribeURL:   "https://example.com/unsubscribe",
		AlertType:        "new_slot",
		NotificationTime: time.Now().Format("15:04 on January 2, 2006"),
	}

	return c.SendCourtAvailabilityAlert(ctx, toEmail, testData)
}

// IsEnabled returns true if the email client is properly configured
func (c *Client) IsEnabled() bool {
	return c != nil && c.sgClient != nil
}

// getEnvWithDefault returns environment variable value or default
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 