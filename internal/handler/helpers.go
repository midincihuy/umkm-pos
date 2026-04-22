package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"umkm-pos/pkg/apperror"
	"umkm-pos/pkg/response"
)

func getUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get("userID")
	return id.(uuid.UUID)
}

func getUserClaims(c *gin.Context) (interface{}, error) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, errors.New("user claims tidak ditemukan di context")
	}
	return claims, nil
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID tidak valid", nil)
		return uuid.Nil, false
	}
	return id, true
}

func intQuery(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func handleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		response.Error(c, http.StatusNotFound, err.Error(), nil)
	case errors.Is(err, apperror.ErrUnauthorized):
		response.Error(c, http.StatusUnauthorized, err.Error(), nil)
	case errors.Is(err, apperror.ErrForbidden):
		response.Error(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, apperror.ErrInsufficientBalance):
		response.Error(c, http.StatusUnprocessableEntity, err.Error(), nil)
	case errors.Is(err, apperror.ErrConflict):
		response.Error(c, http.StatusConflict, err.Error(), nil)
	case errors.Is(err, apperror.ErrSystemCategory):
		response.Error(c, http.StatusForbidden, err.Error(), nil)
	case errors.Is(err, apperror.ErrDebtNotActive):
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, apperror.ErrPaymentExceedsDebt):
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	case errors.Is(err, apperror.ErrBadRequest):
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
	default:
		response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", nil)
	}
}

func nilStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
