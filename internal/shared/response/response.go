package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageResponse 分页响应结构
type PageResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Size    int         `json:"size"`
}

// SuccessResponse 成功响应 (原生HTTP)
func SuccessResponse(w http.ResponseWriter, data interface{}) {
	response := Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse 错误响应 (原生HTTP)
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	response := Response{
		Code:    code,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// PagedResponse 分页响应 (原生HTTP)
func PagedResponse(w http.ResponseWriter, data interface{}, total int64, page, size int) {
	response := PageResponse{
		Code:    200,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Success Gin框架成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// Error Gin框架错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "未授权访问"
	}
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "禁止访问"
	}
	Error(c, http.StatusForbidden, message)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "资源不存在"
	}
	Error(c, http.StatusNotFound, message)
}

// InternalError 500错误响应
func InternalError(c *gin.Context, message string) {
	if message == "" {
		message = "服务器内部错误"
	}
	Error(c, http.StatusInternalServerError, message)
}

// Paged Gin框架分页响应
func Paged(c *gin.Context, data interface{}, total int64, page, size int) {
	c.JSON(http.StatusOK, PageResponse{
		Code:    200,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: message,
		Data:    data,
	})
}
