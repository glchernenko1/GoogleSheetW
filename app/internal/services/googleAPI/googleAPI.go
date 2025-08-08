package googleAPI

import (
	"context"
	"fmt"
	"google.golang.org/api/option"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
	"io"
	"os"
)

func SheetExists(srv *sheets.Service, spreadsheetID, sheetName string) (bool, error) {

	spreadsheet, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return false, fmt.Errorf("Ошибка при получении таблицы: %v", err)
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return true, nil
		}
	}
	return false, nil
}

func SheetIDByName(srv *sheets.Service, spreadsheetID, name string) (int64, error) {
	resp, err := srv.Spreadsheets.Get(spreadsheetID).Fields("sheets.properties").Do()
	if err != nil {
		return 0, err
	}
	for _, s := range resp.Sheets {
		if s.Properties.Title == name {
			return s.Properties.SheetId, nil
		}
	}
	return 0, fmt.Errorf("лист с названием %q не найден", name)
}

func CreateSheetList(srv *sheets.Service, spreadsheetID, sheetName string) error {
	request := &sheets.AddSheetRequest{
		Properties: &sheets.SheetProperties{
			Title: sheetName,
		},
	}
	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: request,
			},
		},
	}

	_, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("Не удалось создать лист: %v", err)
	}
	fmt.Printf("Лист %s успешно создан!\n", sheetName)
	return nil
}

func WriteToSheet(srv *sheets.Service, spreadsheetID string, data []*sheets.ValueRange) error {
	_, err := srv.Spreadsheets.Values.BatchUpdate(spreadsheetID,
		&sheets.BatchUpdateValuesRequest{ValueInputOption: "USER_ENTERED",
			Data: data}).Do()
	if err != nil {
		return fmt.Errorf("Не удалось записать данные: %v", err)
	}
	return nil
}

func DeleteFromSheet(srv *sheets.Service, spreadsheetID string, ranges []string) error {
	_, err := srv.Spreadsheets.Values.BatchClear(spreadsheetID, &sheets.BatchClearValuesRequest{Ranges: ranges}).Do()
	if err != nil {
		return fmt.Errorf("Не удалось очистить данные: %v", err)
	}
	return nil
}

func GetAllSheetIDByName(srv *drive.Service) (map[string]string, error) {
	files, err := srv.Files.List().Q("mimeType='application/vnd.google-apps.spreadsheet'").Fields("files(id, name)").Do()
	if err != nil {
		return nil, fmt.Errorf("Ошибка получения списка файлов: %v", err)
	}
	sheetID := make(map[string]string)
	for _, file := range files.Files {
		sheetID[file.Name] = file.Id
	}
	return sheetID, nil
}

func CreateSheet(srv *sheets.Service, sheetName string) (string, error) {
	sheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title:      sheetName,
			AutoRecalc: "MINUTE",
			TimeZone:   "GMT+00:00",
		},
	}
	spreadsheet, err := srv.Spreadsheets.Create(sheet).Do()
	if err != nil {
		return "", fmt.Errorf("Не удалось создать таблицу: %v", err)
	}

	return spreadsheet.SpreadsheetId, nil
}

func AddPermission(srv *drive.Service, fileID string, emails *[]string) error {

	for _, mail := range *emails {
		permission := &drive.Permission{
			Role:         "writer",
			Type:         "user",
			EmailAddress: mail,
		}
		_, err := srv.Permissions.Create(fileID, permission).Do()
		if err != nil {
			return fmt.Errorf("Не удалось добавить разрешение: %v", err)
		}
	}
	return nil
}

func CreateSheetWithPermission(sheetSrv *sheets.Service, driveSrv *drive.Service, sheetName string, emails *[]string) (string, error) {
	sheetID, err := CreateSheet(sheetSrv, sheetName)
	if err != nil {
		return "", err
	}
	err = AddPermission(driveSrv, sheetID, emails)
	if err != nil {
		return "", err
	}
	return sheetID, nil
}

func GetCredentials(ctx context.Context, pathToCredentialsFile string) (*google.Credentials, error) {
	file, err := os.Open(pathToCredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("Ошибка открытия файла: %v", err)
	}
	defer file.Close()
	dataFile, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Ошибка чтения файла: ", err)
	}
	cred, err := google.CredentialsFromJSON(ctx, dataFile,
		drive.DriveScope,
		drive.DriveFileScope,
		sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("Ошибка загрузки учетных данных: ", err)
	}
	return cred, nil
}

func GetDriveService(ctx context.Context, cred *google.Credentials) (*drive.Service, error) {
	driveSrv, err := drive.NewService(ctx, option.WithCredentials(cred))
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к Google Drive API: %v", err)
	}
	return driveSrv, nil
}

func GetSheetsService(ctx context.Context, cred *google.Credentials) (*sheets.Service, error) {
	sheetsSrv, err := sheets.NewService(ctx, option.WithCredentials(cred))
	if err != nil {
		return nil, fmt.Errorf("Ошибка подключения к Google Sheets API: %v", err)
	}
	return sheetsSrv, nil
}

func CreateSheetFilter(srv *sheets.Service, spreadsheetID string, sheetID int64, startRowIndex, endRowIndex, startColumnIndex, endColumnIndex int64) error {
	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				SetBasicFilter: &sheets.SetBasicFilterRequest{
					Filter: &sheets.BasicFilter{
						Range: &sheets.GridRange{
							SheetId:          sheetID,
							StartRowIndex:    startRowIndex,
							EndRowIndex:      endRowIndex,
							StartColumnIndex: startColumnIndex,
							EndColumnIndex:   endColumnIndex,
						},
					},
				},
			},
		},
	}

	_, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, request).Do()
	if err != nil {
		return fmt.Errorf("не удалось создать фильтр: %v", err)
	}

	fmt.Printf("Базовый фильтр успешно создан для диапазона [%d:%d, %d:%d]\n",
		startRowIndex, endRowIndex, startColumnIndex, endColumnIndex)
	return nil
}

// DeleteSheetByName удаляет лист из таблицы по имени
func DeleteSheetByName(srv *sheets.Service, spreadsheetID, sheetName string) error {
	// Получаем ID листа по имени
	sheetID, err := SheetIDByName(srv, spreadsheetID, sheetName)
	if err != nil {
		return fmt.Errorf("не удалось найти лист %s: %v", sheetName, err)
	}

	// Создаем запрос на удаление листа
	request := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteSheet: &sheets.DeleteSheetRequest{
					SheetId: sheetID,
				},
			},
		},
	}

	// Выполняем запрос
	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, request).Do()
	if err != nil {
		return fmt.Errorf("не удалось удалить лист %s: %v", sheetName, err)
	}

	fmt.Printf("Лист %s успешно удален!\n", sheetName)
	return nil
}

// DeleteSpreadsheetByID удаляет всю таблицу по ID
func DeleteSpreadsheetByID(driveSrv *drive.Service, spreadsheetID string) error {
	err := driveSrv.Files.Delete(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("не удалось удалить таблицу %s: %v", spreadsheetID, err)
	}

	fmt.Printf("Таблица %s успешно удалена!\n", spreadsheetID)
	return nil
}
