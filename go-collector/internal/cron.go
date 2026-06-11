package internal

import (
	"context"
	"log"
	"time"
	"github.com/robfig/cron/v3"
)

func NewCronScheduler(c *Collector, ctx context.Context) *cron.Cron {
	cronScheduler := cron.New(cron.WithLocation(time.FixedZone("GMT", 0)))

	// GitHub CSV: every 15 min, Mon-Fri 9h-15h GMT (BRVM trading hours).
	if _, err := cronScheduler.AddFunc("*/15 9-15 * * 1-5", func() {
		log.Println("[MAIN] Running GitHub CSV collection...")
		if err := c.CollectGithubCSV(ctx); err != nil {
			log.Printf("[MAIN] GitHub CSV error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to add GitHub CSV cron job:", err)
	}

	// Sikafinance: daily at 18h GMT (after market close).
	if _, err := cronScheduler.AddFunc("0 18 * * *", func() {
		log.Println("[MAIN] Running Sikafinance collection...")
		if err := c.CollectSikafinance(ctx); err != nil {
			log.Printf("[MAIN] Sikafinance error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to add Sikafinance cron job:", err)
	}

	// FluxBourse: every 30 min, 9h-15h GMT.
	if _, err := cronScheduler.AddFunc("*/30 9-15 * * *", func() {
		log.Println("[MAIN] Running FluxBourse collection...")
		if err := c.CollectFluxBourse(ctx); err != nil {
			log.Printf("[MAIN] FluxBourse error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to add FluxBourse cron job:", err)
	}

	// Cache cleanup: daily at 2h GMT.
	if _, err := cronScheduler.AddFunc("0 2 * * *", func() {
		log.Println("[MAIN] Running cache cleanup...")
		c.CleanupCache(ctx)
	}); err != nil {
		log.Fatal("Failed to add cache cleanup cron job:", err)
	}

	// BCEAO macroeconomic data: 1st of each month at 8h GMT.
	if _, err := cronScheduler.AddFunc("0 8 1 * *", func() {
		log.Println("[MAIN] Running BCEAO macroeconomic collection...")
		if err := c.CollectBCEAO(ctx); err != nil {
			log.Printf("[MAIN] BCEAO error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to add BCEAO cron job:", err)
	}

	// Commodity prices (cocoa, oil): daily at 17h GMT (after markets close).
	if _, err := cronScheduler.AddFunc("0 17 * * *", func() {
		log.Println("[MAIN] Running commodity price collection...")
		if err := c.CollectCommodities(ctx); err != nil {
			log.Printf("[MAIN] Commodity error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to add commodity cron job:", err)
	}

	// Political risk (ECOWAS, GDELT): daily at 7h GMT (before market opens).
	if _, err := cronScheduler.AddFunc("0 7 * * *", func() {
		log.Println("[MAIN] Running ECOWAS political risk collection...")
		if err := c.CollectPoliticalRisk(ctx); err != nil {
			log.Printf("[MAIN] Political risk error: %v", err)
		}
	}); err != nil {
		log.Fatal("Failed to add political risk cron job:", err)
	}

	cronScheduler.Start()
	log.Println("[MAIN] Scheduler started. Waiting for cron jobs...")
	return cronScheduler
}
