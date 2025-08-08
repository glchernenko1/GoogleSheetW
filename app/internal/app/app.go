package app

import (
	"GoogleSheetW/internal/cache/localCache"
	"GoogleSheetW/internal/config"
	"GoogleSheetW/internal/controller"
	"GoogleSheetW/internal/logger"
	"GoogleSheetW/internal/services/sheetsControl"
	"context"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type App struct {
	config        *config.Config
	log           *zap.SugaredLogger
	sheetsControl *sheetsControl.SheetsControl
	controller    *controller.SheetsController
}

func New() *App {
	cfg := config.GetConfig()
	log := logger.Get()

	// Инициализация кэша
	cache := localCache.GetInstance()

	// Инициализация сервиса для работы с Google Sheets
	ctx := context.Background()
	sheetsCtrl := sheetsControl.New(ctx, cache, "google.json")

	// Инициализация HTTP контроллера
	httpController := controller.NewSheetsController(sheetsCtrl)

	return &App{
		config:        cfg,
		log:           log,
		sheetsControl: sheetsCtrl,
		controller:    httpController,
	}
}

func (a *App) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Роуты для API
	mux.HandleFunc("/api/sheets/set-data", a.controller.SetSheetData)
	mux.HandleFunc("/api/sheets/", a.handleSheetsRequests) // Универсальный обработчик для DELETE запросов

	// Добавим простой health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status": "ok"}`)); err != nil {
			a.log.Errorw("Ошибка записи ответа health check", "error", err)
		}
	})

	return mux
}

// handleSheetsRequests универсальный обработчик для маршрутизации запросов к таблицам
func (a *App) handleSheetsRequests(w http.ResponseWriter, r *http.Request) {
	// Только DELETE запросы обрабатываем здесь
	if r.Method != http.MethodDelete {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	// Удаляем префикс /api/sheets/ и анализируем путь
	path := r.URL.Path[len("/api/sheets/"):]

	if path == "" {
		http.Error(w, "Неверный URL", http.StatusBadRequest)
		return
	}

	// Проверяем структуру пути
	parts := strings.Split(path, "/")

	if len(parts) == 1 && parts[0] != "" {
		// DELETE /api/sheets/{fiat} - удаление всей таблицы
		a.controller.DeleteSpreadsheet(w, r)
	} else if len(parts) == 3 && parts[1] == "sheet" && parts[2] != "" {
		// DELETE /api/sheets/{fiat}/sheet/{sheetName} - удаление листа
		a.controller.DeleteSheet(w, r)
	} else {
		http.Error(w, "Неверный формат URL", http.StatusBadRequest)
		return
	}
}

func (a *App) Run() error {
	mux := a.setupRoutes()

	addr := fmt.Sprintf(":%d", a.config.App.Port)
	a.log.Infow("Сервер запускается", "address", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	a.log.Info("API endpoints:")
	a.log.Info("POST /api/sheets/set-data - установка данных в таблицу")
	a.log.Info("DELETE /api/sheets/{fiat} - удаление всей таблицы")
	a.log.Info("DELETE /api/sheets/{fiat}/sheet/{sheetName} - удаление листа из таблицы")
	a.log.Info("GET /health - проверка состояния сервиса")

	return server.ListenAndServe()
}
