package handler

import (
	"net/http"
	"fmt"

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

	// spreadsheetID boleh kosong, untuk save ID dan email sebelum pick sheet, dan untuk remove sheet
	
	// if req.SpreadsheetID == "" {
	// 	response.Error(c, http.StatusBadRequest, "SpreadsheetID harus diisi", nil)
	// 	return
	// }

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

	response.Success(c, http.StatusOK, "Spreadsheet ID berhasil disimpan", map[string]string{"spreadsheet_id": req.SpreadsheetID, "url" : fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", req.SpreadsheetID)})
}
// GetUserSpreadsheetID mendapatkan spreadsheet ID user yang tersimpan
func (h *SpreadsheetHandler) GetUserSpreadsheetID(c *gin.Context) {
	// Dapatkan user info dari context yang diset middleware auth
	userID := getUserID(c)

	// getUserID already returns uuid.UUID type, so just convert directly to string
	userIDStr := userID.String()

	// Baca data spreadsheet users
	data, err := h.sheetService.ReadData(c.Request.Context(), "Users", "A:C")
	if err != nil {
		// Jika sheet kosong atau tidak ada data, kembalikan empty string
		if err.Error() == "tidak ada data ditemukan di range: Users!A:C" {
			response.Error(c, http.StatusNotFound, "Not Found", err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "Gagal membaca data spreadsheet: ", err.Error())
		return
	}

	// Cari spreadsheet ID untuk user ini
	var spreadsheetID string
	for _, row := range data {
		if len(row) >= 3 {
			// Konversi nilai dari spreadsheet ke string
			var currentUserID string
			switch cellVal := row[0].(type) {
			case string:
				currentUserID = cellVal
			case []byte:
				currentUserID = string(cellVal)
			default:
				currentUserID = fmt.Sprintf("%v", cellVal)
			}

			if currentUserID == userIDStr {
				// Konversi nilai spreadsheet ID ke string
				switch val := row[2].(type) {
				case string:
					spreadsheetID = val
				case []byte:
					spreadsheetID = string(val)
				default:
					spreadsheetID = fmt.Sprintf("%v", val)
				}
				break
			}
		}
	}
	if spreadsheetID == ""{
		response.Error(c, http.StatusOK, "Spreadsheet ID Not Found", "")
		return
	}
	response.Success(c, http.StatusOK, "OK", map[string]string{"spreadsheet_id": spreadsheetID, "url" : fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s", spreadsheetID)})
}
