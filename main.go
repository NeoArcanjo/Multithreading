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

	data := <-response
	log.Println(data)
}

func BuscaCEP(ctx context.Context, provider string, cep string, response chan ResponseCEP) {
	switch provider {
	case "brasilapi":
		url := "https://brasilapi.com.br/api/cep/v1/" + cep
		res, err := BuscaCEPClient(ctx, url)
		var data BrasilApiCEP
		err = json.Unmarshal(res, &data)
		if err != nil {
			log.Println(err)
		}
		response <- ResponseCEP{Cep: data.Cep, State: data.State, City: data.City, Neighborhood: data.Neighborhood, Street: data.Street, Provider: "brasilapi"}
		ctx.Done()
	case "viacep":
		url := "https://viacep.com.br/ws/" + cep + "/json/"
		res, err := BuscaCEPClient(ctx, url)
		var data ViaCEP
		err = json.Unmarshal(res, &data)
		if err != nil {
			log.Println(err)
		}
		response <- ResponseCEP{Cep: data.Cep, State: data.Uf, City: data.Localidade, Neighborhood: data.Bairro, Street: data.Logradouro, Provider: "viacep"}
		ctx.Done()
	}
}

func BuscaCEPClient(ctx context.Context, url string) ([]byte, error) {
	select {
	case <-ctx.Done():
		log.Println("Timeout reached")
		return nil, ctx.Err()
	default:
		req, err := http.Get(url)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer req.Body.Close()
		res, err := io.ReadAll(req.Body)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return res, nil
	}
}
