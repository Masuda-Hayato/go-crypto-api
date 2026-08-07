package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CoinData struct {
	JPY float64 `json:"jpy"`
}

type APIResponse struct {
	Bitcoin CoinData `json:"bitcoin"`
}

func main() {
	db, err := sql.Open("sqlite3", "./market.db")
	if err != nil {
		log.Fatal("DB接続エラー:", err)
	}
	defer db.Close()

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS prices (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT NOT NULL,
		price REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, _ = db.Exec(createTableSQL)

	fmt.Println("🚀 異常検知システム起動（1分ごとに監視中...）")
	fmt.Println("⚠️ 停止するには [Ctrl] + [C] を押してください。")
	fmt.Println("--------------------------------------------------")

	// 🌟 【ポイント1】前回の価格を記憶するための変数を準備（初期値は0）
	var lastPrice float64 = 0.0

	// 🌟 【ポイント2】関数に「変数の住所（&）」を渡す
	fetchAndSavePrice(db, &lastPrice)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// ここでも「変数の住所」を渡して、関数の中で書き換えてもらう
		fetchAndSavePrice(db, &lastPrice)
	}
}

// 🌟 【ポイント3】受け取る側に「*」をつけることで、共有のメモ帳として使える
func fetchAndSavePrice(db *sql.DB, lastPrice *float64) {
	now := time.Now().Format("15:04:05")

	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=jpy"
	resp, err := http.Get(url)
	if err != nil {
		log.Println("⚠️ API通信エラー")
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var response APIResponse
	json.Unmarshal(body, &response)

	symbol := "Bitcoin"
	currentPrice := response.Bitcoin.JPY

	insertSQL := `INSERT INTO prices (symbol, price) VALUES (?, ?)`
	_, err = db.Exec(insertSQL, symbol, currentPrice)
	if err != nil {
		log.Println("⚠️ データ保存エラー")
		return
	}

	fmt.Printf("[%s] ✅ 取得: %.0f円\n", now, currentPrice)

	// ==========================================
	// 🚨 金融ロジック：異常変動の検知アラート
	// ==========================================
	// *lastPrice が 0 じゃない（＝初回実行じゃない）時だけ比較する
	if *lastPrice != 0.0 {
		// 今回の価格と前回の価格の差額を計算
		diff := currentPrice - *lastPrice
		
		// 🌟 アラートを出す基準額（今回は5,000円に設定）
		threshold := 5000.0 

		if diff >= threshold {
			// 急騰した時
			fmt.Printf("   📈 【急騰アラート】 1分間で +%.0f円 上昇しました！\n", diff)
		} else if diff <= -threshold {
			// 急落した時
			fmt.Printf("   📉 【急落アラート】 1分間で %.0f円 下落しました！\n", -diff) // マイナスを外して表示
		}
	}

	// 今回の価格を、次回の比較のために「前回の価格」としてメモ帳に上書き保存
	*lastPrice = currentPrice
}