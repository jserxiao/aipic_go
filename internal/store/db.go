package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动，无需 CGO
)

var db *sql.DB

// Init 初始化数据库
func Init(dbPath string) error {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db dir failed: %w", err)
	}

	d, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("open db failed: %w", err)
	}

	if err := d.Ping(); err != nil {
		return fmt.Errorf("ping db failed: %w", err)
	}

	db = d

	// 启用外键
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("enable foreign keys failed: %w", err)
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *sql.DB {
	return db
}

// Close 关闭数据库
func Close() {
	if db != nil {
		db.Close()
	}
}

// Migrate 执行数据库迁移
func Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT UNIQUE NOT NULL,
		random_id TEXT NOT NULL,
		avatar TEXT DEFAULT '',
		is_vip BOOLEAN DEFAULT FALSE,
		vip_level TEXT DEFAULT 'normal',
		vip_expire_at TEXT,
		credits INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS generate_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		prompt TEXT NOT NULL,
		ratio TEXT NOT NULL,
		resolution TEXT NOT NULL,
		model TEXT NOT NULL,
		custom_width INTEGER DEFAULT 0,
		custom_height INTEGER DEFAULT 0,
		reference_image TEXT DEFAULT '',
		status TEXT DEFAULT 'queued',
		progress INTEGER DEFAULT 0,
		result_url TEXT DEFAULT '',
		error_message TEXT DEFAULT '',
		credits_cost INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS generate_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		record_id TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		prompt TEXT NOT NULL,
		ratio TEXT NOT NULL,
		resolution TEXT NOT NULL,
		image_url TEXT DEFAULT '',
		status TEXT DEFAULT 'pending',
		credits_cost INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS payment_orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		order_id TEXT UNIQUE NOT NULL,
		user_id TEXT NOT NULL,
		order_type TEXT NOT NULL,
		amount INTEGER NOT NULL,
		credits INTEGER DEFAULT 0,
		status TEXT DEFAULT 'pending',
		plan_type TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_generate_tasks_user_id ON generate_tasks(user_id);
	CREATE INDEX IF NOT EXISTS idx_generate_records_user_id ON generate_records(user_id);
	CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id ON payment_orders(user_id);
	CREATE INDEX IF NOT EXISTS idx_generate_records_created_at ON generate_records(created_at DESC);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate failed: %w", err)
	}

	return nil
}
