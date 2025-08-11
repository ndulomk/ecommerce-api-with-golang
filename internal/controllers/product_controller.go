package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"modress/internal/models"
	"modress/internal/services"

	"github.com/gin-gonic/gin"
)

// ProductController handles HTTP requests related to products.
type ProductController struct {
    productService services.ProductService
    storeService   services.StoreService
}

// NewProductController creates a new ProductController instance.
func NewProductController(productService services.ProductService, storeService services.StoreService) *ProductController {
    return &ProductController{
        productService: productService,
        storeService:   storeService,
    }
}

// getUserAndStore retrieves the user ID and store from the context, performing necessary validations.
func (c *ProductController) getUserAndStore(ctx *gin.Context) (int64, *models.StoreResponse, error) {
    userID, exists := ctx.Get("userID")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user ID not found in context"})
        return 0, nil, nil
    }

    userIDInt64, ok := userID.(int64)
    if !ok {
        ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
        return 0, nil, nil
    }

    store, err := c.storeService.GetStoreByOwnerID(ctx.Request.Context(), userIDInt64)
    if err != nil {
        if err.Error() == "store not found" {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "User does not have a store"})
        } else {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve store: " + err.Error()})
        }
        return 0, nil, err
    }

    return userIDInt64, store, nil
}

// parsePaginationParams validates and parses pagination parameters, setting defaults if needed.
func parsePaginationParams(pageStr, limitStr string) (int, int) {
    page, _ := strconv.Atoi(pageStr)
    if page < 1 {
        page = 1
    }

    limit, _ := strconv.Atoi(limitStr)
    if limit < 1 || limit > 100 {
        limit = 20
    }

    return page, limit
}

// CreateProduct handles the creation of a new product.
func (c *ProductController) CreateProduct(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    _, store, err := c.getUserAndStore(ctx)
    if err != nil {
        return 
    }

    var req models.CreateProductRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
        return
    }

    product, err := c.productService.CreateProduct(ctx.Request.Context(), store.ID, &req)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusCreated, product)
}

// GetProduct retrieves a product by its ID.
func (c *ProductController) GetProduct(ctx *gin.Context) {
    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    product, err := c.productService.GetProductByID(ctx.Request.Context(), id)
    if err != nil {
        if err.Error() == "product not found" {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        } else {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product: " + err.Error()})
        }
        return
    }

    ctx.JSON(http.StatusOK, product)
}

// GetProductsByStore retrieves products for a specific store with pagination.
func (c *ProductController) GetProductsByStore(ctx *gin.Context) {
    storeIDStr := ctx.Param("storeId")
    storeID, err := strconv.ParseInt(storeIDStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid store ID"})
        return
    }

    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    page, limit := parsePaginationParams(ctx.Query("page"), ctx.Query("limit"))

    products, err := c.productService.GetProductsByStoreID(ctx.Request.Context(), storeID, page, limit)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, products)
}

// GetProductsByCategory retrieves products for a specific category with pagination.
func (c *ProductController) GetProductsByCategory(ctx *gin.Context) {
    category := ctx.Param("category")
    if category == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Category parameter is required"})
        return
    }

    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    page, limit := parsePaginationParams(ctx.Query("page"), ctx.Query("limit"))

    products, err := c.productService.GetProductsByCategory(ctx.Request.Context(), category, page, limit)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products by category: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, products)
}

// UpdateProduct updates an existing product.
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    _, store, err := c.getUserAndStore(ctx)
    if err != nil {
        return 
    }

    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    var req models.UpdateProductRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
        return
    }

    product, err := c.productService.UpdateProduct(ctx.Request.Context(), id, store.ID, &req)
    if err != nil {
        if err.Error() == "product not found" {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        } else if err.Error() == "unauthorized: product does not belong to store" {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: product does not belong to store"})
        } else {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        }
        return
    }

    ctx.JSON(http.StatusOK, product)
}

// DeleteProduct deletes a product.
func (c *ProductController) DeleteProduct(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    _, store, err := c.getUserAndStore(ctx)
    if err != nil {
        return
    }

    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    err = c.productService.DeleteProduct(ctx.Request.Context(), id, store.ID)
    if err != nil {
        if err.Error() == "product not found" {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        } else if err.Error() == "unauthorized: product does not belong to store" {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: product does not belong to store"})
        } else {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product: " + err.Error()})
        }
        return
    }

    ctx.Status(http.StatusNoContent)
}

