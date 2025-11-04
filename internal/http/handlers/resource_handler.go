package handlers

import (
	"net/http"
	//"os"
	//"os/user"
	//"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
    "teleport_lite/internal/models"
)


func ListResources(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var resource []models.Resource
        if err := db.Find(&resource).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"resources": resource})
    }
}
