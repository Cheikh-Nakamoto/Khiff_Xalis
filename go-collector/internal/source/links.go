// Copyright 2026 BRVM Trading Engine Contributors
// Licensed under the GNU AGPL-3.0. See LICENSE for details.

package source

const (
	// WorldBankInflationURL is the URL format string to fetch inflation data per country.
	WorldBankInflationURL = "https://api.worldbank.org/v2/country/%s/indicator/FP.CPI.TOTL.ZG?format=json&date=2023:2026&per_page=5"

	// BRVMTauxDirecteurURL is the URL to scrape the BCEAO taux directeur.
	BRVMTauxDirecteurURL = "https://www.brvm.org/fr/taux-directeur"

	// ERAPIExchangeRateURL is the URL to fetch USD based exchange rates.
	ERAPIExchangeRateURL = "https://open.er-api.com/v6/latest/USD"

	// ICCOCocoaPriceURL is the URL to fetch daily cocoa prices.
	ICCOCocoaPriceURL = "https://www.icco.org/wp-json/icco/v1/daily-prices?limit=1"

	// TradingEconomicsCommoditiesURL is the URL to fetch guest commodity prices.
	TradingEconomicsCommoditiesURL = "https://api.tradingeconomics.com/markets/commodities?c=guest:guest&f=json"

	// GDELTPoliticalViolenceURL is the URL format string to fetch political violence events per country.
	GDELTPoliticalViolenceURL = "https://api.gdeltproject.org/api/v2/summary/summary?theme=POLITICAL_VIOLENCE&country=%s&format=json&timespan=7days"
)
