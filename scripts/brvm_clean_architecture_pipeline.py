
"""
================================================================================
BRVM DATA PIPELINE & TRADING SIGNALS ENGINE
Clean Architecture - Python Implementation
================================================================================

SOURCES DE DONNÉES BRVM IDENTIFIÉES:
-------------------------------------
1. brvm-data-public (GitHub)     - CSV auto-refresh toutes les 15 min
2. sikafinance.com               - Données historiques téléchargeables
3. fluxbourse.com                - Cours temps réel + PER + rendement
4. brvm.org (flux différé)       - Données officielles BRVM
5. ICE Data API                  - Données institutionnelles Level 1/2
6. richbourse.com                - Scraping via package R BRVM
7. BRVM Intelligence             - API IA + screener + signaux

PROTOCOLE FIX (pour SGI agréées):
---------------------------------
- BRVM FIX Gateway pour routage automatique des ordres
- VPN IPSec entre SGI et BRVM
- Messages FIX standard pour ordres, modifications, annulations
"""

# =============================================================================
# LAYER 1: DOMAIN (Entités métier pures - aucune dépendance externe)
# =============================================================================

from dataclasses import dataclass, field
from datetime import datetime, date
from typing import Optional, List, Dict, Callable
from enum import Enum
import json

class SignalType(Enum):
    STRONG_BUY = "STRONG_BUY"
    BUY = "BUY"
    HOLD = "HOLD"
    SELL = "SELL"
    STRONG_SELL = "STRONG_SELL"

class DataSource(Enum):
    GITHUB_BRVM = "github_brvm_data_public"
    SIKAFINANCE = "sikafinance"
    FLUXBOURSE = "fluxbourse"
    BRVM_OFFICIAL = "brvm_org"
    ICE_API = "ice_data"
    RICHBOURSE = "richbourse"
    BRVM_INTELLIGENCE = "brvm_intelligence"

@dataclass(frozen=True)
class Ticker:
    symbol: str
    name: str
    country: str
    sector: str

@dataclass
class MarketData:
    ticker: Ticker
    date: date
    open_price: float
    high: float
    low: float
    close: float
    volume: int
    source: DataSource

    @property
    def daily_return(self) -> float:
        return (self.close - self.open_price) / self.open_price * 100

    @property
    def volatility(self) -> float:
        return (self.high - self.low) / self.open_price * 100

@dataclass
class FundamentalData:
    ticker: Ticker
    per: Optional[float] = None
    roe: Optional[float] = None
    dividend_yield: Optional[float] = None
    eps: Optional[float] = None
    book_value_per_share: Optional[float] = None
    debt_to_equity: Optional[float] = None
    revenue_growth: Optional[float] = None
    net_profit: Optional[float] = None

    @property
    def pb_ratio(self) -> Optional[float]:
        if self.book_value_per_share and self.book_value_per_share > 0:
            # close price would come from market data
            return None  # computed at scoring time
        return None

@dataclass
class TechnicalIndicators:
    sma_20: Optional[float] = None
    sma_50: Optional[float] = None
    ema_12: Optional[float] = None
    ema_26: Optional[float] = None
    rsi_14: Optional[float] = None
    macd: Optional[float] = None
    macd_signal: Optional[float] = None
    bollinger_upper: Optional[float] = None
    bollinger_lower: Optional[float] = None
    atr_14: Optional[float] = None  # Average True Range
    volume_sma_20: Optional[float] = None

@dataclass
class ScoringResult:
    ticker: Ticker
    score_value: float  # 0-100
    score_growth: float
    score_dividend: float
    score_momentum: float
    score_risk: float
    composite_score: float
    signal: SignalType
    confidence: float  # 0-1
    reasons: List[str] = field(default_factory=list)
    timestamp: datetime = field(default_factory=datetime.now)

# =============================================================================
# LAYER 2: USE CASES (Services métier - dépendent uniquement du Domain)
# =============================================================================

