package main

import (
	"fmt"
	"log"

	"teleport_lite/internal/agent"
	"teleport_lite/internal/config"
	"teleport_lite/internal/db"
	httpserver "teleport_lite/internal/http"
	"teleport_lite/internal/models"
	"teleport_lite/internal/seed"
)

func main() {
	cfg := config.Load()

	gdb := db.Connect(cfg.DSN)

	if gdb == nil {
		log.Fatal("❌ Failed to connect to database, aborting startup.")
	}

	db.AutoMigrate(gdb,
		&models.Organization{},
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Resource{},
		&models.AccessRule{},
		&models.AuditLog{},
	)

	if err := seed.FirstSetup(gdb); err != nil {
		log.Fatalf("❌ Seed failed: %v", err)
	}

	go agent.RunLocalAgent(gdb)

	r := httpserver.NewRouter(gdb, cfg.JWTSecret)
	log.Printf("🚀 Server listening on :%s\n", cfg.AppPort)
	r.Run(fmt.Sprintf(":%s", cfg.AppPort))
}
