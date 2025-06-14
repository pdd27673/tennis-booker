package retention

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"tennis-booker/internal/models"
)

// RetentionConfig holds configuration for the retention service
type RetentionConfig struct {
	// RetentionWindow is how old slots must be before they're eligible for deletion
	RetentionWindow time.Duration

	// BatchSize is the maximum number of slots to process in a single batch
	BatchSize int

	// DryRun mode logs what would be deleted without actually deleting
	DryRun bool

	// EnableMetrics controls whether to collect and log detailed metrics
	EnableMetrics bool

	// LogLevel controls the verbosity of logging
	LogLevel string
}

// DefaultRetentionConfig returns a sensible default configuration
func DefaultRetentionConfig() RetentionConfig {
	return RetentionConfig{
		RetentionWindow: 7 * 24 * time.Hour, // 7 days
		BatchSize:       1000,
		DryRun:          false,
		EnableMetrics:   true,
		LogLevel:        "info",
	}
}

// RetentionMetrics holds metrics about a retention cycle
type RetentionMetrics struct {
	StartTime                  time.Time
	EndTime                    time.Time
	Duration                   time.Duration
	CandidateSlotsFound        int
	SlotsCheckedAgainstPrefs   int
	SlotsIdentifiedForDeletion int
	SlotsActuallyDeleted       int
	ActivePreferencesCount     int
	ErrorsEncountered          int
	DryRunMode                 bool
}

// RetentionService orchestrates the intelligent data retention process
type RetentionService struct {
	config            RetentionConfig
	courtSlotService  *models.CourtSlotService
	preferenceService *models.PreferenceService
	logger            *log.Logger
}

// NewRetentionService creates a new retention service with the given configuration
func NewRetentionService(
	config RetentionConfig,
	db *mongo.Database,
	logger *log.Logger,
) *RetentionService {
	if logger == nil {
		logger = log.Default()
	}

	return &RetentionService{
		config:            config,
		courtSlotService:  models.NewCourtSlotService(db),
		preferenceService: models.NewPreferenceService(db),
		logger:            logger,
	}
}

// RunRetentionCycle executes a complete retention cycle
func (s *RetentionService) RunRetentionCycle(ctx context.Context) (*RetentionMetrics, error) {
	metrics := &RetentionMetrics{
		StartTime:  time.Now(),
		DryRunMode: s.config.DryRun,
	}

	s.logInfo("Starting retention cycle", map[string]interface{}{
		"retention_window": s.config.RetentionWindow,
		"batch_size":       s.config.BatchSize,
		"dry_run":          s.config.DryRun,
	})

	// Step 1: Get all active user preferences
	activePreferences, err := s.preferenceService.GetActiveUserPreferences(ctx)
	if err != nil {
		metrics.ErrorsEncountered++
		return metrics, fmt.Errorf("failed to get active user preferences: %w", err)
	}

	metrics.ActivePreferencesCount = len(activePreferences)
	s.logInfo("Retrieved active user preferences", map[string]interface{}{
		"count": len(activePreferences),
	})

	// Step 2: Find candidate slots for deletion
	candidateSlots, err := s.courtSlotService.FindOldUnnotifiedSlots(ctx, s.config.RetentionWindow)
	if err != nil {
		metrics.ErrorsEncountered++
		return metrics, fmt.Errorf("failed to find candidate slots: %w", err)
	}

	metrics.CandidateSlotsFound = len(candidateSlots)
	s.logInfo("Found candidate slots for retention", map[string]interface{}{
		"count":            len(candidateSlots),
		"retention_window": s.config.RetentionWindow,
	})

	if len(candidateSlots) == 0 {
		s.logInfo("No candidate slots found, retention cycle complete", nil)
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		return metrics, nil
	}

	// Step 3: Filter slots that don't match any active preferences
	slotsToDelete := []string{}

	for _, slot := range candidateSlots {
		metrics.SlotsCheckedAgainstPrefs++

		matches, err := DoesSlotMatchActivePreferences(slot, activePreferences)
		if err != nil {
			metrics.ErrorsEncountered++
			s.logError("Error checking slot against preferences", err, map[string]interface{}{
				"slot_id": slot.ID,
			})
			continue
		}

		// If slot doesn't match any active preferences, mark for deletion
		if !matches {
			slotsToDelete = append(slotsToDelete, slot.ID)

			if s.config.LogLevel == "debug" {
				s.logDebug("Slot marked for deletion", map[string]interface{}{
					"slot_id":    slot.ID,
					"venue_name": slot.VenueName,
					"date":       slot.Date,
					"start_time": slot.StartTime,
					"price":      slot.Price,
				})
			}
		}
	}

	metrics.SlotsIdentifiedForDeletion = len(slotsToDelete)
	s.logInfo("Identified slots for deletion", map[string]interface{}{
		"count": len(slotsToDelete),
	})

	// Step 4: Delete slots (or log in dry-run mode)
	if len(slotsToDelete) > 0 {
		if s.config.DryRun {
			s.logInfo("DRY RUN: Would delete the following slots", map[string]interface{}{
				"slot_ids": slotsToDelete,
				"count":    len(slotsToDelete),
			})
			metrics.SlotsActuallyDeleted = 0 // No actual deletion in dry-run
		} else {
			// Process deletions in batches
			deletedCount, err := s.deleteSlotsInBatches(ctx, slotsToDelete)
			if err != nil {
				metrics.ErrorsEncountered++
				return metrics, fmt.Errorf("failed to delete slots: %w", err)
			}

			metrics.SlotsActuallyDeleted = deletedCount
			s.logInfo("Successfully deleted slots", map[string]interface{}{
				"count": deletedCount,
			})
		}
	}

	// Step 5: Complete metrics and logging
	metrics.EndTime = time.Now()
	metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)

	s.logInfo("Retention cycle completed", map[string]interface{}{
		"duration":                      metrics.Duration,
		"candidate_slots":               metrics.CandidateSlotsFound,
		"slots_checked":                 metrics.SlotsCheckedAgainstPrefs,
		"slots_identified_for_deletion": metrics.SlotsIdentifiedForDeletion,
		"slots_actually_deleted":        metrics.SlotsActuallyDeleted,
		"active_preferences":            metrics.ActivePreferencesCount,
		"errors":                        metrics.ErrorsEncountered,
		"dry_run":                       metrics.DryRunMode,
	})

	return metrics, nil
}

