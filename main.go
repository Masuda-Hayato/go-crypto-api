package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type CoinData struct {
	JPY float64 `json:"jpy"`
}

type APIResponse struct {
	Bitcoin CoinData `json:"bitcoin"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .envファイルが見つかりません。環境変数を直接使用します。")
	}

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
		threshold := 100.0

		if diff >= threshold {
			// 急騰した時
			msg := fmt.Sprintf("📈 【急騰】1分間で +%.0f円 上昇しました！\n現在の価格: %.0f円", diff, currentPrice)
			fmt.Println("   " + msg)
			sendLineMessage(msg) // 🌟 ここでLINE関数を呼び出す！

		} else if diff <= -threshold {
			// 急落した時
			msg := fmt.Sprintf("📉 【急落】1分間で %.0f円 下落しました！\n現在の価格: %.0f円", -diff, currentPrice)
			fmt.Println("   " + msg)
			sendLineMessage(msg) // 🌟 ここでLINE関数を呼び出す！
		}
	}

	// 今回の価格を、次回の比較のために「前回の価格」としてメモ帳に上書き保存
	*lastPrice = currentPrice
}

// ==========================================
// 📱 LINEにメッセージを送信する関数
// ==========================================
func sendLineMessage(message string) {
	// ⚠️ ここに先ほど取得した2つのキーを貼り付けます
	token := os.Getenv("LINE_TOKEN")
	userID := os.Getenv("LINE_USER_ID")

	url := "https://api.line.me/v2/bot/message/push"

	// LINE APIが要求するJSONの形（マトリョーシカ）を作る
	payload := map[string]interface{}{
		"to": userID,
		"messages": []map[string]string{
			{
				"type": "text",
				"text": message,
			},
		},
	}

	// GoのデータをJSONに変換
	body, _ := json.Marshal(payload)

	// 通信の準備
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Println("LINEリクエスト作成エラー:", err)
		return
	}

	// 相手に「JSON形式です」「私は正規のユーザーです」と伝えるヘッダー
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// いざ送信！
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("LINE送信エラー:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   📩 LINEに通知を送信しました！")
	} else {
		fmt.Println("   ⚠️ LINE通知失敗:", resp.StatusCode)
	}
}
