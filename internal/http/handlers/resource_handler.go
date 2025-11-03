package handlers

import (
	"net/http"
	"os"
	"os/user"
	//"strings"

	"github.com/gin-gonic/gin"
)

type LocalResource struct {
	Name        string `json:"name"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	User        string `json:"user"`
	Description string `json:"description"`
}

func GetLocalResource(c *gin.Context) {
	hostname, _ := os.Hostname()
	u, _ := user.Current()
	ip := "127.0.0.1" // can later be replaced with external IP if needed

	resource := LocalResource{
		Name:        hostname,
		Host:        ip,
		Port:        "22",
		User:        u.Username,
		Description: "Local SSH Instance",
	}
	c.JSON(http.StatusOK, gin.H{"resources": []LocalResource{resource}})
}
