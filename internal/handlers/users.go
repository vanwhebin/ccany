package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"ccany/ent"
	"ccany/ent/user"
	"ccany/internal/database"
	"ccany/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// UsersHandler user management handler
type UsersHandler struct {
	db             *database.Database
	authMiddleware *middleware.AuthMiddleware
	logger         *logrus.Logger
}

// NewUsersHandler creates user management handler
func NewUsersHandler(db *database.Database, authMiddleware *middleware.AuthMiddleware, logger *logrus.Logger) *UsersHandler {
	return &UsersHandler{
		db:             db,
		authMiddleware: authMiddleware,
		logger:         logger,
	}
}

// UserRequest user request structure
type UserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
}

// UserUpdateRequest user update request structure
type UserUpdateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role" binding:"omitempty,oneof=admin user"`
	IsActive *bool  `json:"is_active"`
}

// PasswordChangeRequest password change request structure
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// LoginRequest login request structure
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// UserResponse user response structure
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	LastLogin time.Time `json:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Login user login - POST /auth/login
func (h *UsersHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	ctx := context.Background()

	// find user
	foundUser, err := h.db.Client.User.Query().
		Where(user.Or(
			user.Username(req.Username),
			user.Email(req.Username),
		)).
		First(ctx)

	if err != nil {
		h.logger.WithError(err).WithField("username", req.Username).Error("User not found")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// verify password
	if !h.verifyPassword(req.Password, foundUser.PasswordHash, foundUser.Salt) {
		h.logger.WithField("username", req.Username).Error("Invalid password")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// check if user is active
	if !foundUser.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Account is disabled",
		})
		return
	}

	// generate JWT token
	token, err := h.authMiddleware.GenerateToken(foundUser.ID, foundUser.Username, foundUser.Role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	// update last login time
	_, err = foundUser.Update().
		SetLastLogin(time.Now()).
		Save(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update last login")
	}

	// set cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    h.toUserResponse(foundUser),
	})
}

// Logout user logout - POST /auth/logout
func (h *UsersHandler) Logout(c *gin.Context) {
	// clear cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetCurrentUser get current user information - GET /auth/me
func (h *UsersHandler) GetCurrentUser(c *gin.Context) {
	userID, _, _, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	user, err := h.db.Client.User.Get(context.Background(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": h.toUserResponse(user),
	})
}

// GetAllUsers get all users - GET /admin/users
func (h *UsersHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	ctx := context.Background()
	query := h.db.Client.User.Query()

	// search filter
	if search != "" {
		query = query.Where(user.Or(
			user.UsernameContains(search),
			user.EmailContains(search),
		))
	}

	// get total count
	total, err := query.Count(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to count users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get users",
		})
		return
	}

	// paginated query
	users, err := query.
		Limit(limit).
		Offset((page - 1) * limit).
		Order(ent.Desc(user.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		h.logger.WithError(err).Error("Failed to get users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get users",
		})
		return
	}

	// convert to response format
	var userResponses []UserResponse
	for _, u := range users {
		userResponses = append(userResponses, h.toUserResponse(u))
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + limit - 1) / limit,
		},
	})
}

// CreateUser create user - POST /admin/users
func (h *UsersHandler) CreateUser(c *gin.Context) {
	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	ctx := context.Background()

	// check if username already exists
	exists, err := h.db.Client.User.Query().
		Where(user.Or(
			user.Username(req.Username),
			user.Email(req.Email),
		)).
		Exist(ctx)

	if err != nil {
		h.logger.WithError(err).Error("Failed to check user existence")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username or email already exists",
		})
		return
	}

	// generate password hash
	salt := h.generateSalt()
	passwordHash, err := h.hashPassword(req.Password, salt)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// create user
	newUser, err := h.db.Client.User.Create().
		SetUsername(req.Username).
		SetEmail(req.Email).
		SetPasswordHash(passwordHash).
		SetSalt(salt).
		SetRole(req.Role).
		SetIsActive(true).
		Save(ctx)

	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	h.logger.WithField("user_id", newUser.ID).Info("User created successfully")

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    h.toUserResponse(newUser),
	})
}

// GetUser get single user - GET /admin/users/:id
func (h *UsersHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	user, err := h.db.Client.User.Get(context.Background(), userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": h.toUserResponse(user),
	})
}

// UpdateUser update user - PUT /admin/users/:id
func (h *UsersHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	ctx := context.Background()

	// check if user exists
	existingUser, err := h.db.Client.User.Get(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// build update query
	updateQuery := existingUser.Update()

	if req.Username != "" && req.Username != existingUser.Username {
		// check if username already exists
		exists, err := h.db.Client.User.Query().
			Where(user.And(
				user.Username(req.Username),
				user.IDNEQ(userID),
			)).
			Exist(ctx)

		if err != nil {
			h.logger.WithError(err).Error("Failed to check username existence")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update user",
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Username already exists",
			})
			return
		}

		updateQuery = updateQuery.SetUsername(req.Username)
	}

	if req.Email != "" && req.Email != existingUser.Email {
		// check if email already exists
		exists, err := h.db.Client.User.Query().
			Where(user.And(
				user.Email(req.Email),
				user.IDNEQ(userID),
			)).
			Exist(ctx)

		if err != nil {
			h.logger.WithError(err).Error("Failed to check email existence")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update user",
			})
			return
		}

		if exists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
			})
			return
		}

		updateQuery = updateQuery.SetEmail(req.Email)
	}

	if req.Role != "" {
		updateQuery = updateQuery.SetRole(req.Role)
	}

	if req.IsActive != nil {
		updateQuery = updateQuery.SetIsActive(*req.IsActive)
	}

	// execute update
	updatedUser, err := updateQuery.Save(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	h.logger.WithField("user_id", userID).Info("User updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    h.toUserResponse(updatedUser),
	})
}

// DeleteUser delete user - DELETE /admin/users/:id
func (h *UsersHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User ID is required",
		})
		return
	}

	// get current user info
	currentUserID, _, _, _ := middleware.GetCurrentUser(c)
	if currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete your own account",
		})
		return
	}

	ctx := context.Background()

	// check if user exists
	existingUser, err := h.db.Client.User.Get(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// delete user
	err = h.db.Client.User.DeleteOne(existingUser).Exec(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	h.logger.WithField("user_id", userID).Info("User deleted successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

// ChangePassword change password - PUT /auth/password
func (h *UsersHandler) ChangePassword(c *gin.Context) {
	userID, _, _, ok := middleware.GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	var req PasswordChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	ctx := context.Background()

	// get user info
	existingUser, err := h.db.Client.User.Get(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("User not found")
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// verify current password
	if !h.verifyPassword(req.CurrentPassword, existingUser.PasswordHash, existingUser.Salt) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Current password is incorrect",
		})
		return
	}

	// generate new password hash
	newSalt := h.generateSalt()
	newPasswordHash, err := h.hashPassword(req.NewPassword, newSalt)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash new password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to change password",
		})
		return
	}

	// update password
	_, err = existingUser.Update().
		SetPasswordHash(newPasswordHash).
		SetSalt(newSalt).
		Save(ctx)

	if err != nil {
		h.logger.WithError(err).Error("Failed to update password")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to change password",
		})
		return
	}

	h.logger.WithField("user_id", userID).Info("Password changed successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// utility methods

// toUserResponse converts to user response format
func (h *UsersHandler) toUserResponse(u *ent.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,
		LastLogin: u.LastLogin,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// generateSalt generates random salt
func (h *UsersHandler) generateSalt() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		// if random generation fails, use timestamp as fallback
		return fmt.Sprintf("salt_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(salt)
}

// hashPassword hashes password
func (h *UsersHandler) hashPassword(password, salt string) (string, error) {
	// combine password and salt
	combined := password + salt
	hash, err := bcrypt.GenerateFromPassword([]byte(combined), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword verifies password
func (h *UsersHandler) verifyPassword(password, hash, salt string) bool {
	combined := password + salt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(combined))
	return err == nil
}
