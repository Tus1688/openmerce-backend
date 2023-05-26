package freight

type CalculateFreightRequest struct {
	ID     uint32  `json:"id"`
	Weight float64 `json:"weight"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type PrecalculateFreightRequest struct {
	ID     uint32
	Weight float64
	Volume float64
}

type ServiceRate struct {
	ProductCode string `json:"product_code"`
	ProductName string `json:"product_name"`
	Etd         string `json:"etd"`
	Rates       int    `json:"rates"`
}

type WholeResult struct {
	Anteraja []ServiceRate `json:"anteraja"`
	Sicepat  []ServiceRate `json:"sicepat"`
}