// ListProducts lists all products with pagination.
func (c *ProductController) ListProducts(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    page, limit := parsePaginationParams(ctx.Query("page"), ctx.Query("limit"))

    products, err := c.productService.ListProducts(ctx.Request.Context(), page, limit)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, products)
}

// SearchProducts searches products by query with pagination.
func (c *ProductController) SearchProducts(ctx *gin.Context) {
    query := ctx.Query("q")
    if query == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
        return
    }

    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    page, limit := parsePaginationParams(ctx.Query("page"), ctx.Query("limit"))

    products, err := c.productService.SearchProducts(ctx.Request.Context(), query, page, limit)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products: " + err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, products)
}

// UpdateQuantity updates the quantity of a product.
func (c *ProductController) UpdateQuantity(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    _, store, err := c.getUserAndStore(ctx)
    if err != nil {
        return 
    }

    idStr := ctx.Param("id")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    var req struct {
        Quantity int `json:"quantity" validate:"min=0"`
    }

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
        return
    }

    err = c.productService.UpdateProductQuantity(ctx.Request.Context(), id, store.ID, req.Quantity)
    if err != nil {
        if err.Error() == "product not found" {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        } else if err.Error() == "unauthorized: product does not belong to store" {
            ctx.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: product does not belong to store"})
        } else if err.Error() == "quantity cannot be negative" {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "Quantity cannot be negative"})
        } else {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        }
        return
    }

    ctx.Status(http.StatusNoContent)
}


func (c *ProductController) AddProductImage(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    // Get user and store
    _, store, err := c.getUserAndStore(ctx)
    if err != nil {
        return 
    }

    idStr := ctx.Param("id")
    productID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    // Get form file
    file, header, err := ctx.Request.FormFile("image")
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
        return
    }
    defer file.Close()

    // Validate file type and size
    if !isValidImageType(header.Filename) {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only JPEG, PNG, and GIF are allowed"})
        return
    }
    if header.Size > 5*1024*1024 { 
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds 5MB limit"})
        return
    }

    // Generate unique filename
    filename := fmt.Sprintf("%d-%d-%s", productID, time.Now().UnixNano(), sanitizeFilename(header.Filename))
    filePath := filepath.Join("uploads/images", filename)

    // Create uploads directory if it doesn't exist
    if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
        return
    }

    // Save file to disk
    out, err := os.Create(filePath)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
        return
    }
    defer out.Close()

    if _, err := io.Copy(out, file); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
        return
    }

    // Handle alt_text - convert string to *string if not empty
    var altText *string
    if altTextValue := ctx.PostForm("alt_text"); altTextValue != "" {
        altText = &altTextValue
    }

    // Create ProductImage
    image := &models.ProductImage{
        URL:       fmt.Sprintf("/images/%s", filename),
        AltText:   altText,
        Position:  1,                       
        IsPrimary: ctx.PostForm("is_primary") == "true", 
    }

    // Save image metadata
    if err := c.productService.CreateProductImage(ctx.Request.Context(), productID, store.ID, image); err != nil {
        os.Remove(filePath)
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusCreated, image)
}

// GetProductImages retrieves all images for a product.
func (c *ProductController) GetProductImages(ctx *gin.Context) {
    // Check for context cancellation
    if ctx.Request.Context().Err() != nil {
        ctx.JSON(http.StatusRequestTimeout, gin.H{"error": "Request cancelled or timed out"})
        return
    }

    idStr := ctx.Param("id")
    productID, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }

    images, err := c.productService.GetProductImages(ctx.Request.Context(), productID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, images)
}

func isValidImageType(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif"
}

// sanitizeFilename removes unsafe characters from filenames.
func sanitizeFilename(filename string) string {
    ext := filepath.Ext(filename)
    name := strings.TrimSuffix(filename, ext)
    name = strings.ReplaceAll(name, " ", "_")
    name = regexp.MustCompile(`[^a-zA-Z0-9_-]`).ReplaceAllString(name, "")
    return name + ext
}