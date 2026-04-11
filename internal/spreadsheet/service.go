package spreadsheet

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"github.com/google/uuid"

)

type Service struct {
	client *sheets.Service
	spreadsheetID string
}

func NewService(ctx context.Context, spreadsheetID string, serviceAccountPath string) (*Service, error) {
	// Cek apakah file service-account.json ada
	if _, err := os.Stat(serviceAccountPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("service account file tidak ditemukan di path: %s", serviceAccountPath)
	}

	// Buat client sheets dengan service account
	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat sheets service: %w", err)
	}

	return &Service{
		client: sheetsService,
		spreadsheetID: spreadsheetID,
	}, nil
}

// ReadData membaca data dari range tertentu di sheet
func (s *Service) ReadData(ctx context.Context, sheetName string, readRange string) ([][]interface{}, error) {
	rangeFull := fmt.Sprintf("%s!%s", sheetName, readRange)

	resp, err := s.client.Spreadsheets.Values.Get(s.spreadsheetID, rangeFull).Do()
	if err != nil {
		return nil, fmt.Errorf("gagal membaca data dari sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("tidak ada data ditemukan di range: %s", rangeFull)
	}

	return resp.Values, nil
}

// WriteData menulis data ke range tertentu di sheet (overwrite existing data)
func (s *Service) WriteData(ctx context.Context, sheetName string, writeRange string, values [][]interface{}) error {
	rangeFull := fmt.Sprintf("%s!%s", sheetName, writeRange)

	body := &sheets.ValueRange{
		Values: values,
	}

	_, err := s.client.Spreadsheets.Values.Update(s.spreadsheetID, rangeFull, body).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("gagal menulis data ke sheet: %w", err)
	}

	return nil
}

// AppendData menambahkan data di baris terakhir sheet
func (s *Service) AppendData(ctx context.Context, sheetName string, values [][]interface{}) error {
	rangeFull := fmt.Sprintf("%s!A:Z", sheetName)

	body := &sheets.ValueRange{
		Values: values,
	}

	_, err := s.client.Spreadsheets.Values.Append(s.spreadsheetID, rangeFull, body).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("gagal menambahkan data ke sheet: %w", err)
	}

	return nil
}

// UpdateCell memperbarui sel spesifik
func (s *Service) UpdateCell(ctx context.Context, sheetName string, cellRange string, value interface{}) error {
	rangeFull := fmt.Sprintf("%s!%s", sheetName, cellRange)

	body := &sheets.ValueRange{
		Values: [][]interface{}{
			{value},
		},
	}

	_, err := s.client.Spreadsheets.Values.Update(s.spreadsheetID, rangeFull, body).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("gagal memperbarui sel: %w", err)
	}

	return nil
}

// getCellByID mencari baris di sheetName dimana kolom A sama dengan id
// Mengembalikan index baris (dimulai dari 0) dan error jika ada
func (s *Service) getCellByID(ctx context.Context, sheetName string, id interface{}) (int, error) {
	data, err := s.ReadData(ctx, sheetName, "A:C")
	if err != nil {
		// Jika tidak ada data, kembalikan error
		return -1, err
	}

	// Konversi UUID ke string untuk perbandingan
	var targetID string
	switch v := id.(type) {
	case uuid.UUID:
		targetID = v.String()
	case string:
		targetID = v
	default:
		return -1, fmt.Errorf("tipe id tidak didukung: %T", id)
	}

	for i, row := range data {
		if len(row) > 0 {
			// Konversi nilai dari spreadsheet ke string
			var sheetID string
			switch cellVal := row[0].(type) {
			case string:
				sheetID = cellVal
			case []byte:
				sheetID = string(cellVal)
			default:
				// Jika bukan string, coba konversi ke string secara umum
				sheetID = fmt.Sprintf("%v", cellVal)
			}

			if sheetID == targetID {
				return i, nil
			}
		}
	}

	// Tidak ditemukan
	return -1, nil
}

// SaveUserSpreadsheet menyimpan data spreadsheet user ke sheet Users
func (s *Service) SaveUserSpreadsheet(ctx context.Context, sheetName string, googleSubID uuid.UUID, email, spreadsheetID string) error {
	// Cek apakah googleSubID sudah ada
	rowIndex, err := s.getCellByID(ctx, sheetName, googleSubID)
	if err != nil {
		// Jika terjadi error saat membaca data (misal sheet kosong)
		return s.AppendData(ctx, sheetName, [][]interface{}{
			{googleSubID, email, spreadsheetID},
		})
	}

	if rowIndex != -1 {
		// Update baris yang ada
		cellRange := fmt.Sprintf("A%d:C%d", rowIndex+1, rowIndex+1)
		return s.WriteData(ctx, sheetName, cellRange, [][]interface{}{
			{googleSubID, email, spreadsheetID},
		})
	}

	// Jika tidak ada, tambahkan baris baru
	return s.AppendData(ctx, sheetName, [][]interface{}{
		{googleSubID, email, spreadsheetID},
	})
}