class TechnicalAnalysisService:
    """Calcule les indicateurs techniques à partir d'une série de prix."""

    def calculate_indicators(self, prices: List[float], volumes: List[float]) -> TechnicalIndicators:
        if len(prices) < 50:
            return TechnicalIndicators()

        indicators = TechnicalIndicators()

        # SMA
        indicators.sma_20 = sum(prices[-20:]) / 20
        indicators.sma_50 = sum(prices[-50:]) / 50

        # EMA
        indicators.ema_12 = self._ema(prices, 12)
        indicators.ema_26 = self._ema(prices, 26)

        # MACD
        if indicators.ema_12 and indicators.ema_26:
            indicators.macd = indicators.ema_12 - indicators.ema_26
            indicators.macd_signal = self._ema([indicators.macd] + [0]*8, 9)  # simplified

        # RSI
        indicators.rsi_14 = self._rsi(prices, 14)

        # Bollinger Bands
        if len(prices) >= 20:
            sma20 = indicators.sma_20
            std20 = (sum((p - sma20)**2 for p in prices[-20:]) / 20) ** 0.5
            indicators.bollinger_upper = sma20 + 2 * std20
            indicators.bollinger_lower = sma20 - 2 * std20

        # ATR
        indicators.atr_14 = self._atr(prices, 14)

        # Volume SMA
        if len(volumes) >= 20:
            indicators.volume_sma_20 = sum(volumes[-20:]) / 20

        return indicators

    def _ema(self, prices: List[float], period: int) -> float:
        if len(prices) < period:
            return prices[-1] if prices else 0
        k = 2 / (period + 1)
        ema = sum(prices[:period]) / period
        for price in prices[period:]:
            ema = price * k + ema * (1 - k)
        return ema

    def _rsi(self, prices: List[float], period: int = 14) -> float:
        if len(prices) < period + 1:
            return 50.0
        gains, losses = [], []
        for i in range(1, period + 1):
            diff = prices[-i] - prices[-i-1]
            gains.append(max(diff, 0))
            losses.append(abs(min(diff, 0)))
        avg_gain = sum(gains) / period
        avg_loss = sum(losses) / period
        if avg_loss == 0:
            return 100.0
        rs = avg_gain / avg_loss
        return 100 - (100 / (1 + rs))

    def _atr(self, prices: List[float], period: int = 14) -> float:
        if len(prices) < period + 1:
            return 0
        trs = []
        for i in range(1, period + 1):
            tr = abs(prices[-i] - prices[-i-1])
            trs.append(tr)
        return sum(trs) / period


