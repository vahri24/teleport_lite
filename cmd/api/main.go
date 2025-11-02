package main

import (
    "fmt"
    "log"

    "teleport_lite/internal/config"
    "teleport_lite/internal/db"
    "teleport_lite/internal/http"
    "teleport_lite/internal/models"
)

func main() {
    cfg := config.Load()

    gdb := db.Connect(cfg.DSN)
    db.AutoMigrate(gdb,
        &models.Organization{},
        &models.User{},
        &models.Role{},
        &models.Permission{},
        &models.Resource{},
        &models.AccessRule{},
        &models.AuditLog{},
    )

    r := httpserver.NewRouter(gdb, cfg.JWTSecret)
    log.Printf("🚀 Server listening on :%s\n", cfg.AppPort)
    r.Run(fmt.Sprintf(":%s", cfg.AppPort))
}
