package db

import (
    "log"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
    gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("❌ Failed to connect to database: %v", err)
    }

    sqlDB, _ := gdb.DB()
    if err := sqlDB.Ping(); err != nil {
        log.Fatalf("❌ Database ping failed: %v", err)
    }

    log.Println("✅ Database connected successfully")
    return gdb
}
