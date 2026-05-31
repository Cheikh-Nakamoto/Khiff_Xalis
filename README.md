# BRVM Trading Engine - Clean Architecture

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FLUTTER APP                                      │
│                    (iOS / Android / Web / Desktop)                            │
│         Dashboard • Portefeuille • Alertes • Ordres • Heatmap                 │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼ HTTP/gRPC
┌─────────────────────────────────────────────────────────────────────────────┐
│                           GO API GATEWAY (Fiber)                              │
│              Auth • Rate Limiting • Routing • WebSocket • REST                │
│                     /api/v1/market • /api/v1/signals                          │
│                     /api/v1/portfolio • /api/v1/orders                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
        ┌─────────────────────────────┼─────────────────────────────┐
        │                             │                             │
        ▼                             ▼                             ▼
┌───────────────┐           ┌───────────────┐           ┌───────────────┐
│   GO COLLECTOR│           │   RUST CORE   │           │   GO NOTIFIER │
│   (Scheduler) │           │   (Engine)    │           │   (Push/SMS)  │
│               │           │               │           │               │
│ • GitHub CSV  │           │ • Technical   │           │ • Firebase    │
│ • Sikafinance │           │   Analysis    │           │ • SendGrid    │
│ • FluxBourse  │           │ • Fundamental │           │ • Twilio      │
│ • BRVM Intel  │           │   Scoring     │           │ • WebSocket   │
│ • Caching     │           │ • Signal Gen  │           │ • Slack       │
│ • Validation  │           │ • Risk Mgmt   │           │               │
└───────┬───────┘           └───────┬───────┘           └───────────────┘
        │                             │
        └─────────────────────────────┼─────────────────────────────┐
                                      │                             │
                                      ▼                             ▼
                        ┌─────────────────────────┐       ┌─────────────────┐
                        │      POSTGRESQL         │       │     REDIS       │
                        │    + TimescaleDB        │       │   (Cache/Queue) │
                        │                         │       │                 │
                        │ • market_data (hypertable)    │       • Cache TTL 15min│
                        │ • fundamental_data            │       • Pub/Sub signals│
                        │ • scoring_results             │       • Session store  │
                        │ • portfolio                   │       • Rate limiter   │
                        │ • orders                      │                         │
                        │ • users                       │                         │
                        └─────────────────────────┘       └─────────────────┘
```

## Stack Technique

| Couche | Tech | Justification |
|--------|------|---------------|
| **Frontend** | Flutter (Dart) | Cross-platform natif, UI réactive, graphiques performants |
| **API Gateway** | Go + Fiber | Ultra-performant, faible latence, goroutines natives |
| **Data Collection** | Go + Colly/HTTP | Async HTTP, parsing CSV rapide, goroutines parallèles |
| **Calcul Engine** | Rust + Tokio | Safety mémoire, performance brute, calculs numériques |
| **Database** | PostgreSQL + TimescaleDB | Séries temporelles natives, SQL standard |
| **Cache** | Redis | Pub/Sub temps réel, cache rapide, sessions |
| **Queue** | Redis Streams / NATS | Messages entre Go et Rust, async processing |
| **Scheduler** | Go + robfig/cron | Tâches CRON, collecte 15min, scans quotidiens |
| **Auth** | Go + JWT + Argon2 | Stateless, sécurisé, refresh tokens |
| **FIX Protocol** | Rust + quickfix-rs | Messages FIX standard, VPN SGI, low-level safe |
| **Tests** | Go (testify) + Rust (cargo test) | TDD, benchmarks, property-based testing |
| **Container** | Docker + Docker Compose | Déploiement portable, scaling horizontal |
| **Monitoring** | Prometheus + Grafana | Métriques temps réel, alerting |
| **Logs** | Go/Zerolog + Rust/tracing | Structured logging, tracing distribué |

## Pourquoi Go + Rust + Flutter ?

### Go (API + Collecte)
- **Goroutines** : 10 000 requêtes HTTP concurrentes sans effort
- **Parsing CSV** : `encoding/csv` natif, rapide, peu de mémoire
- **HTTP** : `Fiber` ou `Gin` — 2× plus rapide que Node.js
- **Écosystème** : `robfig/cron`, `go-redis`, `pgx` (PostgreSQL driver natif)
- **Déploiement** : Binary unique, statique, cross-compilation facile

### Rust (Engine de Calcul)
- **Performance** : Calculs SMA/EMA/RSI/MACD sur 10 000 points : 10× plus rapide que Python
- **Safety** : Pas de segfault, pas de data race, ownership strict
- **Parallélisme** : `Rayon` pour le scan parallèle des 66 tickers
- **Math** : `nalgebra`, `statrs` (statistiques), `ta` (technical analysis)
- **FFI** : Peut être appelé depuis Go via CGO ou gRPC

### Flutter (Frontend)
- **Cross-platform** : iOS, Android, Web, macOS, Windows, Linux — un seul codebase
- **Graphiques** : `fl_chart`, `syncfusion_flutter_charts` pour les chandeliers
- **State management** : `Riverpod` ou `Bloc` — prévisible, testable
- **WebSocket** : Connexion temps réel au backend Go pour les signaux
- **Offline** : `Hive` ou `Isar` pour le cache local quand la connexion est faible

## Communication Go ↔ Rust

```
┌─────────┐     gRPC / HTTP     ┌─────────┐
│   Go    │ ◄─────────────────► │  Rust   │
│ API GW  │   Protobuf / JSON   │ Engine  │
└─────────┘                     └─────────┘
```

- **gRPC** : Performant, typé, streaming bidirectionnel pour les signaux temps réel
- **Protobuf** : Schéma partagé entre Go et Rust
- **Alternative** : Redis Streams (Go pousse les données, Rust consomme et pousse les résultats)

## Schéma de Données (PostgreSQL + TimescaleDB)

```sql
-- Table hypertable pour les séries temporelles
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

