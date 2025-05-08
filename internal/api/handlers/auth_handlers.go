package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"coupon-system/internal/auth"
)

// AuthHandlers defines the handlers for authentication-related API endpoints.
type AuthHandlers struct {
	// Add any dependencies if needed (e.g., a user service to verify user existence)
}

// NewAuthHandlers creates a new AuthHandlers instance.
func NewAuthHandlers() *AuthHandlers {
	return &AuthHandlers{}
}

// GenerateTokenRequest represents the request body for generating a JWT.
type GenerateTokenRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=admin user"`
}

// GenerateTokenResponse represents the response body containing the JWT.
type GenerateTokenResponse struct {
	Token string `json:"token"`
}

// GenerateTokenHandler handles the generation of a new JWT.
// GenerateTokenHandler godoc
//
//	@Summary		Generate a JWT
//	@Description	Generates a JSON Web Token for a given user ID and role.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		GenerateTokenRequest	true	"User ID and Role"
//	@Success		200		{object}	GenerateTokenResponse	"JWT generated successfully"
//	@Failure		400		{object}	models.ErrorResponse	"Bad request"
//	@Failure		500		{object}	models.ErrorResponse	"Internal server error"
//	@Router			/generate-tokens [post]
func (h *AuthHandlers) GenerateTokenHandler(c *gin.Context) {
	var req GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real application, you would verify the user's credentials here
	// and retrieve their actual role from your database.
	// For this example, we're directly using the provided user_id and role.

	token, err := auth.GenerateJWT(req.UserID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenerateTokenResponse{Token: token})
}
