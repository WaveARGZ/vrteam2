package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type Member struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Approved bool   `json:"approved"`
}

var db *sql.DB

func handleCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	handleCORS(w)

	rows, err := db.Query("SELECT id, name, approved FROM members")
	if err != nil {
		http.Error(w, "DB読み込みエラー", 500)
		return
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		var approvedInt int
		if err := rows.Scan(&m.ID, &m.Name, &approvedInt); err == nil {
			m.Approved = approvedInt == 1
			members = append(members, m)
		}
	}

	json.NewEncoder(w).Encode(members)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	handleCORS(w)
	var members []Member
	if err := json.NewDecoder(r.Body).Decode(&members); err != nil {
		http.Error(w, "JSONデコード失敗", 400)
		return
	}

	tx, _ := db.Begin()
	tx.Exec("DELETE FROM members") // 全削除して再保存（シンプルにするため）

	stmt, _ := tx.Prepare("INSERT INTO members(id, name, approved) VALUES(?, ?, ?)")
	defer stmt.Close()

	for _, m := range members {
		approved := 0
		if m.Approved {
			approved = 1
		}
		stmt.Exec(m.ID, m.Name, approved)
	}
	tx.Commit()

	w.WriteHeader(http.StatusOK)
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./members.db")
	if err != nil {
		panic(err)
	}

	// 初回だけテーブル作成（なければ）
	createTable := `
	CREATE TABLE IF NOT EXISTS members (
		id INTEGER PRIMARY KEY,
		name TEXT,
		approved INTEGER
	);`
	_, err = db.Exec(createTable)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/members", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handleCORS(w)
		} else if r.Method == http.MethodGet {
			handleGet(w, r)
		} else if r.Method == http.MethodPost {
			handlePost(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Go server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
