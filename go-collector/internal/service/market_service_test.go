// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package service

import (
	"context"
	"testing"

	"github.com/brvm/go-collector/internal/domain"
)

// --- fakes ---

type fakeGithub struct {
	tickers []string
	records []domain.MarketRecord
}

func (f *fakeGithub) Tickers() []string { return f.tickers }
func (f *fakeGithub) Fetch(_ context.Context, _ string) ([]domain.MarketRecord, error) {
	return f.records, nil
}

type insertCall struct {
	ticker  string
	source  string
	records []domain.MarketRecord
}

type fakeMarketRepo struct {
	calls []insertCall
}

func (f *fakeMarketRepo) BatchInsertMarketData(_ context.Context, ticker, source string, records []domain.MarketRecord) (int, error) {
	f.calls = append(f.calls, insertCall{ticker: ticker, source: source, records: records})
	return len(records), nil
}

type publishCall struct {
	channel string
	message string
}

type fakePublisher struct {
	calls []publishCall
}

func (f *fakePublisher) Publish(_ context.Context, channel, message string) error {
	f.calls = append(f.calls, publishCall{channel: channel, message: message})
	return nil
}

// --- tests ---

func TestCollectGithubCSV_InsertsAndPublishes(t *testing.T) {
	recs := []domain.MarketRecord{{Open: 1}, {Open: 2}}
	gh := &fakeGithub{tickers: []string{"SNTS", "ORAC"}, records: recs}
	repo := &fakeMarketRepo{}
	pub := &fakePublisher{}

	svc := NewMarketService(gh, nil, repo, pub)
	if err := svc.CollectGithubCSV(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// One insert call per ticker, each receiving the fetched records and the right source tag.
	if len(repo.calls) != 2 {
		t.Fatalf("expected 2 insert calls, got %d", len(repo.calls))
	}
	for _, c := range repo.calls {
		if c.source != "github_brvm_data_public" {
			t.Errorf("source: got %q, want github_brvm_data_public", c.source)
		}
		if len(c.records) != len(recs) {
			t.Errorf("records: got %d, want %d", len(c.records), len(recs))
		}
	}
	if repo.calls[0].ticker != "SNTS" || repo.calls[1].ticker != "ORAC" {
		t.Errorf("tickers: got %q,%q want SNTS,ORAC", repo.calls[0].ticker, repo.calls[1].ticker)
	}

	// Exactly one scan trigger published on the expected channel/message.
	if len(pub.calls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(pub.calls))
	}
	if pub.calls[0].channel != "signals:scan" || pub.calls[0].message != "github_updated" {
		t.Errorf("publish: got %q/%q, want signals:scan/github_updated", pub.calls[0].channel, pub.calls[0].message)
	}
}

type fakeSika struct {
	configured bool
	ticker     string
	records    []domain.MarketRecord
}

func (f *fakeSika) Configured() bool { return f.configured }
func (f *fakeSika) Ticker() string   { return f.ticker }
func (f *fakeSika) Fetch(_ context.Context) ([]domain.MarketRecord, error) {
	return f.records, nil
}

func TestCollectSikafinance_SkipsWhenUnconfigured(t *testing.T) {
	repo := &fakeMarketRepo{}
	pub := &fakePublisher{}
	svc := NewMarketService(&fakeGithub{}, &fakeSika{configured: false}, repo, pub)

	if err := svc.CollectSikafinance(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.calls) != 0 || len(pub.calls) != 0 {
		t.Errorf("expected no inserts/publishes when unconfigured, got %d/%d", len(repo.calls), len(pub.calls))
	}
}

func TestCollectSikafinance_InsertsWithDerivedTicker(t *testing.T) {
	repo := &fakeMarketRepo{}
	pub := &fakePublisher{}
	sika := &fakeSika{configured: true, ticker: "PALC", records: []domain.MarketRecord{{Close: 10}}}
	svc := NewMarketService(&fakeGithub{}, sika, repo, pub)

	if err := svc.CollectSikafinance(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.calls) != 1 {
		t.Fatalf("expected 1 insert call, got %d", len(repo.calls))
	}
	if repo.calls[0].ticker != "PALC" || repo.calls[0].source != "sikafinance" {
		t.Errorf("insert: got ticker=%q source=%q want PALC/sikafinance", repo.calls[0].ticker, repo.calls[0].source)
	}
	if len(pub.calls) != 1 || pub.calls[0].message != "sikafinance_updated" {
		t.Errorf("expected one sikafinance_updated publish, got %+v", pub.calls)
	}
}