// deleteSlotsInBatches deletes slots in configurable batch sizes
func (s *RetentionService) deleteSlotsInBatches(ctx context.Context, slotIDs []string) (int, error) {
	totalDeleted := 0
	batchSize := s.config.BatchSize

	for i := 0; i < len(slotIDs); i += batchSize {
		end := i + batchSize
		if end > len(slotIDs) {
			end = len(slotIDs)
		}

		batch := slotIDs[i:end]
		deletedCount, err := s.courtSlotService.DeleteSlotsByIDs(ctx, batch)
		if err != nil {
			return totalDeleted, fmt.Errorf("failed to delete batch %d-%d: %w", i, end, err)
		}

		totalDeleted += int(deletedCount)

		s.logDebug("Deleted batch of slots", map[string]interface{}{
			"batch_start": i,
			"batch_end":   end,
			"deleted":     deletedCount,
		})
	}

	return totalDeleted, nil
}

// ValidateConfiguration checks if the retention configuration is valid
func (s *RetentionService) ValidateConfiguration() error {
	if s.config.RetentionWindow <= 0 {
		return fmt.Errorf("retention window must be positive, got %v", s.config.RetentionWindow)
	}

	if s.config.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive, got %d", s.config.BatchSize)
	}

	if s.config.BatchSize > 10000 {
		return fmt.Errorf("batch size too large (max 10000), got %d", s.config.BatchSize)
	}

	return nil
}

// GetConfiguration returns the current configuration
func (s *RetentionService) GetConfiguration() RetentionConfig {
	return s.config
}

// UpdateConfiguration updates the service configuration
func (s *RetentionService) UpdateConfiguration(config RetentionConfig) error {
	// Create a temporary service to validate the new config
	tempService := &RetentionService{config: config}
	if err := tempService.ValidateConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	s.config = config
	s.logInfo("Configuration updated", map[string]interface{}{
		"retention_window": config.RetentionWindow,
		"batch_size":       config.BatchSize,
		"dry_run":          config.DryRun,
		"enable_metrics":   config.EnableMetrics,
		"log_level":        config.LogLevel,
	})

	return nil
}

// Logging helper methods
func (s *RetentionService) logInfo(message string, fields map[string]interface{}) {
	s.logWithLevel("INFO", message, fields)
}

func (s *RetentionService) logError(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["error"] = err.Error()
	s.logWithLevel("ERROR", message, fields)
}

func (s *RetentionService) logDebug(message string, fields map[string]interface{}) {
	if s.config.LogLevel == "debug" {
		s.logWithLevel("DEBUG", message, fields)
	}
}

func (s *RetentionService) logWithLevel(level, message string, fields map[string]interface{}) {
	logMessage := fmt.Sprintf("[%s] %s", level, message)
	if fields != nil && len(fields) > 0 {
		logMessage += fmt.Sprintf(" %+v", fields)
	}
	s.logger.Println(logMessage)
}
