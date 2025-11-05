package handlers

import (
    "net/http"
    //"teleport_lite/internal/auth" 
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "teleport_lite/internal/models"
)

func ListAudit(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var Audit []models.AuditLog
        if err := db.Find(&Audit).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"Audit": Audit})
    }
}