-- Données fondamentales
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

-- Résultats de scoring
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

-- Portefeuille utilisateur
CREATE TABLE portfolios (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    ticker VARCHAR(10),
    quantity INTEGER,
    avg_buy_price DOUBLE PRECISION,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Ordres
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    ticker VARCHAR(10),
    side VARCHAR(4), -- BUY / SELL
    quantity INTEGER,
    price DOUBLE PRECISION,
    order_type VARCHAR(20), -- LIMIT / MARKET / STOP
    status VARCHAR(20), -- PENDING / FILLED / CANCELLED
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## API Endpoints (Go + Fiber)

```
GET  /api/v1/market/tickers              → Liste des 66 tickers
GET  /api/v1/market/data/:ticker         → Données OHLCV (param: days)
GET  /api/v1/market/latest/:ticker       → Dernière cotation
GET  /api/v1/fundamental/:ticker         → Ratios fondamentaux
GET  /api/v1/signals/:ticker             → Signal actuel + score
GET  /api/v1/signals/scan                → Scan complet du marché
POST /api/v1/signals/subscribe           → WebSocket subscription
GET  /api/v1/portfolio                   → Portefeuille utilisateur
POST /api/v1/orders                      → Passer un ordre (via SGI)
GET  /api/v1/orders                      → Historique des ordres
POST /api/v1/auth/register               → Inscription
POST /api/v1/auth/login                  → Connexion (JWT)
POST /api/v1/auth/refresh                → Refresh token
```

## Déploiement Local (Docker Compose)

```bash
# 1. Clone
git clone https://github.com/yourname/brvm-trading-engine.git
cd brvm-trading-engine

# 2. Lancer l'infrastructure
docker-compose up -d postgres redis

# 3. Build Rust engine
cd rust-engine && cargo build --release && cd ..

# 4. Build Go API
cd go-api && go build -o bin/api cmd/api/main.go && cd ..

# 5. Lancer
docker-compose up -d

# 6. Accès
# API: http://localhost:8080
# Flutter Web: http://localhost:3000
# Grafana: http://localhost:3001
# Redis Commander: http://localhost:8081
```

## Horaires de Collecte (Scheduler Go)

| Tâche | Fréquence | Description |
|-------|-----------|-------------|
| `collect_github` | Toutes les 15 min (9h-15h GMT, L-V) | CSV GitHub brvm-data-public |
| `collect_sikafinance` | Quotidien à 18h | Téléchargement CSV historique |
| `collect_fluxbourse` | Toutes les 30 min (9h-15h GMT) | Scraping PER + rendement |
| `run_scoring` | Toutes les 15 min | Calcul indicateurs + signaux |
| `notify_signals` | Toutes les 15 min | Push aux utilisateurs abonnés |
| `cleanup_cache` | Quotidien à 2h | Purge Redis TTL expirés |
| `backup_db` | Quotidien à 3h | Dump PostgreSQL → S3/MinIO |

## Sécurité

- **JWT** : Access token (15 min) + Refresh token (7 jours)
- **Rate Limiting** : 100 req/min par IP, 1000 req/min par user
- **CORS** : Origines contrôlées (Flutter app uniquement)
- **Input Validation** : `go-playground/validator` + `validator` Rust
- **SQL Injection** : Requêtes paramétrées (pgx), jamais de string interpolation
- **FIX** : VPN IPSec entre SGI et BRVM, messages FIX chiffrés

## Tests

```bash
# Go tests
cd go-api && go test ./... -race -coverprofile=coverage.out

# Rust tests
cd rust-engine && cargo test && cargo bench

# Flutter tests
cd flutter-app && flutter test

# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## Roadmap

- [ ] MVP : Collecte GitHub + Scoring basique + Flutter dashboard
- [ ] V1 : Intégration Sikafinance + FluxBourse + Notifications push
- [ ] V2 : Portefeuille virtuel + Historique P&L + Backtesting
- [ ] V3 : Intégration SGI (FIX) pour ordres réels
- [ ] V4 : Machine Learning (Rust + linfa) pour prédiction de tendance
- [ ] V5 : Social trading (copier les meilleurs portefeuilles)

## Licence

Dual licensing :

- **AGPL-3.0** — open source (voir [LICENSE](LICENSE))
- **Commercial License** — pour usage proprietary / SaaS sans obligation AGPL (voir [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md))

Copyright 2026 BRVM Trading Engine Contributors
