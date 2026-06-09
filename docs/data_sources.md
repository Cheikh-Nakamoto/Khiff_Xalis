# Sources de Données du Collector BRVM

Ce document liste et détaille de manière approfondie toutes les sources de données externes utilisées par le composant `go-collector`. Il documente également chaque donnée potentiellement récupérée et son utilité précise dans les modèles de calcul du moteur de trading.

Toutes les URLs fixes (hormis celles configurées dynamiquement par environnement) sont centralisées et documentées au sein du fichier de code `go-collector/internal/links.go`.

---

## 1. Données Macroéconomiques (Banque Mondiale & BCEAO)

### 1.1 Inflation Annuelle
- **Lien de l'API** : `https://api.worldbank.org/v2/country/{countryCode}/indicator/FP.CPI.TOTL.ZG?format=json&date=2023:2026&per_page=5`
- **Données potentiellement récupérées** : Historique récent et valeur actuelle de l'inflation (exprimée en pourcentage annuel) par pays de l'UEMOA (Côte d'Ivoire, Sénégal, Togo, Burkina Faso, Bénin, Mali, Niger, Guinée-Bissau). Seule la donnée la plus récente est extraite.
- **Utilité détaillée** : L'inflation affecte massivement le pouvoir d'achat des ménages dans l'espace sous-régional. Elle a un impact majeur sur les prévisions de consommation (ex: impact sur les valeurs du secteur de la distribution de détail) et sur les charges de production (ex: entreprises industrielles). Le modèle de trading intègre cette variable pour ajuster le score de croissance attendue d'un titre lié à l'économie locale.

### 1.2 Taux Directeur BCEAO
- **Lien du Scraping** : `https://www.brvm.org/fr/taux-directeur`
- **Données potentiellement récupérées** : Le taux d'intérêt de base (taux directeur) appliqué par la Banque Centrale (ex: 3.50%). En cas d'échec du scraping, la dernière valeur enregistrée en base de données ou un taux de secours par défaut est utilisé.
- **Utilité détaillée** : Il s'agit du coût fondamental de l'argent. Un taux élevé rend le crédit cher, ce qui limite l'investissement des entreprises cotées à la BRVM et rend les placements obligataires (Titres Publics) beaucoup plus attractifs que les actions. C'est une donnée pivot pour le module de répartition de portefeuille (Asset Allocation) afin d'arbitrer entre actions et obligations.

### 1.3 Taux de Change (XOF/USD)
- **Lien de l'API** : `https://open.er-api.com/v6/latest/USD`
- **Données potentiellement récupérées** : Le dictionnaire complet des parités monétaires vis-à-vis du Dollar Américain. Le système extrait spécifiquement le taux "XOF" (Franc CFA BCEAO).
- **Utilité détaillée** : Bien que le XOF soit arrimé à l'Euro, l'immense majorité des marchés internationaux de matières premières (pétrole, or, cacao) se négocie en Dollars USD. Ce taux de change sert de coefficient multiplicateur pour traduire en devise locale l'impact réel des variations de prix mondiales. Un baril de pétrole stable en dollars peut coûter plus cher aux entreprises locales si le Franc CFA se déprécie face au Dollar.

---

## 2. Matières Premières (Commodities)

Le système scrute les matières premières stratégiques qui corroborent directement le bilan financier des valeurs de la BRVM.

### 2.1 Cacao
- **Lien de l'API** : `https://www.icco.org/wp-json/icco/v1/daily-prices?limit=1`
- **Données potentiellement récupérées** : Le dernier prix journalier du cacao, exprimé en USD par tonne, fourni par l'International Cocoa Organization.
- **Utilité détaillée** : Principal produit d'exportation de la sous-région, particulièrement de la Côte d'Ivoire. Un prix du cacao élevé dynamise l'économie ivoirienne entière et augmente la masse monétaire en circulation (favorable aux banques locales). En revanche, ce même prix élevé est un surcoût pour certaines sociétés agroalimentaires de transformation cotées (ex: Nestlé CI), dont le modèle ajuste alors le score de risque à la hausse.

### 2.2 Pétrole, Or, Caoutchouc et Huile de Palme
- **Lien de l'API** : `https://api.tradingeconomics.com/markets/commodities?c=guest:guest&f=json`
- **Données potentiellement récupérées** : Un panel de cours de marché (en JSON) pour diverses matières premières. Le collector filtre pour récupérer les cours du "Brent" (USD/barrel), "Gold" (USD/troy oz), "Rubber" (USD/kg) et "Palm Oil" (USD/tonne).
- **Utilité détaillée** :
  - **Pétrole (Brent)** : Impacte les coûts d'approvisionnement (transport/logistique) de presque toutes les entreprises. Impact direct sur le chiffre d'affaires des sociétés de distribution de carburant et de raffinage (SMB, Total SN, Total CI).
  - **Or** : Importante valeur d'exportation pour le Mali, le Burkina Faso et la Côte d'Ivoire, l'or influence positivement leurs balances commerciales et renforce la santé du système bancaire sous-régional en période d'incertitude.
  - **Caoutchouc & Huile de Palme** : Ces deux commodités dictent de manière ultra-corrélée le chiffre d'affaires de mastodontes agricoles cotés à la BRVM tels que la SAPH (caoutchouc), la SOGB (caoutchouc/palme) et PalmCI (palme). Leurs variations dirigent les signaux d'achat ou de vente de ces titres (Beta de sensibilité).

### 2.3 Anacarde (Noix de Cajou)
- **Source Actuelle** : Génération / Fallback interne algorithmique (simulateur dynamique).
- **Données potentiellement récupérées** : Un prix reconstitué autour d'une moyenne de marché (USD/tonne).
- **Utilité détaillée** : Face à l'absence d'API mondiale liquide (L'anacarde se négocie souvent en gré à gré), le système génère un prix de base. Étant une exportation agricole majeure, ce prix joue sur le score global de la croissance économique régionale et impacte les sociétés de transport ou de l'agro-industrie (ex: acteurs logistiques liés aux campagnes de cajou).

---

## 3. Risque Politique & Stabilité Régionale (GDELT)

### 3.1 Veille des Crises et Émeutes
- **Lien de l'API** : `https://api.gdeltproject.org/api/v2/summary/summary?theme=POLITICAL_VIOLENCE&country={countryName}&format=json&timespan=7days`
- **Données potentiellement récupérées** : Une métrique (`eventcount`) quantifiant précisément le nombre de couvertures médiatiques et d'incidents signalés sur le thème de la violence politique (coups d'état, grèves dures, terrorisme, émeutes) dans un pays spécifique de la CEDEAO sur les 7 derniers jours.
- **Utilité détaillée** : Le marché ouest-africain est parfois sensible à la conjoncture socio-politique. Ce module assigne un score de stabilité à chaque pays. Si le seuil d'alerte des événements violents est franchi (ex: > 35 événements), le système comprend qu'une crise majeure est en cours.
- **Action Déclenchée** : En cas de "Crise", le système peut invalider les modèles de prédiction usuels et forcer une commande de type "Signal Freeze" (Gel des prises de positions) sur tous les titres exposés majoritairement à l'économie de ce pays, par principe de précaution.

---

## 4. Données de Marché Historiques BRVM (GitHub Public)

### 4.1 Cours de Bourse (OHLCV)
- **Lien du Référentiel** : `[GITHUB_CSV_URL]/{ticker}.csv` (L'URL de base est configurée par la variable d'environnement `GITHUB_CSV_URL`).
- **Données potentiellement récupérées** : Des fichiers CSV pour chaque société cotée (ex: `SNTS.csv` pour Sonatel) contenant les historiques boursiers : Date, Open (Ouverture), High (Plus haut), Low (Plus bas), Close (Clôture), et Volume d'échange.
- **Utilité détaillée** : Ces fichiers CSV constituent la mémoire à long terme du moteur de trading. Ces données pures sont ingérées par lots (Batch Insert) dans la table de base de données `market_data`. Elles alimentent ensuite l'analyse technique, le calcul des moyennes mobiles (SMA, EMA), le backtesting et la calibration de l'intelligence artificielle pour prévoir les tendances de chaque action (ticker) de la BRVM.
