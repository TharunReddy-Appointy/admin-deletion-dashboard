package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.appointy.com/admin-deletion-dashboard/internal/auth"
	"go.appointy.com/admin-deletion-dashboard/internal/models"
	"go.appointy.com/admin-deletion-dashboard/internal/service"
)

// AccountHandler handles account-related endpoints
type AccountHandler struct {
	accountService *service.AccountService
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// HandleLookup looks up an account by email
func (h *AccountHandler) HandleLookup(c *gin.Context) {
	var req models.AccountLookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Lookup account
	result, err := h.accountService.LookupAccount(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleDelete performs account deletion
func (h *AccountHandler) HandleDelete(c *gin.Context) {
	var req models.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get authenticated user's email from context
	deletedBy, err := auth.GetUserEmailFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Set deleted_by field
	req.DeletedBy = deletedBy

	// Perform deletion
	result, err := h.accountService.DeleteAccount(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleGetAuditLogs retrieves audit logs
func (h *AccountHandler) HandleGetAuditLogs(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get audit logs
	logs, err := h.accountService.GetAuditLogs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}
