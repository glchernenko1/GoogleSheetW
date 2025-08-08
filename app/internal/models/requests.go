package models

// SetSheetDataRequest структура для запроса установки данных в таблицу
type SetSheetDataRequest struct {
	SheetData SheetData `json:"sheet_data"`
}

// APIResponse общая структура ответа API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
