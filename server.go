package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"text/template"

	_ "github.com/mattn/go-sqlite3"
)

// データベースから取り出したデータを入れる構造体
type PriceRecord struct {
	ID        int     `json:"id"`
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

// ブラウザに送信するHTMLのテンプレート
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>金融データ監視ダッシュボード</title>
    <style>
        body { font-family: sans-serif; background-color: #f4f4f9; padding: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 80%; max-width: 800px; background-color: white; box-shadow: 0 1px 3px rgba(0,0,0,0.2); }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: center; }
        th { background-color: #0056b3; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>📈 金融データ監視ダッシュボード</h1>
    <p>データベースに保存された直近30件の価格履歴を表示しています。</p>
    <table>
        <tr>
            <th>ID</th>
            <th>銘柄</th>
            <th>価格 (円)</th>
            <th>取得日時</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.Symbol}}</td>
            <td>{{.Price}}</td>
            <td>{{.CreatedAt}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`

// ブラウザからアクセスがあった時に実行される関数
func handler(w http.ResponseWriter, r *http.Request) {
	// 1. データベースを開く
	db, err := sql.Open("sqlite3", "./market.db")
	if err != nil {
		http.Error(w, "DB接続エラー", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 2. 最新の30件を取得するSQL
	rows, err := db.Query("SELECT id, symbol, price, created_at FROM prices ORDER BY id DESC LIMIT 30")
	if err != nil {
		http.Error(w, "データ取得エラー", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// 3. 取得したデータをスライス（配列）にまとめる
	var records []PriceRecord
	for rows.Next() {
		var record PriceRecord
		err := rows.Scan(&record.ID, &record.Symbol, &record.Price, &record.CreatedAt)
		if err != nil {
			log.Println("スキャンエラー:", err)
			continue
		}
		records = append(records, record) // スライスに追加
	}

	// 4. HTMLテンプレートにデータを流し込んで、ブラウザに送信する
	tmpl, err := template.New("dashboard").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, "テンプレートエラー", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, records)
}

// APIとしてJSONを返す関数
func apiHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./market.db")
	if err != nil {
		http.Error(w, "DB接続エラー", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, symbol, price, created_at FROM prices ORDER BY id DESC LIMIT 30")
	if err != nil {
		http.Error(w, "データ取得エラー", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var records []PriceRecord
	for rows.Next() {
		var record PriceRecord
		err := rows.Scan(&record.ID, &record.Symbol, &record.Price, &record.CreatedAt)
		if err != nil {
			log.Println("スキャンエラー:", err)
			continue
		}
		records = append(records, record)
	}

	// 【ここがAPIの心臓部！】
	// ブラウザ（通信相手）に「今から送るのはHTMLじゃなくてJSONですよ」と宣言する
	w.Header().Set("Content-Type", "application/json")

	// スライス（records）をJSON形式に変換して、そのまま送信（Encode）する
	json.NewEncoder(w).Encode(records)
}

func main() {
	// "/"（トップページ）にアクセスが来たら、handler関数を実行するよう設定
	http.HandleFunc("/", handler)

	http.HandleFunc("/api/prices", apiHandler)

	fmt.Println("🌐 Webサーバーを起動しました！")
	fmt.Println("👉 ブラウザで http://localhost:8080 にアクセスしてください。")
	fmt.Println("⚠️ 停止するにはターミナルで [Ctrl] + [C] を押してください。")

	// ポート8080番でサーバーを起動して待機
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("サーバー起動エラー:", err)
	}
}
