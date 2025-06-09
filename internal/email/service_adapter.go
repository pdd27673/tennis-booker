package email

import (
	"context"
)

// ServiceAdapter adapts the SendGrid email client to the EmailService interface
type ServiceAdapter struct {
	client *Client
}

// NewServiceAdapter creates a new email service adapter
func NewServiceAdapter(client *Client) *ServiceAdapter {
	return &ServiceAdapter{
		client: client,
	}
}

// SendCourtAvailabilityAlert implements the EmailService interface
func (s *ServiceAdapter) SendCourtAvailabilityAlert(ctx context.Context, toEmail string, data *CourtAvailabilityData) error {
	if s.client == nil {
		return nil // Silently skip if no client configured
	}
	
	return s.client.SendCourtAvailabilityAlert(ctx, toEmail, data)
}

// IsEnabled returns true if the email service is properly configured
func (s *ServiceAdapter) IsEnabled() bool {
	return s.client != nil && s.client.IsEnabled()
} 