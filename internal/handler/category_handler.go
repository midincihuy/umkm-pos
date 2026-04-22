package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/dto"
	"umkm-pos/internal/service"
	"umkm-pos/pkg/response"
)

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) List(c *gin.Context) {
	userID := getUserID(c)
	catType := c.Query("type")
	resp, err := h.svc.List(userID, nilStr(catType), false)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", resp)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	userID := getUserID(c)
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	cat, err := h.svc.Create(userID, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Kategori berhasil dibuat", cat)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Input tidak valid", err.Error())
		return
	}
	cat, err := h.svc.Update(id, &req)
	if err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Kategori diperbarui", cat)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id); err != nil {
		handleErr(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Kategori dihapus", nil)
}