class FundamentalScoringService:
    """Calcule un score fondamental basé sur les ratios financiers."""

    def score(self, fundamental: FundamentalData, current_price: float) -> Dict[str, float]:
        scores = {}
        reasons = []

        # Score PER (inversé: plus bas = mieux)
        if fundamental.per:
            if fundamental.per < 8:
                scores['per'] = 95
                reasons.append(f"PER très attractif: {fundamental.per}")
            elif fundamental.per < 12:
                scores['per'] = 80
                reasons.append(f"PER attractif: {fundamental.per}")
            elif fundamental.per < 18:
                scores['per'] = 60
            elif fundamental.per < 25:
                scores['per'] = 40
                reasons.append(f"PER élevé: {fundamental.per}")
            else:
                scores['per'] = 20
                reasons.append(f"PER très élevé: {fundamental.per}")
        else:
            scores['per'] = 50

        # Score ROE
        if fundamental.roe:
            if fundamental.roe > 25:
                scores['roe'] = 95
                reasons.append(f"ROE excellent: {fundamental.roe}%")
            elif fundamental.roe > 20:
                scores['roe'] = 85
                reasons.append(f"ROE très bon: {fundamental.roe}%")
            elif fundamental.roe > 15:
                scores['roe'] = 70
            elif fundamental.roe > 10:
                scores['roe'] = 50
            else:
                scores['roe'] = 30
                reasons.append(f"ROE faible: {fundamental.roe}%")
        else:
            scores['roe'] = 50

        # Score Dividend Yield
        if fundamental.dividend_yield:
            if fundamental.dividend_yield > 7:
                scores['dividend'] = 95
                reasons.append(f"Yield exceptionnel: {fundamental.dividend_yield}%")
            elif fundamental.dividend_yield > 5:
                scores['dividend'] = 80
                reasons.append(f"Yield attractif: {fundamental.dividend_yield}%")
            elif fundamental.dividend_yield > 3:
                scores['dividend'] = 60
            elif fundamental.dividend_yield > 1:
                scores['dividend'] = 40
            else:
                scores['dividend'] = 20
        else:
            scores['dividend'] = 30
            reasons.append("Pas de dividende")

        # Score dette
        if fundamental.debt_to_equity:
            if fundamental.debt_to_equity < 0.5:
                scores['debt'] = 90
            elif fundamental.debt_to_equity < 1.0:
                scores['debt'] = 70
            elif fundamental.debt_to_equity < 1.5:
                scores['debt'] = 50
            else:
                scores['debt'] = 30
                reasons.append(f"Dette élevée: {fundamental.debt_to_equity}")
        else:
            scores['debt'] = 50

        # Score croissance
        if fundamental.revenue_growth:
            if fundamental.revenue_growth > 20:
                scores['growth'] = 90
                reasons.append(f"Croissance forte: {fundamental.revenue_growth}%")
            elif fundamental.revenue_growth > 10:
                scores['growth'] = 75
            elif fundamental.revenue_growth > 0:
                scores['growth'] = 60
            else:
                scores['growth'] = 30
                reasons.append(f"Croissance négative: {fundamental.revenue_growth}%")
        else:
            scores['growth'] = 50

        return {'scores': scores, 'reasons': reasons}


