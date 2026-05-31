# BRVM Trading Engine

## Project Overview
Multi-service trading platform for the Bourse Régionale des Valeurs Mobilières (BRVM).
42 entreprises cotées, 8 pays UEMOA, données temps réel.

## Architecture
```
Flutter App → Go API Gateway (Fiber) → Rust Core Engine (gRPC)
                                   → PostgreSQL + TimescaleDB
                                   → Redis (cache/pub-sub)
Go Collector → GitHub CSV / Sikafinance / FluxBourse → DB
Go Notifier ← Redis Pub/Sub → FCM / Twilio / SendGrid / Slack
```

## Stack
| Layer | Technology |
|---|---|
| Frontend | Flutter (Dart) + Riverpod + fl_chart |
| API Gateway | Go + Fiber v2 + WebSocket |
| Data Collection | Go + cron scheduler |
| Calculation Engine | Rust + Tokio + rayon |
| Database | PostgreSQL + TimescaleDB (hypertables) |
| Cache/Queue | Redis (Streams / Pub/Sub) |
| Auth | JWT + Argon2 |
| Containerization | Docker + Docker Compose |
| Monitoring | Prometheus + Grafana |

## Services
- `go-api/` — REST + WebSocket API (port 8080)
- `rust-engine/` — Technical analysis + scoring (port 50051 gRPC)
- `go-collector/` — Data collection scheduler (GitHub CSV, Sikafinance, FluxBourse)
- `go-notifier/` — Push/SMS/Email notifications
- `flutter-app/` — Cross-platform UI

## API Endpoints
- `/api/v1/market/{ticker}` — OHLCV data
- `/api/v1/fundamental/{ticker}` — Ratios fondamentaux
- `/api/v1/signals` — Signaux d'achat/vente
- `/api/v1/portfolio` — Gestion portefeuille
- `/api/v1/orders` — Passage d'ordres
- `/api/v1/auth/*` — JWT auth (register, login, refresh)

## Database Schema
- `market_data` — TimescaleDB hypertable (OHLCV)
- `fundamental_data` — PER, ROE, dividend yield, EPS
- `scoring_results` — Composite scores + signals
- `portfolios` — User portfolios
- `orders` — Buy/sell orders
- `users` — User accounts
- `refresh_tokens` — JWT refresh tokens

## Collection Schedule
- GitHub CSV: every 15 min (9h-15h GMT, Mon-Fri)
- Sikafinance: daily at 18h
- FluxBourse: every 30 min (9h-15h GMT)
- Scoring: every 15 min
- Cache cleanup: daily at 2h
- DB backup: daily at 3h

## Security
- JWT: 15min access / 7j refresh
- Rate limiting: 100 req/min per IP, 1000 per user
- CORS enabled
- Parameterized SQL queries (pgx)
- FIX over IPSec VPN (SGI integration, V3)

## Agents
- **aziz** — CTO (arbitrage, vision stratégique)
- **lucas** — Tech Lead / Architecte
- **mouha** — Flutter Developer
- **mounzil** — Backend Developer (Go API)
- **malang** — DevOps / SRE
- **massar** — Security Engineer
- **janel** — QA / Test Engineer
- **product-manager** — Product Manager
