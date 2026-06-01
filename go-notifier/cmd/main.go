// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config
var (
	RedisURL        = os.Getenv("REDIS_URL")
	SendGridAPIKey  = os.Getenv("SENDGRID_API_KEY")
	SlackWebhookURL = os.Getenv("SLACK_WEBHOOK_URL")
)

// SignalNotification represents a signal event to notify about
type SignalNotification struct {
	Ticker         string   `json:"ticker"`
	Signal         string   `json:"signal"`
	CompositeScore float64  `json:"composite_score"`
	Confidence     float64  `json:"confidence"`
	Reasons        []string `json:"reasons"`
}

// Notifier handles push notifications, SMS, email, and Slack
type Notifier struct {
	RDB  *redis.Client
	HTTP *http.Client
}

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{Addr: RedisURL})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer rdb.Close()

	n := &Notifier{
		RDB:  rdb,
		HTTP: &http.Client{Timeout: 10 * time.Second},
	}

	// Subscribe to signal notifications
	pubsub := rdb.Subscribe(ctx, "signals:notify")
	defer pubsub.Close()

	ch := pubsub.Channel()
	log.Println("[NOTIFIER] Listening for signal notifications...")

	for msg := range ch {
		var signal SignalNotification
		if err := json.Unmarshal([]byte(msg.Payload), &signal); err != nil {
			log.Printf("[NOTIFIER] Invalid message: %v", err)
			continue
		}

		log.Printf("[NOTIFIER] Signal: %s → %s (score: %.1f)", signal.Ticker, signal.Signal, signal.CompositeScore)

		// Process notifications without blocking the main event loop
		go func(sig SignalNotification) {
			notifyCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			var wg sync.WaitGroup
			senders := []func(context.Context, SignalNotification){
				n.sendFirebasePush,
				n.sendEmail,
				n.sendSMS,
				n.sendSlack,
			}

			for _, sender := range senders {
				wg.Add(1)
				go func(s func(context.Context, SignalNotification)) {
					defer wg.Done()
					s(notifyCtx, sig)
				}(sender)
			}
			wg.Wait()
		}(signal)
	}
}

// sendFirebasePush sends push notification via Firebase Cloud Messaging
func (n *Notifier) sendFirebasePush(ctx context.Context, signal SignalNotification) {
	// TODO: Implement with firebase-admin SDK
	log.Printf("[NOTIFIER][FCM] Would send push for %s → %s", signal.Ticker, signal.Signal)
}

// sendEmail sends email notification via SendGrid
func (n *Notifier) sendEmail(ctx context.Context, signal SignalNotification) {
	if SendGridAPIKey == "" {
		log.Printf("[NOTIFIER][EMAIL] Skipped (no SENDGRID_API_KEY)")
		return
	}

	body := fmt.Sprintf(`{
		"personalizations": [{"to": [{"email": "investors@brvm-trading.local"}]}],
		"from": {"email": "alerts@brvm-trading.local"},
		"subject": "BRVM Alert: %s - %s",
		"content": [{"type": "text/plain", "value": "Signal: %s\nScore: %.1f\nConfidence: %.2f\nReasons: %v"}]
	}`, signal.Ticker, signal.Signal, signal.Signal, signal.CompositeScore, signal.Confidence, signal.Reasons)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewBufferString(body))
	req.Header.Set("Authorization", "Bearer "+SendGridAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.HTTP.Do(req)
	if err != nil {
		log.Printf("[NOTIFIER][EMAIL] Failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[NOTIFIER][EMAIL] Sent email for %s (Status: %s)", signal.Ticker, resp.Status)
}

// sendSMS sends SMS notification via Twilio
func (n *Notifier) sendSMS(ctx context.Context, signal SignalNotification) {
	// TODO: Implement with Twilio API
	log.Printf("[NOTIFIER][SMS] Would send SMS for %s → %s", signal.Ticker, signal.Signal)
}

// sendSlack sends notification to Slack channel
func (n *Notifier) sendSlack(ctx context.Context, signal SignalNotification) {
	if SlackWebhookURL == "" {
		log.Printf("[NOTIFIER][SLACK] Skipped (no SLACK_WEBHOOK_URL)")
		return
	}

	msg := map[string]string{
		"text": fmt.Sprintf("🚨 *BRVM Alert* 🚨\n*Ticker:* %s\n*Signal:* %s\n*Score:* %.1f", signal.Ticker, signal.Signal, signal.CompositeScore),
	}
	payload, _ := json.Marshal(msg)

	req, _ := http.NewRequestWithContext(ctx, "POST", SlackWebhookURL, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.HTTP.Do(req)
	if err != nil {
		log.Printf("[NOTIFIER][SLACK] Failed: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[NOTIFIER][SLACK] Sent to Slack for %s (Status: %s)", signal.Ticker, resp.Status)
}
