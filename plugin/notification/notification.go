package notification

import (
	"context"
	"log/slog"
	"sync"
)

// Notification represents a notification to be sent to a user.
type Notification struct {
	// UserID is the ID of the user to notify.
	UserID int32
	// Title is the notification title.
	Title string
	// Body is the notification body/content.
	Body string
	// URL is an optional URL to link to.
	URL string
	// Data contains additional metadata for the notification.
	Data map[string]any
}

// Provider is the interface that notification providers must implement.
type Provider interface {
	// Name returns the unique name of this provider.
	Name() string
	// Send sends a notification to the specified user.
	Send(ctx context.Context, notification *Notification) error
	// IsEnabled returns whether this provider is enabled for the given user.
	// If userID is 0, returns whether the provider is globally enabled.
	IsEnabled(ctx context.Context, userID int32) bool
}

// Service manages notification providers and dispatches notifications.
type Service struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewService creates a new notification service.
func NewService() *Service {
	return &Service{
		providers: make(map[string]Provider),
	}
}

// RegisterProvider registers a notification provider.
// If a provider with the same name already exists, it will be replaced.
func (s *Service) RegisterProvider(provider Provider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[provider.Name()] = provider
	slog.Info("registered notification provider", "name", provider.Name())
}

// UnregisterProvider removes a notification provider.
func (s *Service) UnregisterProvider(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.providers, name)
}

// GetProvider returns a provider by name.
func (s *Service) GetProvider(name string) (Provider, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.providers[name]
	return p, ok
}

// ListProviders returns all registered provider names.
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// Notify sends a notification through all enabled providers.
// It returns the number of providers that successfully sent the notification.
func (s *Service) Notify(ctx context.Context, notification *Notification) int {
	s.mu.RLock()
	providers := make([]Provider, 0, len(s.providers))
	for _, p := range s.providers {
		providers = append(providers, p)
	}
	s.mu.RUnlock()

	slog.Info("notification service: dispatching notification",
		"userID", notification.UserID,
		"title", notification.Title,
		"providerCount", len(providers),
	)

	successCount := 0
	for _, provider := range providers {
		providerName := provider.Name()
		isEnabled := provider.IsEnabled(ctx, notification.UserID)
		slog.Info("notification service: checking provider",
			"provider", providerName,
			"isEnabled", isEnabled,
			"userID", notification.UserID,
		)

		if !isEnabled {
			continue
		}

		if err := provider.Send(ctx, notification); err != nil {
			slog.Error("failed to send notification",
				"provider", providerName,
				"userID", notification.UserID,
				"error", err,
			)
			continue
		}

		slog.Info("notification sent successfully",
			"provider", providerName,
			"userID", notification.UserID,
		)
		successCount++
	}

	slog.Info("notification service: dispatch complete",
		"successCount", successCount,
		"totalProviders", len(providers),
	)

	return successCount
}

// NotifyWithProvider sends a notification through a specific provider.
func (s *Service) NotifyWithProvider(ctx context.Context, providerName string, notification *Notification) error {
	provider, ok := s.GetProvider(providerName)
	if !ok {
		slog.Warn("notification provider not found", "name", providerName)
		return nil
	}

	if !provider.IsEnabled(ctx, notification.UserID) {
		slog.Debug("notification provider disabled for user",
			"provider", providerName,
			"userID", notification.UserID,
		)
		return nil
	}

	return provider.Send(ctx, notification)
}
