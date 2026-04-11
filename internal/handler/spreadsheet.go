package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"umkm-pos/internal/spreadsheet"
)

type SpreadsheetHandler struct {
	sheetService *spreadsheet.Service
}

type ReadRequest struct {
	SheetName  string `json:"sheet_name"`
	ReadRange  string `json:"read_range"` // Contoh: "A1:Z100"
}

type WriteRequest struct {
	SheetName  string            `json:"sheet_name"`
	WriteRange string            `json:"write_range"` // Contoh: "A1:Z100"
	Values     [][]interface{}   `json:"values"`
}

type AppendRequest struct {
	SheetName  string            `json:"sheet_name"`
	Values     [][]interface{}   `json:"values"`
}

type UpdateCellRequest struct {
	SheetName  string      `json:"sheet_name"`
	CellRange  string      `json:"cell_range"` // Contoh: "A1"
	Value      interface{} `json:"value"`
}

func NewSpreadsheetHandler(sheetService *spreadsheet.Service) *SpreadsheetHandler {
	return &SpreadsheetHandler{
		sheetService: sheetService,
	}
}

// ReadData membaca data dari spreadsheet
func (h *SpreadsheetHandler) ReadData(c *gin.Context) {
	var req ReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.SheetName == "" || req.ReadRange == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "SheetName dan ReadRange harus diisi",
		})
		return
	}

	data, err := h.sheetService.ReadData(c.Request.Context(), req.SheetName, req.ReadRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal membaca data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// WriteData menulis data ke spreadsheet (overwrite)
func (h *SpreadsheetHandler) WriteData(c *gin.Context) {
	var req WriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.SheetName == "" || req.WriteRange == "" || len(req.Values) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "SheetName, WriteRange, dan Values harus diisi",
		})
		return
	}

	err := h.sheetService.WriteData(c.Request.Context(), req.SheetName, req.WriteRange, req.Values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menulis data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// AppendData menambahkan data di baris terakhir
func (h *SpreadsheetHandler) AppendData(c *gin.Context) {
	var req AppendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.SheetName == "" || len(req.Values) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "SheetName dan Values harus diisi",
		})
		return
	}

	err := h.sheetService.AppendData(c.Request.Context(), req.SheetName, req.Values)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menambahkan data: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateCell memperbarui sel spesifik
func (h *SpreadsheetHandler) UpdateCell(c *gin.Context) {
	var req UpdateCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.SheetName == "" || req.CellRange == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "SheetName dan CellRange harus diisi",
		})
		return
	}

	err := h.sheetService.UpdateCell(c.Request.Context(), req.SheetName, req.CellRange, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal memperbarui sel: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}