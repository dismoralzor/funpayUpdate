// Command funpayupdater periodically bumps a FunPay lot category so it stays
// at the top of the listing.
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dismoralzor/funpayUpdate/internal/funpay"
)

const defaultInterval = 4*time.Hour + 18*time.Minute

func main() {
	log.SetFlags(log.LstdFlags)

	var (
		cfg      funpay.Config
		interval time.Duration
		once     bool
	)
	flag.StringVar(&cfg.Username, "username", os.Getenv("FUNPAY_USERNAME"), "VK username (or FUNPAY_USERNAME)")
	flag.StringVar(&cfg.ChromeDriverPath, "chromedriver", os.Getenv("CHROMEDRIVER_PATH"), "path to chromedriver (default: found on PATH)")
	flag.StringVar(&cfg.LotsURL, "lots-url", "https://funpay.com/lots/1120/trade", "lot category page to bump")
	flag.StringVar(&cfg.ScreenshotPath, "screenshot", "Screenshot.jpg", "where to save a screenshot after each bump (empty to disable)")
	flag.IntVar(&cfg.Port, "port", 5050, "local port for chromedriver")
	flag.DurationVar(&interval, "interval", defaultInterval, "delay between bumps")
	flag.BoolVar(&once, "once", false, "bump a single time and exit")
	flag.Parse()

	// The password is env-only on purpose: command-line flags are visible to
	// every process on the machine via the process list.
	cfg.Password = os.Getenv("FUNPAY_PASSWORD")
	if cfg.Password == "" {
		log.Fatal("FUNPAY_PASSWORD is not set")
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, interval, once); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, cfg funpay.Config, interval time.Duration, once bool) error {
	for {
		if err := funpay.Update(ctx, cfg); err != nil {
			if once || ctx.Err() != nil {
				return err
			}
			// A single failed bump is usually a transient page-layout or
			// network hiccup, so keep the schedule instead of dying.
			log.Printf("bump failed: %v", err)
		} else {
			log.Printf("bumped %s", cfg.LotsURL)
		}

		if once {
			return nil
		}

		select {
		case <-ctx.Done():
			log.Print("shutting down")
			return nil
		case <-time.After(interval):
		}
	}
}
