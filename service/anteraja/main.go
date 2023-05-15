package anteraja

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func GetRates(client *http.Client, ctx context.Context, origin, destination, weight, length, width, height string) ([]AnterajaServiceRate, error) {
	url := "https://anteraja.id/api/trackparcel/getRates"
	jsonBytes, err := json.Marshal(AnterajaAPIRequest{
		Origin:      origin,
		Destination: destination,
		Weight:      weight,
		Length:      length,
		Width:       width,
		Height:      height,
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36")
	client.Jar = nil
	//response, err := client.Do(request) with context
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var result AnterajaAPIResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	//	bind type into AnterajaServiceRate
	var rates []AnterajaServiceRate
	for _, service := range result.Content.Services {
		if service.Enable {
			rates = append(rates, AnterajaServiceRate{
				ProductCode: service.ProductCode,
				ProductName: service.ProductName,
				Etd:         service.Etd,
				Rates:       service.Rates,
			})
		}
	}
	return rates, nil
}
