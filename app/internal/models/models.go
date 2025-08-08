package models

type InfoFilterSheet struct {
	Exchange                 string   `json:"exchange"`
	BanksName                []string `json:"banks_name"`
	MonthOrder               int      `json:"month_order"`
	MonthFinishRate          float32  `json:"month_finish_rate"`
	MaxLowSingleTransAmount  float32  `json:"max_low_single_trans_amount"`
	MinHighSingleTransAmount float32  `json:"min_high_single_trans_amount"`
	AverageSize              int      `json:"average_size"`
}

type Soup struct {
	Name          string            `json:"name"`
	FixedPrice    float32           `json:"fixed_price"`
	BestPrice     float32           `json:"best_price"`
	BestPriceLink string            `json:"best_price_link"`
	Date          string            `json:"date"`
	MoneySupply   float32           `json:"money_supply"`
	AverageSize   int               `json:"average_size"`
	InfoFilters   []InfoFilterSheet `json:"info_filters"`
	Data          [][]string        `json:"data"`
}

type RAWData struct {
	Date string     `json:"date"`
	Data [][]string `json:"raw_data"`
}

type SheetData struct {
	Fiat     string  `json:"fiat"`
	SoupList []Soup  `json:"soup_list"`
	RAWData  RAWData `json:"raw_data"`
}
