package controllers

import (
	"modress/internal/models"
	"modress/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StoreController struct {
	storeService services.StoreService
}

func NewStoreController(storeService services.StoreService) *StoreController {
	return &StoreController{
		storeService: storeService,
	}
}

func (c *StoreController) CreateStore(ctx *gin.Context) {
	// Obter o ID do usuário do contexto
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.CreateStoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	store, err := c.storeService.CreateStore(ctx.Request.Context(), userIDInt64, &req)
	if err != nil {
		if err.Error() == "user already has a store" {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User already has a store"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, store)
}

func (c *StoreController) GetStore(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	store, err := c.storeService.GetStoreByID(ctx.Request.Context(), id)
	if err != nil {
		if err.Error() == "store not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, store)
}

func (c *StoreController) GetStoreBySlug(ctx *gin.Context) {
	slug := ctx.Param("slug")

	store, err := c.storeService.GetStoreBySlug(ctx.Request.Context(), slug)
	if err != nil {
		if err.Error() == "store not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, store)
}

func (c *StoreController) GetMyStore(ctx *gin.Context) {
	// Obter o ID do usuário do contexto
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	store, err := c.storeService.GetStoreByOwnerID(ctx.Request.Context(), userIDInt64)
	if err != nil {
		if err.Error() == "store not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, store)
}

func (c *StoreController) UpdateStore(ctx *gin.Context) {
	// Obter o ID do usuário do contexto
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	var req models.UpdateStoreRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	store, err := c.storeService.UpdateStore(ctx.Request.Context(), id, userIDInt64, &req)
	if err != nil {
		if err.Error() == "store not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		if err.Error() == "unauthorized: not store owner" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, store)
}

func (c *StoreController) DeleteStore(ctx *gin.Context) {
	// Obter o ID do usuário do contexto
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	err = c.storeService.DeleteStore(ctx.Request.Context(), id, userIDInt64)
	if err != nil {
		if err.Error() == "store not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		if err.Error() == "unauthorized: not store owner" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *StoreController) ListStores(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stores, err := c.storeService.ListStores(ctx.Request.Context(), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stores)
}

func (c *StoreController) ListApprovedStores(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stores, err := c.storeService.ListApprovedStores(ctx.Request.Context(), page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stores)
}

func (c *StoreController) ApproveStore(ctx *gin.Context) {
	// Verificar se o usuário é admin
	userRole, exists := ctx.Get("userRole")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role, ok := userRole.(string)
	if !ok || role != "admin" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: admin access required"})
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
		return
	}

	err = c.storeService.ApproveStore(ctx.Request.Context(), id)
	if err != nil {
		if err.Error() == "store not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}