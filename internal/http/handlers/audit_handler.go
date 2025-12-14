package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"teleport_lite/internal/auth"
	"teleport_lite/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListAudit(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsI, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		cl := claimsI.(*auth.Claims)

		limit := 20
		if limitStr := c.Query("limit"); limitStr != "" {
			if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}

		var afterID int64
		if cursorStr := c.Query("after_id"); cursorStr != "" {
			if parsed, err := strconv.ParseInt(cursorStr, 10, 64); err == nil && parsed > 0 {
				afterID = parsed
			}
		}

		search := strings.TrimSpace(c.Query("q"))

		query := db.Model(&models.AuditLog{}).Where("org_id = ?", cl.OrgID).Order("id DESC")
		if afterID > 0 {
			query = query.Where("id < ?", afterID)
		}
		if search != "" {
			like := "%" + search + "%"
			query = query.Where("(initiator_name LIKE ? OR action LIKE ? OR resource_type LIKE ? OR ip LIKE ?)",
				like, like, like, like)
		}

		var logs []models.AuditLog
		if err := query.Limit(limit + 1).Find(&logs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var nextCursor *int64
		if len(logs) > limit {
			next := logs[limit].ID
			logs = logs[:limit]
			nextCursor = &next
		}

		c.JSON(http.StatusOK, gin.H{
			"logs":        logs,
			"next_cursor": nextCursor,
		})
	}
}