class SignalGenerationService:
    """Génère les signaux d'achat/vente en combinant technique + fondamental."""

    def __init__(self, tech_service: TechnicalAnalysisService, fund_service: FundamentalScoringService):
        self.tech = tech_service
        self.fund = fund_service

    def generate_signal(
        self,
        ticker: Ticker,
        prices: List[float],
        volumes: List[float],
        fundamental: FundamentalData,
        current_price: float
    ) -> ScoringResult:

        # 1. Indicateurs techniques
        tech = self.tech.calculate_indicators(prices, volumes)

        # 2. Score fondamental
        fund_result = self.fund.score(fundamental, current_price)
        fund_scores = fund_result['scores']
        reasons = fund_result['reasons']

        # 3. Score technique
        tech_score = self._score_technical(tech, current_price, prices[-1] if prices else current_price, volumes)

        # 4. Score risque (volatilité + liquidité)
        risk_score = self._score_risk(tech, volumes)

        # 5. Composite (pondération)
        composite = (
            fund_scores.get('per', 50) * 0.20 +
            fund_scores.get('roe', 50) * 0.20 +
            fund_scores.get('dividend', 50) * 0.15 +
            fund_scores.get('growth', 50) * 0.10 +
            fund_scores.get('debt', 50) * 0.05 +
            tech_score * 0.20 +
            risk_score * 0.10
        )

        # 6. Détermination du signal
        signal, confidence = self._determine_signal(composite, tech, fund_scores, reasons)

        return ScoringResult(
            ticker=ticker,
            score_value=fund_scores.get('per', 50),
            score_growth=fund_scores.get('growth', 50),
            score_dividend=fund_scores.get('dividend', 50),
            score_momentum=tech_score,
            score_risk=risk_score,
            composite_score=composite,
            signal=signal,
            confidence=confidence,
            reasons=reasons
        )

    def _score_technical(self, tech: TechnicalIndicators, price: float, last_price: float, volumes: List[float]) -> float:
        score = 50

        # RSI
        if tech.rsi_14 is not None:
            if tech.rsi_14 < 30:
                score += 20  # Survente = bullish
            elif tech.rsi_14 > 70:
                score -= 20  # Surachat = bearish
            elif 40 <= tech.rsi_14 <= 60:
                score += 5   # Zone neutre = stable

        # Trend (SMA)
        if tech.sma_20 and tech.sma_50:
            if tech.sma_20 > tech.sma_50:
                score += 10  # Golden cross tendency
            else:
                score -= 10

        # MACD
        if tech.macd and tech.macd_signal:
            if tech.macd > tech.macd_signal:
                score += 10
            else:
                score -= 10

        # Bollinger
        if tech.bollinger_upper and tech.bollinger_lower:
            if price < tech.bollinger_lower:
                score += 15  # Prix sous bande inf = survente
            elif price > tech.bollinger_upper:
                score -= 15

        # Volume
        if tech.volume_sma_20 and volumes:
            if volumes[-1] > tech.volume_sma_20 * 1.5:
                score += 5  # Volume anormalement haut = confirmation

        return max(0, min(100, score))

    def _score_risk(self, tech: TechnicalIndicators, volumes: List[float]) -> float:
        score = 70

        # Volatilité (ATR)
        if tech.atr_14:
            if tech.atr_14 > 5:  # > 5% de volatilité moyenne
                score -= 20
            elif tech.atr_14 > 3:
                score -= 10

        # Liquidité
        if volumes:
            avg_vol = sum(volumes[-20:]) / min(20, len(volumes))
            if avg_vol < 1000:
                score -= 15  # Très illiquide
            elif avg_vol < 5000:
                score -= 5

        return max(0, min(100, score))

    def _determine_signal(self, composite: float, tech: TechnicalIndicators, 
                          fund_scores: Dict, reasons: List[str]) -> tuple:

        signal = SignalType.HOLD
        confidence = 0.5

        # Règles de décision
        if composite >= 75:
            if tech.rsi_14 and tech.rsi_14 < 40:
                signal = SignalType.STRONG_BUY
                confidence = 0.85
                reasons.append("Signal FORT ACHAT: Value + Survente technique")
            else:
                signal = SignalType.BUY
                confidence = 0.75
                reasons.append("Signal ACHAT: Fondamentaux solides")
        elif composite >= 60:
            signal = SignalType.BUY
            confidence = 0.65
            reasons.append("Signal ACHAT modéré")
        elif composite <= 25:
            if tech.rsi_14 and tech.rsi_14 > 65:
                signal = SignalType.STRONG_SELL
                confidence = 0.80
                reasons.append("Signal FORTE VENTE: Faible valeur + Surachat")
            else:
                signal = SignalType.SELL
                confidence = 0.70
                reasons.append("Signal VENTE: Fondamentaux faibles")
        elif composite <= 40:
            signal = SignalType.SELL
            confidence = 0.60
            reasons.append("Signal VENTE modéré")
        else:
            reasons.append("Signal NEUTRE: Attendre une meilleure opportunité")

        return signal, confidence


# =============================================================================
# LAYER 3: INTERFACES (Ports - définitions abstraites)
# =============================================================================

from abc import ABC, abstractmethod

class MarketDataRepository(ABC):
    """Port pour la récupération des données de marché."""

    @abstractmethod
    def fetch_historical(self, ticker: Ticker, start: date, end: date) -> List[MarketData]:
        pass

    @abstractmethod
    def fetch_latest(self, ticker: Ticker) -> Optional[MarketData]:
        pass

    @abstractmethod
    def fetch_all_tickers(self) -> List[Ticker]:
        pass

class FundamentalDataRepository(ABC):
    """Port pour la récupération des données fondamentales."""

    @abstractmethod
    def fetch(self, ticker: Ticker) -> Optional[FundamentalData]:
        pass

