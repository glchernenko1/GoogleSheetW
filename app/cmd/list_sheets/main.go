package main

import (
	"GoogleSheetW/internal/services/googleAPI"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("=== Управление Google Sheets таблицами ===")

	// Инициализация контекста
	ctx := context.Background()

	// Получение учетных данных
	cred, err := googleAPI.GetCredentials(ctx, "../app/google.json")
	if err != nil {
		log.Fatalf("❌ Ошибка получения учетных данных: %v", err)
	}

	// Создание Drive сервиса
	driveSrv, err := googleAPI.GetDriveService(ctx, cred)
	if err != nil {
		log.Fatalf("❌ Ошибка создания Drive сервиса: %v", err)
	}

	// Создание Sheets сервиса
	sheetsSrv, err := googleAPI.GetSheetsService(ctx, cred)
	if err != nil {
		log.Fatalf("❌ Ошибка создания Sheets сервиса: %v", err)
	}

	for {
		// Получение списка всех таблиц
		sheetIDMap, err := googleAPI.GetAllSheetIDByName(driveSrv)
		if err != nil {
			log.Fatalf("❌ Ошибка получения списка таблиц: %v", err)
		}

		if len(sheetIDMap) == 0 {
			fmt.Println("📋 Таблиц не найдено")
			return
		}

		fmt.Printf("\n📊 Найдено таблиц: %d\n\n", len(sheetIDMap))

		// Создаем срезы для удобства работы с индексами
		var names []string
		var ids []string

		fmt.Println("№  | Название таблицы")
		fmt.Println("---|------------------")

		i := 1
		for name, id := range sheetIDMap {
			fmt.Printf("%2d | %s\n", i, name)
			names = append(names, name)
			ids = append(ids, id)
			i++
		}

		fmt.Println("\nВыберите действие:")
		fmt.Println("1. Показать подробную информацию о таблице")
		fmt.Println("2. Удалить таблицу")
		fmt.Println("3. Обновить список")
		fmt.Println("0. Выход")
		fmt.Print("\nВведите номер действия: ")

		reader := bufio.NewReader(os.Stdin)
		actionStr, _ := reader.ReadString('\n')
		actionStr = strings.TrimSpace(actionStr)

		action, err := strconv.Atoi(actionStr)
		if err != nil {
			fmt.Println("❌ Неверный ввод. Введите число.")
			continue
		}

		switch action {
		case 0:
			fmt.Println("👋 До свидания!")
			return

		case 1:
			showTableDetails(names, ids, sheetsSrv)

		case 2:
			deleteTable(names, ids, driveSrv)

		case 3:
			fmt.Println("🔄 Обновляю список таблиц...")
			continue

		default:
			fmt.Println("❌ Неверный выбор. Попробуйте снова.")
		}
	}
}

func showTableDetails(names []string, ids []string, sheetsSrv *sheets.Service) {
	if len(names) == 0 {
		fmt.Println("❌ Нет доступных таблиц")
		return
	}

	fmt.Print("\nВведите номер таблицы для просмотра деталей (1-" + strconv.Itoa(len(names)) + "): ")
	reader := bufio.NewReader(os.Stdin)
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)

	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(names) {
		fmt.Println("❌ Неверный номер таблицы")
		return
	}

	tableName := names[choice-1]
	tableID := ids[choice-1]

	fmt.Printf("\n📋 Детали таблицы: %s\n", tableName)
	fmt.Printf("🆔 ID: %s\n", tableID)
	fmt.Printf("🔗 Ссылка: https://docs.google.com/spreadsheets/d/%s/edit\n", tableID)

	// Получаем список листов в таблице
	spreadsheet, err := sheetsSrv.Spreadsheets.Get(tableID).Do()
	if err != nil {
		fmt.Printf("⚠️ Не удалось получить информацию о листах: %v\n", err)
		return
	}

	fmt.Printf("📄 Количество листов: %d\n", len(spreadsheet.Sheets))
	if len(spreadsheet.Sheets) > 0 {
		fmt.Println("📄 Листы в таблице:")
		for i, sheet := range spreadsheet.Sheets {
			fmt.Printf("   %d. %s\n", i+1, sheet.Properties.Title)
		}
	}

	fmt.Println("\nНажмите Enter для продолжения...")
	reader.ReadString('\n')
}

func deleteTable(names []string, ids []string, driveSrv *drive.Service) {
	if len(names) == 0 {
		fmt.Println("❌ Нет доступных таблиц для удаления")
		return
	}

	fmt.Print("\nВведите номер таблицы для удаления (1-" + strconv.Itoa(len(names)) + "): ")
	reader := bufio.NewReader(os.Stdin)
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)

	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(names) {
		fmt.Println("❌ Неверный номер таблицы")
		return
	}

	tableName := names[choice-1]
	tableID := ids[choice-1]

	fmt.Printf("\n⚠️ ВНИМАНИЕ! Вы действительно хотите удалить таблицу?\n")
	fmt.Printf("📋 Название: %s\n", tableName)
	fmt.Printf("🆔 ID: %s\n", tableID)
	fmt.Print("❗ Это действие НЕОБРАТИМО! Введите 'yes' для подтверждения: ")

	confirmStr, _ := reader.ReadString('\n')
	confirmStr = strings.TrimSpace(strings.ToLower(confirmStr))

	if confirmStr != "yes" {
		fmt.Printf("❌ Удаление отменено (введено: '%s')\n", confirmStr)
		fmt.Println("\nНажмите Enter для продолжения...")
		reader.ReadString('\n')
		return
	}

	fmt.Println("🗑️ Удаляю таблицу...")

	err = googleAPI.DeleteSpreadsheetByID(driveSrv, tableID)
	if err != nil {
		fmt.Printf("❌ Ошибка удаления таблицы: %v\n", err)
		fmt.Println("\nНажмите Enter для продолжения...")
		reader.ReadString('\n')
		return
	}

	fmt.Printf("✅ Таблица '%s' успешно удалена!\n", tableName)

	fmt.Println("\nНажмите Enter для продолжения...")
	reader.ReadString('\n')
}
