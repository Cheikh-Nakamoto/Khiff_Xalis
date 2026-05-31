-- Copyright 2026 BRVM Trading Engine Contributors
-- Licensed under the GNU AGPL-3.0. See LICENSE for details.
--
-- BRVM Trading Engine - Database Schema
-- PostgreSQL + TimescaleDB

-- Extensions
CREATE EXTENSION IF NOT EXISTS timescaledb;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Market data (hypertable)
CREATE TABLE market_data (
    time TIMESTAMPTZ NOT NULL,
    ticker VARCHAR(10) NOT NULL,
    open DOUBLE PRECISION,
    high DOUBLE PRECISION,
    low DOUBLE PRECISION,
    close DOUBLE PRECISION,
    volume BIGINT,
    source VARCHAR(50)
);

SELECT create_hypertable('market_data', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_market_ticker_time ON market_data (ticker, time DESC);

-- Fundamental data
CREATE TABLE fundamental_data (
    ticker VARCHAR(10) PRIMARY KEY,
    per DOUBLE PRECISION,
    roe DOUBLE PRECISION,
    dividend_yield DOUBLE PRECISION,
    eps DOUBLE PRECISION,
    book_value_per_share DOUBLE PRECISION,
    debt_to_equity DOUBLE PRECISION,
    revenue_growth DOUBLE PRECISION,
    net_profit DOUBLE PRECISION,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Scoring results
CREATE TABLE scoring_results (
    id SERIAL PRIMARY KEY,
    ticker VARCHAR(10) NOT NULL,
    composite_score DOUBLE PRECISION,
    signal_type VARCHAR(20),
    confidence DOUBLE PRECISION,
    reasons TEXT[],
    technical_json JSONB,
    fundamental_json JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_scoring_ticker_time ON scoring_results (ticker, created_at DESC);

-- Portfolios
CREATE TABLE portfolios (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    ticker VARCHAR(10),
    quantity INTEGER,
    avg_buy_price DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Orders
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    ticker VARCHAR(10),
    side VARCHAR(4),
    quantity INTEGER,
    price DOUBLE PRECISION,
    order_type VARCHAR(20),
    status VARCHAR(20) DEFAULT 'PENDING',
    stop_price DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(500) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_refresh_token ON refresh_tokens (token);
