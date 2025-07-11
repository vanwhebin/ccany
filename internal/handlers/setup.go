package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"ccany/ent/user"
	"ccany/internal/database"
	"ccany/internal/i18n"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// SetupHandler setup wizard handler
type SetupHandler struct {
	db     *database.Database
	logger *logrus.Logger
}

// NewSetupHandler creates setup wizard handler
func NewSetupHandler(db *database.Database, logger *logrus.Logger) *SetupHandler {
	return &SetupHandler{
		db:     db,
		logger: logger,
	}
}

// CheckSetupRequired checks if setup wizard is required
func (h *SetupHandler) CheckSetupRequired(c *gin.Context) {
	// check if admin user exists
	adminExists, err := h.db.Client.User.Query().Where(user.Role("admin")).Exist(c.Request.Context())
	if err != nil {
		h.logger.WithError(err).Error("Failed to check admin user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.internal_server_error")})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"setup_required": !adminExists,
	})
}

// SetupAdmin creates admin user
func (h *SetupHandler) SetupAdmin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.T(c, "errors.validation_error")})
		return
	}

	ctx := c.Request.Context()

	// check if admin user already exists
	adminExists, err := h.db.Client.User.Query().Where(user.Role("admin")).Exist(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check admin user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.internal_server_error")})
		return
	}

	if adminExists {
		c.JSON(http.StatusConflict, gin.H{"error": i18n.T(c, "errors.admin_exists")})
		return
	}

	// check if username already exists
	userExists, err := h.db.Client.User.Query().Where(user.Username(req.Username)).Exist(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check username")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.internal_server_error")})
		return
	}

	if userExists {
		c.JSON(http.StatusConflict, gin.H{"error": i18n.T(c, "errors.username_exists")})
		return
	}

	// generate salt
	salt, err := generateSalt()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate salt")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.internal_server_error")})
		return
	}

	// generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password+salt), bcrypt.DefaultCost)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.internal_server_error")})
		return
	}

	// generate user ID
	userID, err := generateID()
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate user ID")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.internal_server_error")})
		return
	}

	// create admin user
	admin, err := h.db.Client.User.Create().
		SetID(userID).
		SetUsername(req.Username).
		SetPasswordHash(string(hashedPassword)).
		SetSalt(salt).
		SetRole("admin").
		SetIsActive(true).
		Save(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create admin user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.T(c, "errors.operation_failed")})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  admin.ID,
		"username": admin.Username,
		"role":     admin.Role,
	}).Info("Admin user created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(c, "common.operation_success"),
		"user": gin.H{
			"id":       admin.ID,
			"username": admin.Username,
			"role":     admin.Role,
		},
	})
}

// generateSalt generates random salt
func generateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// generateID generates random user ID
func generateID() (string, error) {
	id := make([]byte, 8)
	_, err := rand.Read(id)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}