class NotificationService(ABC):
    """Port pour les notifications (push, email, webhook)."""

    @abstractmethod
    def notify(self, signal: ScoringResult) -> None:
        pass

class OrderGateway(ABC):
    """Port pour l'envoi d'ordres (via SGI / FIX)."""

    @abstractmethod
    def place_order(self, ticker: str, side: str, quantity: int, price: float, order_type: str) -> dict:
        pass


# =============================================================================
# LAYER 4: INFRASTRUCTURE (Adapters - implémentations concrètes)
# =============================================================================

import requests
import csv
import io

class GitHubBRVMDataAdapter(MarketDataRepository):
    """
    Adapter pour brvm-data-public (GitHub)
    URL: https://github.com/Fredysessie/brvm-data-public
    Mise à jour: toutes les 15 min (9h-15h UTC, Lundi-Vendredi)
    """

    BASE_URL = "https://raw.githubusercontent.com/Fredysessie/brvm-data-public/main/data"

    def fetch_historical(self, ticker: Ticker, start: date, end: date) -> List[MarketData]:
        url = f"{self.BASE_URL}/{ticker.symbol}/{ticker.symbol}.daily.csv"
        response = requests.get(url, timeout=10)
        response.raise_for_status()

        data = []
        reader = csv.DictReader(io.StringIO(response.text))
        for row in reader:
            d = datetime.strptime(row['Date'], '%Y-%m-%d').date()
            if start <= d <= end:
                data.append(MarketData(
                    ticker=ticker,
                    date=d,
                    open_price=float(row['Open']),
                    high=float(row['High']),
                    low=float(row['Low']),
                    close=float(row['Close']),
                    volume=int(row['Volume']),
                    source=DataSource.GITHUB_BRVM
                ))
        return data

    def fetch_latest(self, ticker: Ticker) -> Optional[MarketData]:
        historical = self.fetch_historical(ticker, date(2020, 1, 1), date(2099, 12, 31))
        return historical[-1] if historical else None

    def fetch_all_tickers(self) -> List[Ticker]:
        # Liste des 66 tickers disponibles
        tickers = [
            Ticker("SNTS", "Sonatel", "Sénégal", "Télécom"),
            Ticker("ORAC", "Orange CI", "Côte d'Ivoire", "Télécom"),
            Ticker("SGBC", "SGB CI", "Côte d'Ivoire", "Finance"),
            Ticker("ECOC", "Ecobank CI", "Côte d'Ivoire", "Finance"),
            Ticker("PALC", "PalmCI", "Côte d'Ivoire", "Agriculture"),
            Ticker("SPHC", "SAPH CI", "Côte d'Ivoire", "Agriculture"),
            # ... 60 autres tickers
        ]
        return tickers


class SikafinanceAdapter(MarketDataRepository):
    """
    Adapter pour Sikafinance (téléchargement CSV historique)
    URL: https://www.sikafinance.com/marches/download/{SYMBOL}
    """

    BASE_URL = "https://www.sikafinance.com/marches/download"

    def fetch_historical(self, ticker: Ticker, start: date, end: date) -> List[MarketData]:
        url = f"{self.BASE_URL}/{ticker.symbol}"
        response = requests.get(url, timeout=15)
        # Parsing CSV spécifique à Sikafinance
        return self._parse_csv(response.text, ticker, start, end)

    def _parse_csv(self, text: str, ticker: Ticker, start: date, end: date) -> List[MarketData]:
        data = []
        reader = csv.DictReader(io.StringIO(text), delimiter=';')
        for row in reader:
            d = datetime.strptime(row['Date'], '%d/%m/%Y').date()
            if start <= d <= end:
                data.append(MarketData(
                    ticker=ticker,
                    date=d,
                    open_price=float(row['Ouverture'].replace(',', '.')),
                    high=float(row['Plus Haut'].replace(',', '.')),
                    low=float(row['Plus Bas'].replace(',', '.')),
                    close=float(row['Cloture'].replace(',', '.')),
                    volume=int(row['Volume']),
                    source=DataSource.SIKAFINANCE
                ))
        return data

    def fetch_latest(self, ticker: Ticker) -> Optional[MarketData]:
        return None  # Sikafinance n'a pas de flux temps réel

    def fetch_all_tickers(self) -> List[Ticker]:
        return []


