package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const URL = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type CotacaoDolar struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Dolar struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = createTable(db)
	if err != nil {
		panic(err)
	}

	log.Println("Inicializando servidor")
	http.HandleFunc("/cotacao", buscaCotacaoDolarHandler)
	http.ListenAndServe(":8080", nil)
}

func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high REAL,
		low REAL,
		varBid REAL,
		pctChange REAL,
		bid REAL,
		ask REAL,
		timestamp INTEGER,
		createDate TEXT
	);
	`
	_, err := db.Exec(query)
	return err
}

func buscaCotacaoDolarHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Println("Buscando cotação do dólar")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	cotacao, err := BuscaCotacaoDolar(ctx)
	if err != nil {
		log.Println("Erro ao realizar a requisição - ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("Cotação do dólar:", cotacao.Usdbrl.Bid)

	log.Println("Salvando cotação no banco de dados")
	ctxdb, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = insertDBCotacao(ctxdb, db, cotacao)
	if err != nil {
		log.Println("Erro ao persistir a cotação no banco de dados - ", err)
		return
	}
	log.Println("Cotação salva no banco de dados")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao.Usdbrl.Bid)
}

func BuscaCotacaoDolar(ctx context.Context) (*CotacaoDolar, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
	if err != nil {
		return nil, err
	}
	log.Println("Fazendo requisição")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var c CotacaoDolar
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func insertDBCotacao(ctx context.Context, db *sql.DB, cotacao *CotacaoDolar) error {
	stmt, err := db.Prepare("INSERT INTO cotacoes (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, createDate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, cotacao.Usdbrl.Code, cotacao.Usdbrl.Codein, cotacao.Usdbrl.Name, cotacao.Usdbrl.High, cotacao.Usdbrl.Low, cotacao.Usdbrl.VarBid, cotacao.Usdbrl.PctChange, cotacao.Usdbrl.Bid, cotacao.Usdbrl.Ask, cotacao.Usdbrl.Timestamp, cotacao.Usdbrl.CreateDate)
	if err != nil {
		return err
	}
	return nil
}
