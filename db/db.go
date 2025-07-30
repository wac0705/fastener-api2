package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL 驅動註冊
)

var DB *sql.DB // 全局資料庫連接實例

// InitDB 初始化資料庫連接
func InitDB(connStr string) {
	if connStr == "" {
		log.Fatal("Database connection string is empty. Please set DATABASE_URL in environment or .env file.")
	}

	var err error
	DB, err = sql.Open("postgres", connStr) // 打開資料庫連接
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	// 設定連接池參數
	DB.SetMaxOpenConns(25)                  // 最大打開連接數
	DB.SetMaxIdleConns(25)                  // 最大閒置連接數
	DB.SetConnMaxLifetime(5 * time.Minute)  // 連接最長生命週期 (防止長期空閒連接被資料庫斷開)
	DB.SetConnMaxIdleTime(1 * time.Minute)  // 連接在被連接池回收前可以閒置的最大時間

	// 測試連接
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	fmt.Println("Database connected successfully!")
}