class FluxBourseAdapter(MarketDataRepository, FundamentalDataRepository):
    """
    Adapter pour FluxBourse (cours temps réel + PER + rendement)
    URL: https://fluxbourse.com/
    Nécessite scraping ou API si disponible
    """

    def fetch_historical(self, ticker: Ticker, start: date, end: date) -> List[MarketData]:
        # Scraping ou API REST
        return []

    def fetch_latest(self, ticker: Ticker) -> Optional[MarketData]:
        # Récupération temps réel après chaque BOC
        return None

    def fetch_all_tickers(self) -> List[Ticker]:
        return []

    def fetch(self, ticker: Ticker) -> Optional[FundamentalData]:
        # Récupère PER, rendement, BPA depuis FluxBourse
        return None


class BRVMIntelligenceAdapter(MarketDataRepository, FundamentalDataRepository):
    """
    Adapter pour BRVM Intelligence (API IA + screener + signaux)
    URL: https://brvmintelligence.com/
    """

    def fetch_historical(self, ticker: Ticker, start: date, end: date) -> List[MarketData]:
        return []

    def fetch_latest(self, ticker: Ticker) -> Optional[MarketData]:
        return None

    def fetch_all_tickers(self) -> List[Ticker]:
        return []

    def fetch(self, ticker: Ticker) -> Optional[FundamentalData]:
        return None


class ConsoleNotificationAdapter(NotificationService):
    """Adapter pour notification console (dev/test)."""

    def notify(self, signal: ScoringResult) -> None:
        print(f"\n{'='*60}")
        print(f"SIGNAL: {signal.signal.value} | {signal.ticker.symbol} - {signal.ticker.name}")
        print(f"Score Composite: {signal.composite_score:.1f}/100 | Confiance: {signal.confidence:.0%}")
        print(f"Raison(s):")
        for r in signal.reasons:
            print(f"  - {r}")
        print(f"{'='*60}\n")


class MockOrderGateway(OrderGateway):
    """Adapter mock pour tests (sans envoi réel à la SGI)."""

    def place_order(self, ticker: str, side: str, quantity: int, price: float, order_type: str) -> dict:
        print(f"[MOCK ORDER] {side} {quantity} {ticker} @ {price} ({order_type})")
        return {"status": "simulated", "order_id": f"MOCK-{ticker}-{datetime.now().timestamp()}"}


# =============================================================================
# LAYER 5: APPLICATION (Orchestration - Controllers / Schedulers)
# =============================================================================

