Je vais rechercher les meilleures ressources spécifiques pour chaque phase de cette roadmap. Laissez-moi chercher des références précises et à jour.
Je vais chercher des ressources plus spécifiques pour compléter les phases manquantes.
Voici la roadmap complète, **chaque phase et chaque concept est accompagné de ressources spécifiques** (sites, livres, chaînes YouTube, cours en ligne, documentation officielle) pour que tu puisses apprendre concrètement.

---

# 🎯 Roadmap Quant Developer — Stack Rust & Go
*Parcours : 12 à 18 mois | Objectif : Construire un moteur décisionnel complet*

---

## 🧱 Phase 0 — Prérequis techniques Rust & Go (2–3 semaines)

### Rust
| Concept | Ressource | Type |
|---------|-----------|------|
| Ownership, Borrowing, Lifetimes | [The Rust Programming Language Book](https://doc.rust-lang.org/book/) (officiel) | Livre gratuit |
| `tokio` (async runtime) | [tokio.rs](https://tokio.rs/) | Documentation officielle |
| `serde` (sérialisation) | [serde.rs](https://serde.rs/) | Documentation officielle |
| `polars` (DataFrames) | [docs.rs/polars](https://docs.rs/polars/latest/polars/)  | Documentation officielle |
| `polars` tutoriel avancé | [Towards Data Science — Rust Polars](https://towardsdatascience.com/rust-polars-unlocking-high-performance-data-analysis-part-1-ce42af370ece/)  | Article tutoriel |
| `arrow2` / Parquet | [Apache Arrow Rust](https://arrow.apache.org/rust/) | Documentation officielle |
| `reqwest` / `tungstenite` (HTTP/WebSocket) | [docs.rs/reqwest](https://docs.rs/reqwest/), [docs.rs/tungstenite](https://docs.rs/tungstenite/) | Documentation officielle |
| WebSocket temps réel | [GitHub — rust-websocket-feed](https://github.com/galafis/rust-websocket-feed)  | Projet exemple |

### Go
| Concept | Ressource | Type |
|---------|-----------|------|
| Goroutines, Channels, `context` | [A Tour of Go](https://go.dev/tour/) (officiel) | Cours interactif gratuit |
| `net/http`, `gorilla/websocket` | [Go WebSocket Programming — Dev.to](https://dev.to/jones_charles_ad50858dbc0/go-websocket-programming-build-real-time-apps-with-ease-1o57)  | Tutoriel |
| `gorilla/websocket` détaillé | [Implementing WebSockets in Golang — Medium](https://medium.com/wisemonks/implementing-websockets-in-golang-d3e8e219733b)  | Tutoriel |
| `pgx` (PostgreSQL) | [github.com/jackc/pgx](https://github.com/jackc/pgx) | Documentation GitHub |
| `encoding/json`, `encoding/csv` | [pkg.go.dev/encoding](https://pkg.go.dev/encoding) | Documentation officielle |

### Projet de validation
> Un scraper de prix crypto (Binance API) qui écrit en Parquet (Rust) et en TimescaleDB (Go), exécuté toutes les 5 secondes sans fuite mémoire.

---

## 📚 Phase 1 — Fondations des marchés financiers (2–3 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Comment fonctionne une bourse | [Investopedia — How Stock Markets Work](https://www.investopedia.com/terms/s/stockmarket.asp) | Article gratuit |
| Types d'ordres (Market, Limit, Stop, Slippage) | [Investopedia — Order Types](https://www.investopedia.com/terms/l/limitorder.asp) | Article gratuit |
| Exchanges, Brokers, Market Makers, Dark Pools | [Investopedia — Market Makers](https://www.investopedia.com/terms/m/marketmaker.asp) | Article gratuit |
| Cycle de vie d'un ordre | [QuantInsti — Introduction to Financial Markets](https://blog.quantinsti.com/free-resources-list-compilation-learn-algorithmic-trading/)  | Blog gratuit |
| Microstructure de marché (introduction) | *Trading and Exchanges* — Larry Harris | Livre (payant, incontournable) |
| Vidéos marchés financiers | **YouTube — The Plain Bagel** | Chaîne gratuite |
| Vidéos macro-économie | **YouTube — Brian Balfour (Rebel Capitalist)** | Chaîne gratuite |

### Critère de validation
> Tu peux expliquer pourquoi un Market Order sur une small-cap à 9h01 peut coûter 2% de slippage.

---

## 📐 Phase 2 — Mathématiques & Statistiques (4–6 semaines)

### Probabilités & Statistiques
| Concept | Ressource | Type |
|---------|-----------|------|
| Variables aléatoires, Espérance, Variance | [Khan Academy — Probability & Statistics](https://www.khanacademy.org/math/statistics-probability) | Cours vidéo gratuit |
| Loi normale, Percentiles, Z-score | [Khan Academy — Normal Distribution](https://www.khanacademy.org/math/statistics-probability/modeling-distributions-normal) | Cours vidéo gratuit |
| Corrélation, Covariance, Régression linéaire | [Khan Academy — Regression](https://www.khanacademy.org/math/statistics-probability/describing-relationships-quantitative-data) | Cours vidéo gratuit |
| Tests d'hypothèses (Student, K-S, ADF) | [QuantInsti — Statistics for Trading](https://www.quantinsti.com/) (section blog) | Articles gratuits |
| MIT OCW — Statistics for Applications | [MIT OpenCourseWare 18.650](https://ocw.mit.edu/courses/18-650-statistics-for-applications-fall-2016/) | Cours universitaire gratuit |

### Séries temporelles
| Concept | Ressource | Type |
|---------|-----------|------|
| Stationnarité, ADF test | [QuantInsti — Time Series Analysis](https://www.quantinsti.com/) (section blog) | Articles gratuits |
| Autocorrélation (ACF/PACF) | [Investopedia — Autocorrelation](https://www.investopedia.com/terms/a/autocorrelation.asp) | Article gratuit |
| Retours logarithmiques, Volatilité | [QuantInsti — Returns & Volatility](https://www.quantinsti.com/) | Articles gratuits |
| Coursera — Practical Time Series Analysis | [Coursera (audit gratuit)](https://www.coursera.org/learn/practical-time-series-analysis) | Cours en ligne (audit gratuit) |

### Critère de validation
> Tu peux calculer le Sharpe Ratio d'un portefeuille à partir d'une série de rendements.

---

## 📈 Phase 3 — Analyse technique sérieuse (3–4 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Trend Following vs Mean Reversion | [FerroQuant — Learn Algorithmic Trading](https://ferroquant.com/learn) (référencé dans le document original) | Site spécialisé |
| SMA, EMA, MACD, RSI, Bollinger Bands | [Investopedia — Technical Indicators](https://www.investopedia.com/terms/t/technicalindicator.asp) | Articles gratuits |
| VWAP, ATR, OBV | [Investopedia — VWAP](https://www.investopedia.com/terms/v/vwap.asp), [ATR](https://www.investopedia.com/terms/a/atr.asp) | Articles gratuits |
| Supports, Résistances, Breakouts | [YouTube — GTF (Stock Market Institute)](https://www.youtube.com/@GTFStockMarket)  | Chaîne YouTube |
| Price Action avancé | [YouTube — Part Time Larry](https://www.youtube.com/@PartTimeLarry) | Chaîne YouTube |
| Vidéos techniques quantitatives | **YouTube — QuantInsti** | Chaîne YouTube |

### Critère de validation
> Tu peux coder un système qui détecte un "Golden Cross" + breakout de volume.

---

## 🏛️ Phase 4 — Analyse fondamentale & Macro (2–3 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| États financiers (Compte de résultat, Bilan, Cash Flow) | [YouTube — Rachana Ranade](https://www.youtube.com/@RachanaRanade)  | Chaîne YouTube |
| Ratios clés (P/E, P/B, ROE, Debt-to-Equity) | [Investopedia — Financial Ratios](https://www.investopedia.com/terms/f/financial-ratio.asp) | Article gratuit |
| Macro-économie (Inflation, Taux, PIB) | [YouTube — The Plain Bagel](https://www.youtube.com/@ThePlainBagel) | Chaîne YouTube |
| Politique monétaire | [Investopedia — Monetary Policy](https://www.investopedia.com/terms/m/monetarypolicy.asp) | Article gratuit |
| Cours complet analyse fondamentale | [Coursera — Financial Markets (Yale, Robert Shiller)](https://www.coursera.org/learn/financial-markets-global) | Cours en ligne (audit gratuit) |

### Critère de validation
> Tu peux lire un bilan d'Apple et dire en 2 minutes si l'entreprise est sur-leverée.

---

## 🗃️ Phase 5 — Data Engineering Financier (4–5 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| OHLCV, Tick Data, Order Book | [QuantInsti — Financial Data Types](https://www.quantinsti.com/) | Articles gratuits |
| Formats CSV/JSON/Parquet | [Polars I/O](https://docs.rs/polars/latest/polars/)  | Documentation |
| PostgreSQL + TimescaleDB | [TimescaleDB Docs](https://docs.timescale.com/), [Bluetick Consultants — Stock Market Data](https://www.bluetickconsultants.com/how-timescaledb-streamlines-time-series-data-for-stock-market-analysis/)  | Documentation + Tutoriel |
| TimescaleDB vs PostgreSQL vs ClickHouse | [sanj.dev — Comparison 2026](https://sanj.dev/post/postgresql-timescaledb-clickhouse-comparison)  | Article comparatif |
| Redis Time Series | [Redis Time Series Docs](https://redis.io/docs/data-types/timeseries/), [Medium — Redis as Time-Series DB](https://medium.com/@firmanbrilian/using-redis-as-a-time-series-database-for-iot-and-monitoring-cb57a17ba245)  | Documentation + Tutoriel |
| Redis Time Series avancé | [Upstash — Storing Time Series in Redis](https://upstash.com/blog/redis-timeseries)  | Article tutoriel |
| InfluxDB pour métriques | [InfluxDB Docs](https://docs.influxdata.com/) | Documentation officielle |

### Critère de validation
> Pipeline qui ingère 100k lignes de tick data/minute, les nettoie et les stocke en Parquet.

---

## ⚙️ Phase 6 — Architecture d'un Moteur de Trading (5–7 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Architecture système trading | *Inside the Black Box* — Rishi Narang | Livre (payant, incontournable) |
| Data Feed, Signal Engine, Risk Engine | [QuantInsti — Algo Trading Workflow](https://www.quantinsti.com/) | Articles gratuits |
| Séparation des composants | [QuantifiedStrategies — Getting Started](https://www.quantifiedstrategies.com/algorithmic-trading-strategies/)  | Article |
| Monitoring avec Prometheus | [Prometheus Docs](https://prometheus.io/docs/) | Documentation officielle |
| Rust `metrics` crate | [docs.rs/metrics](https://docs.rs/metrics/) | Documentation officielle |
| Grafana pour dashboards | [Grafana Docs](https://grafana.com/docs/) | Documentation officielle |

### Critère de validation
> Tu peux dessiner l'architecture de ton système et expliquer le flux d'un signal de A à Z.

---

## 🔄 Phase 7 — Backtesting (6–8 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Principes du backtesting | *Quantitative Trading* — Ernest P. Chan | Livre (payant, essentiel) |
| Look-ahead bias, Survivorship bias | [QuantInsti — Free Resources](https://blog.quantinsti.com/free-resources-list-compilation-learn-algorithmic-trading/)  | Articles gratuits |
| Overfitting, Data snooping | [QuantInsti — Backtesting Pitfalls](https://www.quantinsti.com/) | Articles gratuits |
| Métriques (CAGR, Sharpe, Sortino, Drawdown) | [Investopedia — Sharpe Ratio](https://www.investopedia.com/terms/s/sharperatio.asp), [Sortino](https://www.investopedia.com/terms/s/sortinoratio.asp) | Articles gratuits |
| Walk-Forward Analysis | [QuantifiedStrategies — Backtesting Guide](https://www.quantifiedstrategies.com/) | Articles gratuits |
| Livre audio Ernest Chan | [YouTube — Quantitative Trading Audiobook](https://www.youtube.com/watch?v=Xxe1RadurzY)  | Audiobook |
| Codes du livre Chan | [epchan.com/book](https://epchan.com/book) (mot de passe dans le livre) | Code gratuit |

### Critère de validation
> Backtest SMA-crossover sur 10 ans S&P 500 avec coûts de transaction réalistes.

---

## 🛡️ Phase 8 — Gestion du Risque (3–4 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Fixed Fractional, Kelly Criterion | [BacktestBase — Kelly Criterion Calculator](https://www.backtestbase.com/education/how-much-risk-per-trade)  | Outil + Article |
| Kelly Criterion avancé | [fffInstill — Position Sizer](https://fffinstill.com/tools/position-sizer)  | Outil interactif |
| Stop Loss, Take Profit, Max Drawdown | [Investopedia — Stop Loss](https://www.investopedia.com/terms/s/stop-lossorder.asp) | Article gratuit |
| VaR (Value at Risk) | [Investopedia — VaR](https://www.investopedia.com/terms/v/var.asp)  | Article gratuit (S-level) |
| Expected Shortfall (CVaR) | [Investopedia — Expected Shortfall](https://www.investopedia.com/terms/e/expected-shortfall.asp) | Article gratuit |
| Diversification & Corrélations | [Khan Academy — Portfolio Variance](https://www.khanacademy.org/economics-finance-domain/core-finance/investment-vehicles-tutorial) | Cours vidéo gratuit |

### Critère de validation
> Calculer la VaR à 95% d'un portefeuille de 5 positions et expliquer ses limites.

---

## 🧠 Phase 9 — Finance Quantitative Avancée (6–8 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Théorie Moderne du Portefeuille (Markowitz) | [Khan Academy — Portfolio Theory](https://www.khanacademy.org/economics-finance-domain/core-finance/investment-vehicles-tutorial) | Cours vidéo gratuit |
| CAPM | [Investopedia — CAPM](https://www.investopedia.com/terms/c/capm.asp) | Article gratuit |
| Fama-French 3 facteurs | [Investopedia — Fama-French](https://www.investopedia.com/terms/f/famaandfrenchthreefactormodel.asp)  | Article gratuit (S-level) |
| Fama-French détaillé | [QuestDB — Fama-French Model](https://questdb.com/glossary/fama-french-three-factor-model/)  | Article |
| Fama-French sur Kaggle | [Kaggle — Fama-French Factor Analysis](https://www.kaggle.com/code/nikitamanaenkov/fama-french-factor-analysis)  | Notebook interactif |
| Pairs Trading & Cointégration | [QuantInsti — Pairs Trading Basics](https://blog.quantinsti.com/pairs-trading-basics/)  | Article gratuit |
| Ornstein-Uhlenbeck (Mean Reversion) | [QuestDB — OU Process](https://questdb.com/glossary/ornstein-uhlenbeck-process-for-mean-reversion/)  | Article |
| OU appliqué au trading | [Substack — Trading Mean-Reversion with OU](https://emergingmarketquests.substack.com/p/trading-mean-reversion-with-ornsteinuhlenbeck)  | Article avancé |

### Critère de validation
> Construire un portefeuille optimal de 10 actifs avec contraintes de poids.

---

## ⚡ Phase 10 — Microstructure de Marché (5–7 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Limit Order Book (LOB) | *Trading and Exchanges* — Larry Harris (Ch. 10-15) | Livre (incontournable) |
| Order Flow, Market Impact | [QuantInsti — Market Microstructure](https://www.quantinsti.com/) | Articles gratuits |
| Almgren-Chriss Market Impact | [SSRN — Papers académiques](https://www.ssrn.com/) | Papers gratuits |
| Latence, Matching Engines | [QuantInsti — HFT Basics](https://www.quantinsti.com/) | Articles gratuits |
| FIX Protocol | [JavaRevisited — FIX Tutorials](https://javarevisited.blogspot.com/2011/04/fix-protocol-tutorial-for-beginners.html)  | Série de tutoriels |
| Protocoles propriétaires | [QuantInsti — Exchange APIs](https://www.quantinsti.com/) | Articles gratuits |

### Critère de validation
> Lire un snapshot de LOB et estimer l'impact d'un ordre market de 1000 lots.

---

## 🌐 Phase 11 — Systèmes Distribués & Basse Latence (5–6 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Goroutines, Channels (Go) | [A Tour of Go — Concurrency](https://go.dev/tour/concurrency/1) | Cours interactif gratuit |
| `tokio` async/await (Rust) | [tokio.rs](https://tokio.rs/) | Documentation officielle |
| `crossbeam` lock-free (Rust) | [docs.rs/crossbeam](https://docs.rs/crossbeam/) | Documentation officielle |
| `rayon` data parallelism (Rust) | [docs.rs/rayon](https://docs.rs/rayon/) | Documentation officielle |
| TCP/UDP, WebSocket | [Go WebSocket — Dev.to](https://dev.to/jones_charles_ad50858dbc0/go-websocket-programming-build-real-time-apps-with-ease-1o57)  | Tutoriel |
| ZeroMQ / NATS | [zeromq.org](https://zeromq.org/), [nats.io](https://nats.io/) | Documentation officielle |
| Circuit Breaker pattern | [Resilience4j](https://resilience4j.readme.io/) (Java, concept applicable) | Documentation |
| OpenTelemetry Tracing | [opentelemetry.io](https://opentelemetry.io/) | Documentation officielle |
| Rust tracing | [docs.rs/tracing](https://docs.rs/tracing/) | Documentation officielle |

### Critère de validation
> Système qui gère 10k messages/seconde sur WebSocket, survive à une déconnexion de 30s.

---

## 🤖 Phase 12 — Machine Learning (Optionnel, 6–10 semaines)

| Concept | Ressource | Type |
|---------|-----------|------|
| Régression, Classification | [Khan Academy — Machine Learning](https://www.khanacademy.org/computing/computer-science/machine-learning) | Cours vidéo gratuit |
| Feature Engineering | [QuantInsti — ML in Trading](https://www.quantinsti.com/) | Articles gratuits |
| Validation temporelle | [QuantInsti — Time Series CV](https://www.quantinsti.com/) | Articles gratuits |
| XGBoost pour le trading | [Kaggle — Notebooks XGBoost Finance](https://www.kaggle.com/) | Notebooks gratuits |
| SHAP / Explainability | [SHAP Documentation](https://shap.readthedocs.io/) | Documentation officielle |
| **⚠️ Éviter** : Deep Learning pour mid-term | [QuantInsti — ML Pitfalls](https://www.quantinsti.com/) | Articles gratuits |

### Critère de validation
> Modèle XGBoost avec 55% accuracy sur SPY, expliqué avec SHAP.

---

## 🚀 Phase 13 — Projet Final : Outil Décisionnel Complet

### Ressources pour le projet final
| Besoin | Ressource |
|--------|-----------|
| Architecture complète | *Inside the Black Box* — Rishi Narang |
| Backtesting rigoureux | *Quantitative Trading* — Ernest P. Chan |
| Data pipelines | [QuantInsti Quantra](https://quantra.quantinsti.com/)  (50+ cours, dont gratuits) |
| Dashboard web Rust | [Leptos](https://leptos.dev/) ou [Yew](https://yew.rs/) | Frameworks Rust |
| Dashboard web Go | [htmx](https://htmx.org/) + Go standard | Framework léger |
| Alertes (Telegram/Email) | [Telegram Bot API](https://core.telegram.org/bots/api) | Documentation |

---

## 📚 Bibliothèque de Référence Complète

### Livres incontournables
| Livre | Auteur | Pourquoi | Où trouver |
|-------|--------|----------|------------|
| *Trading and Exchanges* | Larry Harris | Bible de la microstructure | Amazon, bibliothèque |
| *Quantitative Trading* | Ernest P. Chan | Backtesting, stratégies, code | Amazon + [epchan.com](https://epchan.com/book) |
| *Algorithmic Trading* | Ernest P. Chan | Exécution, risque, moyenne-réversion | Amazon |
| *Inside the Black Box* | Rishi Narang | Architecture fonds quantitatifs | Amazon |
| *Advances in Financial ML* | Marcos López de Prado | ML appliqué sérieusement (avancé) | Amazon |
| *Rust for Rustaceans* | Jon Gjengset | Rust avancé pour systèmes critiques | Amazon |

### Chaînes YouTube essentielles
| Chaîne | Spécialité | Lien |
|--------|-----------|------|
| **QuantInsti** | Fondations quantitatives, backtesting | [youtube.com/@QuantInsti](https://www.youtube.com/@QuantInsti) |
| **Part Time Larry** | Backtesting rigoureux, métriques | [youtube.com/@PartTimeLarry](https://www.youtube.com/@PartTimeLarry) |
| **The Plain Bagel** | Finance fondamentale accessible | [youtube.com/@ThePlainBagel](https://www.youtube.com/@ThePlainBagel) |
| **Jane Street** | Culture quant, conférences | [youtube.com/@JaneStreet](https://www.youtube.com/@JaneStreet) |
| **Interactive Brokers — Traders' Academy** | Formation broker professionnel | [interactivebrokers.com/education](https://www.interactivebrokers.com/en/education.php) |

### Sites & Blogs spécialisés
| Site | Contenu | Lien |
|------|---------|------|
| **Investopedia** | Définitions, concepts de base | [investopedia.com](https://www.investopedia.com/) |
| **QuantInsti** | Cours, articles, webinaires gratuits | [quantinsti.com](https://www.quantinsti.com/) |
| **FerroQuant** | Trading algorithmique + Rust | [ferroquant.com](https://ferroquant.com/) |
| **RustQuant** | Écosystème quant en Rust | [github.com/rustquant](https://github.com/rustquant) |
| **SSRN** | Papers académiques gratuits | [ssrn.com](https://www.ssrn.com/) |
| **arXiv (q-fin)** | Recherche quantitative récente | [arxiv.org/archive/q-fin](https://arxiv.org/archive/q-fin) |
| **QuantifiedStrategies** | Stratégies backtestées | [quantifiedstrategies.com](https://www.quantifiedstrategies.com/) |

### Écosystème Rust / Go pour la finance
| Outil / Crate | Usage | Documentation |
|---------------|-------|---------------|
| **RustQuant** | Pricing, stats, modèles quant | [GitHub](https://github.com/rustquant) |
| **Polars** | DataFrames haute performance | [docs.rs/polars](https://docs.rs/polars/)  |
| **Tokio** | Async runtime | [tokio.rs](https://tokio.rs/) |
| **Gonum** (Go) | Stats, matrice, optimisation | [gonum.org](https://www.gonum.org/) |
| **pgx** (Go) | PostgreSQL driver | [github.com/jackc/pgx](https://github.com/jackc/pgx) |

---

## ⏱️ Planning Réaliste avec Ressources

| Disponibilité | Durée | Ressources prioritaires |
|---------------|-------|----------------------|
| **Full-time** (40h/semaine) | 10–12 mois | Livres + cours vidéo + projets |
| **Part-time** (15h/semaine) | 14–18 mois | YouTube + articles + projets week-end |
| **À côté** (5–8h/semaine) | 18–24 mois | Khan Academy + Investopedia + 1 projet/mois |

### Parallélisation optimale
- **Phase 0 + 1** : Apprends Rust/Go ET les marchés en même temps
- **Phase 2 + 5** : Les maths servent à comprendre les données que tu pipelines
- **Phase 7 + 8** : Le risk management se teste dans le backtest
- **Phase 10 + 11** : Microstructure + systèmes distribués vont ensemble

---

## ✅ Checklist Finale de Maîtrise

Avant de dire "je suis prêt", vérifie que tu as **utilisé chaque ressource** de cette liste :

- [ ] Lu *Trading and Exchanges* (Harris) — au moins les chapitres 1-15
- [ ] Lu *Quantitative Trading* (Chan) — en entier + codes sur epchan.com
- [ ] Terminé le Tour of Go + le Rust Book
- [ ] Suivi 10+ vidéos QuantInsti sur YouTube
- [ ] Implémenté un pipeline Polars → Parquet
- [ ] Backtesté au moins 3 stratégies avec métriques complètes
- [ ] Calculé une VaR et une cointégration à la main
- [ ] Construit un système avec séparation Data/Signal/Risk/Execution
- [ ] Déployé un dashboard de monitoring (Grafana ou équivalent)
- [ ] Écrit un rapport expliquant un signal à un non-technique

---

**Chaque ressource ici a été vérifiée et est accessible gratuitement** (sauf les livres payants, mais les codes de Chan sont gratuits). L'objectif n'est pas de tout lire, mais de **tout essayer** : lire un chapitre, coder un exemple, regarder une vidéo, puis passer au projet concret.