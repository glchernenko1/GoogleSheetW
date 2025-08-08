package controller

import (
	"GoogleSheetW/internal/logger"
	"GoogleSheetW/internal/models"
	"GoogleSheetW/internal/services/sheetsControl"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
)

type SheetsController struct {
	sheetsControl *sheetsControl.SheetsControl
	log           *zap.SugaredLogger
}

func NewSheetsController(sheetsControl *sheetsControl.SheetsControl) *SheetsController {
	return &SheetsController{
		sheetsControl: sheetsControl,
		log:           logger.Get(),
	}
}

// SetSheetData обрабатывает запрос на установку данных в таблицу
func (sc *SheetsController) SetSheetData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sc.sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	var req models.SetSheetDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sc.log.Errorw("Ошибка декодирования JSON", "error", err)
		sc.sendErrorResponse(w, http.StatusBadRequest, "Неверный формат JSON")
		return
	}

	if err := sc.sheetsControl.SetSheetData(req.SheetData); err != nil {
		sc.log.Errorw("Ошибка установки данных в таблицу", "error", err)
		sc.sendErrorResponse(w, http.StatusInternalServerError, "Ошибка обработки данных")
		return
	}

	sc.sendSuccessResponse(w, "Данные успешно установлены", nil)
}

// DeleteSheet обрабатывает запрос на удаление листа из таблицы
// URL: DELETE /api/sheets/{fiat}/sheet/{sheetName}
func (sc *SheetsController) DeleteSheet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sc.sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	// Извлекаем параметры из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/sheets/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[1] != "sheet" {
		sc.sendErrorResponse(w, http.StatusBadRequest, "Неверный формат URL. Ожидается: /api/sheets/{fiat}/sheet/{sheetName}")
		return
	}

	fiat, err := url.QueryUnescape(parts[0])
	if err != nil {
		sc.sendErrorResponse(w, http.StatusBadRequest, "Неверный параметр fiat")
		return
	}

	sheetName, err := url.QueryUnescape(parts[2])
	if err != nil {
		sc.sendErrorResponse(w, http.StatusBadRequest, "Неверный параметр sheetName")
		return
	}

	if err := sc.sheetsControl.DeleteSheet(fiat, sheetName); err != nil {
		sc.log.Errorw("Ошибка удаления листа", "error", err, "fiat", fiat, "sheetName", sheetName)
		sc.sendErrorResponse(w, http.StatusInternalServerError, "Ошибка удаления листа")
		return
	}

	sc.sendSuccessResponse(w, "Лист успешно удален", map[string]interface{}{
		"fiat":       fiat,
		"sheet_name": sheetName,
	})
}

// DeleteSpreadsheet обрабатывает запрос на удаление всей таблицы
// URL: DELETE /api/sheets/{fiat}
func (sc *SheetsController) DeleteSpreadsheet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sc.sendErrorResponse(w, http.StatusMethodNotAllowed, "Метод не разрешен")
		return
	}

	// Извлекаем fiat из URL
	path := strings.TrimPrefix(r.URL.Path, "/api/sheets/")
	if path == "" || strings.Contains(path, "/") {
		sc.sendErrorResponse(w, http.StatusBadRequest, "Неверный формат URL. Ожидается: /api/sheets/{fiat}")
		return
	}

	fiat, err := url.QueryUnescape(path)
	if err != nil {
		sc.sendErrorResponse(w, http.StatusBadRequest, "Неверный параметр fiat")
		return
	}

	if err := sc.sheetsControl.DeleteSpreadsheet(fiat); err != nil {
		sc.log.Errorw("Ошибка удаления таблицы", "error", err, "fiat", fiat)
		sc.sendErrorResponse(w, http.StatusInternalServerError, "Ошибка удаления таблицы")
		return
	}

	sc.sendSuccessResponse(w, "Таблица успешно удалена", map[string]interface{}{
		"fiat": fiat,
	})
}

// sendSuccessResponse отправляет успешн��й ответ
func (sc *SheetsController) sendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	sc.sendJSONResponse(w, http.StatusOK, response)
}

// sendErrorResponse отправляет ответ с ошибкой
func (sc *SheetsController) sendErrorResponse(w http.ResponseWriter, statusCode int, errorMsg string) {
	response := models.APIResponse{
		Success: false,
		Message: "Ошибка",
		Error:   errorMsg,
	}
	sc.sendJSONResponse(w, statusCode, response)
}

// sendJSONResponse отправляет JSON ответ
func (sc *SheetsController) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		sc.log.Errorw("Ошибка кодирования JSON ответа", "error", err)
	}
}
