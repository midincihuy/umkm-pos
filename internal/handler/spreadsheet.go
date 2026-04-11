package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"umkm-pos/pkg/jwt"
	"umkm-pos/internal/spreadsheet"
	"umkm-pos/pkg/response"

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

type SaveSpreadsheetRequest struct {
	SpreadsheetID string `json:"spreadsheet_id"`
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
		response.Error(c, http.StatusBadRequest, "Invalid request body: ", err.Error())
		return
	}

	if req.SheetName == "" || req.ReadRange == "" {
		response.Error(c, http.StatusBadRequest, "SheetName dan ReadRange harus diisi", nil)
		return
	}

	data, err := h.sheetService.ReadData(c.Request.Context(), req.SheetName, req.ReadRange)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal membaca data: ", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "OK", data)
}

// WriteData menulis data ke spreadsheet (overwrite)
func (h *SpreadsheetHandler) WriteData(c *gin.Context) {
	var req WriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: ", err.Error())
		return
	}

	if req.SheetName == "" || req.WriteRange == "" || len(req.Values) == 0 {
		response.Error(c, http.StatusBadRequest, "SheetName, WriteRange, dan Values harus diisi", nil)
		return
	}

	err := h.sheetService.WriteData(c.Request.Context(), req.SheetName, req.WriteRange, req.Values)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menulis data: ", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Data Tersimpan", req)
}

// AppendData menambahkan data di baris terakhir
func (h *SpreadsheetHandler) AppendData(c *gin.Context) {
	var req AppendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: ", err.Error())
		return
	}

	if req.SheetName == "" || len(req.Values) == 0 {
		response.Error(c, http.StatusBadRequest, "SheetName dan Values harus diisi", nil)
		return
	}

	err := h.sheetService.AppendData(c.Request.Context(), req.SheetName, req.Values)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menambahkan data: ", err.Error())
		return
	}

	response.Success(c, http.StatusCreated, "Data berhasil ditambahkan", req)
}

// UpdateCell memperbarui sel spesifik
func (h *SpreadsheetHandler) UpdateCell(c *gin.Context) {
	var req UpdateCellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: ", err.Error())
		return
	}

	if req.SheetName == "" || req.CellRange == "" {
		response.Error(c, http.StatusBadRequest, "SheetName dan CellRange harus diisi", nil)
		return
	}

	err := h.sheetService.UpdateCell(c.Request.Context(), req.SheetName, req.CellRange, req.Value)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal memperbarui cell: ", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Data berhasil diperbarui", req)
}

// SaveUserSpreadsheet menyimpan data spreadsheet user ke sheet Users
func (h *SpreadsheetHandler) SaveUserSpreadsheet(c *gin.Context) {
	var req SaveSpreadsheetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body: ", err.Error())
		return
	}

	if req.SpreadsheetID == "" {
		response.Error(c, http.StatusBadRequest, "SpreadsheetID harus diisi", nil)
		return
	}

	// Dapatkan user info dari context yang diset middleware auth
	userID := getUserID(c)


	// Dapatkan claims user untuk mendapatkan email
	claims, err := getUserClaims(c)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Tidak dapat mendapatkan data user: ", err.Error())
		return
	}

	var email string
	switch cl := claims.(type) {
	case *jwt.GoogleClaims:
		email = cl.Email
	case *jwt.GoogleTokenInfo:
		email = cl.Email
	default:
		email = ""
	}

	err = h.sheetService.SaveUserSpreadsheet(c.Request.Context(), "Users", userID, email, req.SpreadsheetID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Gagal menyimpan spreadsheet id: ", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "Spreadsheet ID berhasil disimpan", req)
}