package model

type ExchangeAPI []struct {
	Name         string `json:"Name"`
	Code         string `json:"Code"`
	OperatingMIC string `json:"OperatingMIC"`
	Country      string `json:"Country"`
	Currency     string `json:"Currency"`
	CountryISO2  string `json:"CountryISO2"`
	CountryISO3  string `json:"CountryISO3"`
}

type SymbolApi struct {
	Name     string `json:"Name"`
	Code     string `json:"Code"`
	Country  string `json:"Country"`
	Currency string `json:"Currency"`
	Type     string `json:"Type"`
}
