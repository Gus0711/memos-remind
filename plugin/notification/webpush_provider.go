package notification

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/SherClockHolmes/webpush-go"

	storepb "github.com/usememos/memos/proto/gen/store"
	"github.com/usememos/memos/store"
)

const WebPushProviderName = "webpush"

// WebPushProvider sends push notifications via Web Push API.
type WebPushProvider struct {
	store *store.Store
}

// NewWebPushProvider creates a new WebPush notification provider.
func NewWebPushProvider(store *store.Store) *WebPushProvider {
	return &WebPushProvider{
		store: store,
	}
}

// Name returns the provider name.
func (p *WebPushProvider) Name() string {
	return WebPushProviderName
}

// IsEnabled returns whether web push notifications are enabled.
func (p *WebPushProvider) IsEnabled(ctx context.Context, userID int32) bool {
	slog.Debug("webpush IsEnabled check", "userID", userID)

	setting, err := p.store.GetInstanceWebPushSetting(ctx)
	if err != nil {
		slog.Error("failed to get web push setting", "error", err)
		return false
	}
	if setting == nil {
		slog.Debug("webpush disabled: setting is nil")
		return false
	}
	if !setting.Enabled {
		slog.Debug("webpush disabled: setting.Enabled is false")
		return false
	}
	if setting.VapidPrivateKey == "" || setting.VapidPublicKey == "" {
		slog.Debug("webpush disabled: VAPID keys are missing",
			"hasPrivateKey", setting.VapidPrivateKey != "",
			"hasPublicKey", setting.VapidPublicKey != "")
		return false
	}

	// Check if user has any push subscriptions
	if userID > 0 {
		subscriptions, err := p.store.ListPushSubscriptions(ctx, &store.FindPushSubscription{
			UserID: &userID,
		})
		if err != nil {
			slog.Error("failed to list push subscriptions", "error", err)
			return false
		}
		if len(subscriptions) == 0 {
			slog.Debug("webpush disabled: no push subscriptions for user", "userID", userID)
			return false
		}
		slog.Debug("webpush enabled", "userID", userID, "subscriptionCount", len(subscriptions))
		return true
	}

	slog.Debug("webpush enabled (global check)")
	return true
}

// Send sends a push notification to all user's subscribed devices.
func (p *WebPushProvider) Send(ctx context.Context, notification *Notification) error {
	slog.Info("webpush Send called", "userID", notification.UserID, "title", notification.Title)

	setting, err := p.store.GetInstanceWebPushSetting(ctx)
	if err != nil {
		slog.Error("webpush Send: failed to get setting", "error", err)
		return err
	}
	if setting == nil || !setting.Enabled {
		slog.Debug("webpush Send: disabled", "settingNil", setting == nil)
		return nil
	}

	subscriptions, err := p.store.ListPushSubscriptions(ctx, &store.FindPushSubscription{
		UserID: &notification.UserID,
	})
	if err != nil {
		slog.Error("webpush Send: failed to list subscriptions", "error", err)
		return err
	}

	slog.Info("webpush Send: found subscriptions", "count", len(subscriptions), "userID", notification.UserID)

	if len(subscriptions) == 0 {
		slog.Debug("no push subscriptions for user", "userID", notification.UserID)
		return nil
	}

	// Build notification payload
	payload := map[string]any{
		"title": notification.Title,
		"body":  notification.Body,
	}
	if notification.URL != "" {
		payload["url"] = notification.URL
	}
	if notification.Data != nil {
		payload["data"] = notification.Data
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Get instance URL for VAPID subject
	instanceURL := p.getInstanceURL(ctx)

	// Send to all subscriptions
	for _, sub := range subscriptions {
		if err := p.sendToSubscription(ctx, sub, payloadBytes, setting, instanceURL); err != nil {
			slog.Error("failed to send push notification",
				"subscriptionID", sub.ID,
				"userID", notification.UserID,
				"error", err,
			)
			// If subscription is expired or invalid, delete it
			if isSubscriptionGone(err) {
				slog.Info("removing invalid push subscription", "subscriptionID", sub.ID)
				_ = p.store.DeletePushSubscription(ctx, &store.DeletePushSubscription{
					ID: &sub.ID,
				})
			}
		}
	}

	return nil
}

func (p *WebPushProvider) sendToSubscription(ctx context.Context, sub *store.PushSubscription, payload []byte, setting *storepb.InstanceWebPushSetting, instanceURL string) error {
	slog.Info("webpush sendToSubscription: starting",
		"subscriptionID", sub.ID,
		"endpoint", sub.Endpoint,
	)

	// Create webpush subscription
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	// Determine VAPID subject (must be mailto: or https:)
	vapidSubject := instanceURL
	if vapidSubject == "" {
		vapidSubject = "mailto:noreply@memos.local"
	}

	slog.Debug("webpush sendToSubscription: sending",
		"vapidSubject", vapidSubject,
		"hasPublicKey", setting.VapidPublicKey != "",
		"hasPrivateKey", setting.VapidPrivateKey != "",
	)

	// Send the notification
	resp, err := webpush.SendNotificationWithContext(ctx, payload, subscription, &webpush.Options{
		VAPIDPublicKey:  setting.VapidPublicKey,
		VAPIDPrivateKey: setting.VapidPrivateKey,
		Subscriber:      vapidSubject,
		TTL:             86400, // 24 hours
	})
	if err != nil {
		slog.Error("webpush sendToSubscription: send failed", "error", err)
		return err
	}
	defer resp.Body.Close()

	slog.Info("webpush sendToSubscription: response received",
		"statusCode", resp.StatusCode,
		"subscriptionID", sub.ID,
	)

	// Check response status
	if resp.StatusCode >= 400 {
		slog.Warn("push notification failed",
			"status", resp.StatusCode,
			"endpoint", sub.Endpoint,
		)
	} else {
		slog.Info("webpush sendToSubscription: SUCCESS",
			"subscriptionID", sub.ID,
			"statusCode", resp.StatusCode,
		)
	}

	return nil
}

func (p *WebPushProvider) getInstanceURL(ctx context.Context) string {
	// Try to get instance URL from basic settings
	basicSetting, err := p.store.GetInstanceBasicSetting(ctx)
	if err != nil {
		slog.Debug("failed to get basic setting", "error", err)
		return ""
	}
	if basicSetting != nil {
		// Basic setting doesn't have instance URL, try general setting
		// For now, return empty
	}
	return ""
}

// isSubscriptionGone checks if the error indicates the subscription is no longer valid.
func isSubscriptionGone(err error) bool {
	if err == nil {
		return false
	}
	// WebPush library returns specific errors for 404 and 410 status codes
	errStr := err.Error()
	return errStr == "410 Gone" || errStr == "404 Not Found"
}
