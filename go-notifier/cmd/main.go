// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// Config
var (
	RedisURL = os.Getenv("REDIS_URL")
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
	RDB *redis.Client
}

func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{Addr: RedisURL})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer rdb.Close()

	n := &Notifier{RDB: rdb}

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

		// Send notifications in parallel
		go n.sendFirebasePush(ctx, signal)
		go n.sendEmail(signal)
		go n.sendSMS(signal)
		go n.sendSlack(signal)
	}
}

// sendFirebasePush sends push notification via Firebase Cloud Messaging
func (n *Notifier) sendFirebasePush(ctx context.Context, signal SignalNotification) {
	// TODO: Implement with firebase-admin SDK
	// topics: "signals_ALL", "signals_STRONG_BUY", "signals_{TICKER}"
	log.Printf("[NOTIFIER][FCM] Would send push for %s → %s", signal.Ticker, signal.Signal)
}

// sendEmail sends email notification via SendGrid
func (n *Notifier) sendEmail(signal SignalNotification) {
	// TODO: Implement with SendGrid API
	log.Printf("[NOTIFIER][EMAIL] Would send email for %s → %s", signal.Ticker, signal.Signal)
}

// sendSMS sends SMS notification via Twilio
func (n *Notifier) sendSMS(signal SignalNotification) {
	// TODO: Implement with Twilio API
	log.Printf("[NOTIFIER][SMS] Would send SMS for %s → %s", signal.Ticker, signal.Signal)
}

// sendSlack sends notification to Slack channel
func (n *Notifier) sendSlack(signal SignalNotification) {
	// TODO: Implement with Slack webhook
	log.Printf("[NOTIFIER][SLACK] Would send to Slack for %s → %s", signal.Ticker, signal.Signal)
}