class BRVMTradingEngine:
    """
    Moteur principal d'orchestration.
    Responsabilité: coordonner la collecte, l'analyse et la décision.
    """

    def __init__(
        self,
        market_repo: MarketDataRepository,
        fund_repo: FundamentalDataRepository,
        signal_service: SignalGenerationService,
        notifier: NotificationService,
        order_gateway: OrderGateway
    ):
        self.market_repo = market_repo
        self.fund_repo = fund_repo
        self.signal_service = signal_service
        self.notifier = notifier
        self.order_gateway = order_gateway

    def analyze_ticker(self, ticker: Ticker, lookback_days: int = 90) -> ScoringResult:
        """Pipeline complet: collecte -> analyse -> signal."""

        # 1. COLLECTE
        end = date.today()
        start = end - __import__('datetime').timedelta(days=lookback_days)

        market_data = self.market_repo.fetch_historical(ticker, start, end)
        if not market_data:
            raise ValueError(f"Aucune donnée pour {ticker.symbol}")

        fundamental = self.fund_repo.fetch(ticker)
        if not fundamental:
            fundamental = FundamentalData(ticker=ticker)

        # 2. TRANSFORMATION
        prices = [d.close for d in market_data]
        volumes = [d.volume for d in market_data]
        current_price = prices[-1]

        # 3. ANALYSE & SIGNAL
        result = self.signal_service.generate_signal(
            ticker=ticker,
            prices=prices,
            volumes=volumes,
            fundamental=fundamental,
            current_price=current_price
        )

        # 4. NOTIFICATION
        self.notifier.notify(result)

        return result

    def scan_all_tickers(self, tickers: List[Ticker]) -> List[ScoringResult]:
        """Scan complet du marché BRVM."""
        results = []
        for ticker in tickers:
            try:
                result = self.analyze_ticker(ticker)
                results.append(result)
            except Exception as e:
                print(f"Erreur sur {ticker.symbol}: {e}")
        return results

    def execute_strategy(self, result: ScoringResult, portfolio: dict, max_risk_per_trade: float = 0.02):
        """Exécution d'une stratégie basée sur le signal."""

        if result.signal in (SignalType.STRONG_BUY, SignalType.BUY):
            # Calcul position sizing
            portfolio_value = portfolio.get('total_value', 0)
            risk_amount = portfolio_value * max_risk_per_trade

            # Sizing basé sur l'ATR (volatilité)
            # Simplifié: investir 5% du portefeuille par signal BUY
            invest_amount = portfolio_value * 0.05
            quantity = int(invest_amount / result.ticker.symbol)  # simplifié

            # Ordre à cours limité (recommandé BRVM)
            self.order_gateway.place_order(
                ticker=result.ticker.symbol,
                side="BUY",
                quantity=100,  # minimum BRVM
                price=0,  # prix actuel
                order_type="LIMIT"
            )

        elif result.signal in (SignalType.STRONG_SELL, SignalType.SELL):
            # Vente si position détenue
            holdings = portfolio.get('holdings', {})
            if result.ticker.symbol in holdings:
                self.order_gateway.place_order(
                    ticker=result.ticker.symbol,
                    side="SELL",
                    quantity=holdings[result.ticker.symbol],
                    price=0,
                    order_type="LIMIT"
                )


# =============================================================================
# EXEMPLE D'UTILISATION
# =============================================================================

if __name__ == "__main__":
    # 1. Construction des dépendances (Dependency Injection)
    market_repo = GitHubBRVMDataAdapter()
    fund_repo = FluxBourseAdapter()  # ou BRVMIntelligenceAdapter
    tech_service = TechnicalAnalysisService()
    fund_service = FundamentalScoringService()
    signal_service = SignalGenerationService(tech_service, fund_service)
    notifier = ConsoleNotificationAdapter()
    order_gateway = MockOrderGateway()

    # 2. Moteur
    engine = BRVMTradingEngine(
        market_repo=market_repo,
        fund_repo=fund_repo,
        signal_service=signal_service,
        notifier=notifier,
        order_gateway=order_gateway
    )

    # 3. Analyse d'une valeur
    sonatel = Ticker("SNTS", "Sonatel", "Sénégal", "Télécom")
    result = engine.analyze_ticker(sonatel, lookback_days=60)

    # 4. Scan complet
    all_tickers = market_repo.fetch_all_tickers()
    all_results = engine.scan_all_tickers(all_tickers)

    # 5. Filtrer les meilleures opportunités
    buy_signals = [r for r in all_results if r.signal in (SignalType.BUY, SignalType.STRONG_BUY)]
    buy_signals.sort(key=lambda x: x.composite_score, reverse=True)

    print(f"\nTop 5 opportunités d'achat:")
    for r in buy_signals[:5]:
        print(f"  {r.ticker.symbol}: {r.signal.value} (Score: {r.composite_score:.1f})")
