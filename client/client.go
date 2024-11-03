package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const URL = "http://localhost:8080/cotacao"

func main() {
	log.Println("Iniciando a consulta da cotação do dólar")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	dolar, err := cotacaoDolar(ctx)
	if err != nil {
		log.Fatalf("Erro ao buscar cotação do dólar: %v", ctx.Err())
		return
	}

	log.Println("Salvando cotação do dólar")
	err = salvarCotacao(dolar)
	if err != nil {
		log.Fatalf("Erro ao salvar cotação do dólar: %v", err)
	}
	log.Println("Cotação salva com sucesso")
}

func cotacaoDolar(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func salvarCotacao(bid string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("Dólar: " + bid)
	if err != nil {
		return err
	}
	return nil
}
