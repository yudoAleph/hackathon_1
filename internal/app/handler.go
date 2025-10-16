package app

import (
	"net/http"

	"user-service/configs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	service *Service
}

func NewHandler(cfg configs.Config, db *gorm.DB) *Handler {
	repo := NewUserRepository(db)
	service := NewService(repo, cfg.JWTSecret)
	return &Handler{db: db, service: service}
}

func (h *Handler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func (h *Handler) Get(c *gin.Context) {
	id := c.Param("id")

	if len(id) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
