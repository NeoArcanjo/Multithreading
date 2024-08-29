package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type BrasilApiCEP struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type ResponseCEP struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Provider     string `json:"provider"`
}


func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	response := make(chan ResponseCEP)

	go BuscaCEP(ctx, "brasilapi", "29102-385", response)
	go BuscaCEP(ctx, "viacep", "29102-385", response)

	select {
	case data := <-response:
		log.Println(data)
	case <-ctx.Done():
		log.Println("Timeout reached, no response received")
	}
}

func BuscaCEP(ctx context.Context, provider string, cep string, response chan ResponseCEP) {
	var url string
	switch provider {
	case "brasilapi":
		url = "https://brasilapi.com.br/api/cep/v1/" + cep
	case "viacep":
		url = "https://viacep.com.br/ws/" + cep + "/json/"
	default:
		log.Println("Unknown provider:", provider)
		return
	}

	res, err := BuscaCEPClient(ctx, url, provider)
	if err != nil {
		log.Println("Error fetching data from", provider, ":", err)
		return
	}

	var data ResponseCEP
	switch provider {
	case "brasilapi":
		var brasilApiData BrasilApiCEP
		if err := json.Unmarshal(res, &brasilApiData); err != nil {
			log.Println("Error unmarshalling BrasilApiCEP data:", err)
			return
		}
		data = ResponseCEP{
			Cep:          brasilApiData.Cep,
			State:        brasilApiData.State,
			City:         brasilApiData.City,
			Neighborhood: brasilApiData.Neighborhood,
			Street:       brasilApiData.Street,
			Provider:     "brasilapi",
		}
	case "viacep":
		var viaCepData ViaCEP
		if err := json.Unmarshal(res, &viaCepData); err != nil {
			log.Println("Error unmarshalling ViaCEP data:", err)
			return
		}
		data = ResponseCEP{
			Cep:          viaCepData.Cep,
			State:        viaCepData.Uf,
			City:         viaCepData.Localidade,
			Neighborhood: viaCepData.Bairro,
			Street:       viaCepData.Logradouro,
			Provider:     "viacep",
		}
	}

	select {
	case response <- data:
	case <-ctx.Done():
		log.Println("Context cancelled before sending response")
	}
}

func BuscaCEPClient(ctx context.Context, url string, provider string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}
