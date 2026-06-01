package model

import (
	"log"
	"os"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Config — all secrets from env vars, never hardcoded
var (
	Port       = GetEnv("PORT", "8080")
	JWTSecret  = GetEnv("JWT_SECRET", "")
	RateLimit  = 100
	RateWindow = 60 * time.Second
	TickerRe   = regexp.MustCompile(`^[A-Z]{2,10}$`)
)

func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if fallback == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return fallback
}

// AppContext holds shared dependencies for all layers.
type AppContext struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

// Ticker represents a BRVM-listed security.
type Ticker struct {
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Sector  string `json:"sector"`
}

// MarketDataPoint is a single OHLCV bar.
type MarketDataPoint struct {
	Date        string  `json:"date"`
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	Volume      int64   `json:"volume"`
	DailyReturn float64 `json:"daily_return"`
	Source      string  `json:"source"`
}

// FundamentalData holds valuation metrics for a ticker.
type FundamentalData struct {
	Ticker            string    `json:"ticker"`
	PER               *float64  `json:"per,omitempty"`
	ROE               *float64  `json:"roe,omitempty"`
	DividendYield     *float64  `json:"dividend_yield,omitempty"`
	EPS               *float64  `json:"eps,omitempty"`
	BookValuePerShare *float64  `json:"book_value_per_share,omitempty"`
	DebtToEquity      *float64  `json:"debt_to_equity,omitempty"`
	RevenueGrowth     *float64  `json:"revenue_growth,omitempty"`
	NetProfit         *float64  `json:"net_profit,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// MacroData holds macroeconomic indicators for scoring.
type MacroData struct {
	Inflation          *float64 `json:"inflation,omitempty"`
	TauxDirecteur      *float64 `json:"taux_directeur,omitempty"`
	ChangeXofEur       *float64 `json:"change_xof_eur,omitempty"`
	CocoaPrice         *float64 `json:"cocoa_price,omitempty"`
	OilPrice           *float64 `json:"oil_price,omitempty"`
	PoliticalStability *float64 `json:"political_stability,omitempty"`
	SovereignRating    *int32   `json:"sovereign_rating,omitempty"`
	ChangeXofUsd       *float64 `json:"change_xof_usd,omitempty"`
	CashewPrice        *float64 `json:"cashew_price,omitempty"`
	GoldPrice          *float64 `json:"gold_price,omitempty"`
	RubberPrice        *float64 `json:"rubber_price,omitempty"`
	PalmOilPrice       *float64 `json:"palm_oil_price,omitempty"`
	PoliticalCrisis    bool     `json:"political_crisis"`
	CommodityBeta      *float64 `json:"commodity_beta,omitempty"`
}

// SignalResult is the output of the scoring engine.
type SignalResult struct {
	Ticker               Ticker                 `json:"ticker"`
	CompositeScore       float64                `json:"composite_score"`
	Signal               string                 `json:"signal"`
	Confidence           float64                `json:"confidence"`
	Reasons              []string               `json:"reasons"`
	TechnicalIndicators  map[string]interface{} `json:"technical_indicators"`
	FundamentalScores    map[string]float64     `json:"fundamental_scores"`
	MacroScore           float64                `json:"macro_score,omitempty"`
	MacroReasons         []string               `json:"macro_reasons,omitempty"`
	SeasonalityScore     float64                `json:"seasonality_score,omitempty"`
	KellyFraction        float64                `json:"kelly_fraction,omitempty"`
	SuggestedPositionPct float64                `json:"suggested_position_pct,omitempty"`
	Timestamp            time.Time              `json:"timestamp"`
}

// PortfolioPosition represents a user's holding.
type PortfolioPosition struct {
	Ticker           string  `json:"ticker"`
	Quantity         int     `json:"quantity"`
	AvgBuyPrice      float64 `json:"avg_buy_price"`
	CurrentPrice     float64 `json:"current_price"`
	MarketValue      float64 `json:"market_value"`
	UnrealizedPnL    float64 `json:"unrealized_pnl"`
	UnrealizedPnLPct float64 `json:"unrealized_pnl_pct"`
}

// OrderRequest is the payload for creating a new order.
type OrderRequest struct {
	Ticker    string  `json:"ticker"`
	Side      string  `json:"side"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	OrderType string  `json:"order_type"`
	StopPrice float64 `json:"stop_price,omitempty"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after successful authentication.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// BRVMTickers returns the static list of BRVM-listed securities.
func BRVMTickers() []Ticker {
	return []Ticker{
		{Symbol: "SNTS", Name: "Sonatel", Country: "Sénégal", Sector: "Télécom"},
		{Symbol: "ORAC", Name: "Orange CI", Country: "Côte d'Ivoire", Sector: "Télécom"},
		{Symbol: "SGBC", Name: "SGB CI", Country: "Côte d'Ivoire", Sector: "Finance"},
		{Symbol: "ECOC", Name: "Ecobank CI", Country: "Côte d'Ivoire", Sector: "Finance"},
		{Symbol: "PALC", Name: "PalmCI", Country: "Côte d'Ivoire", Sector: "Agriculture"},
		{Symbol: "SPHC", Name: "SAPH CI", Country: "Côte d'Ivoire", Sector: "Agriculture"},
		{Symbol: "SMB", Name: "SMB CI", Country: "Côte d'Ivoire", Sector: "Industrie"},
		{Symbol: "SOGC", Name: "SOGB CI", Country: "Côte d'Ivoire", Sector: "Industrie"},
		{Symbol: "SLBC", Name: "Solibra CI", Country: "Côte d'Ivoire", Sector: "Consommation"},
		{Symbol: "TTLC", Name: "Total CI", Country: "Côte d'Ivoire", Sector: "Distribution"},
		{Symbol: "TTLS", Name: "Total SN", Country: "Sénégal", Sector: "Distribution"},
		{Symbol: "UNLC", Name: "Unilever CI", Country: "Côte d'Ivoire", Sector: "Consommation"},
		{Symbol: "NEST", Name: "Nestlé CI", Country: "Côte d'Ivoire", Sector: "Consommation"},
		{Symbol: "NSIA", Name: "NSIA Banque", Country: "Côte d'Ivoire", Sector: "Finance"},
		{Symbol: "ETI", Name: "ETI TG", Country: "Togo", Sector: "Finance"},
		{Symbol: "ORGT", Name: "Oragroup TG", Country: "Togo", Sector: "Finance"},
		{Symbol: "ONTBF", Name: "Onatel BF", Country: "Burkina Faso", Sector: "Télécom"},
		{Symbol: "LNB", Name: "Loterie Nationale Bénin", Country: "Bénin", Sector: "Services publics"},
	}
}
