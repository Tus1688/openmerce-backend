package anteraja

type AnterajaAPIRequest struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Weight      string `json:"weight"`
	Length      string `json:"length"`
	Width       string `json:"width"`
	Height      string `json:"height"`
}

type AnterajaAPIResponse struct {
	Status  float64 `json:"status"`
	Info    string  `json:"info"`
	Content struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		Services    []struct {
			ProductCode string  `json:"product_code"`
			ProductName string  `json:"product_name"`
			Etd         string  `json:"etd"`
			Rates       float64 `json:"rates"`
			IsCod       bool    `json:"is_cod,omitempty"`
			ImgUrl      string  `json:"img_url"`
			Idx         float64 `json:"idx"`
			MsgId       string  `json:"msg_id"`
			MsgEn       string  `json:"msg_en"`
			Enable      bool    `json:"enable"`
		} `json:"services"`
		Rates  float64 `json:"rates"`
		Maxsds string  `json:"maxsds"`
		Weight float64 `json:"weight"`
	} `json:"content"`
}

type AnterajaServiceRate struct {
	ProductCode string  `json:"product_code"`
	ProductName string  `json:"product_name"`
	Etd         string  `json:"etd"`
	Rates       float64 `json:"rates"`
}
