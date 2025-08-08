package sheetsControl

import (
	"GoogleSheetW/internal/apperrors"
	"GoogleSheetW/internal/cache"
	"GoogleSheetW/internal/logger"
	"GoogleSheetW/internal/models"
	"GoogleSheetW/internal/services/googleAPI"
	"GoogleSheetW/internal/settings"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
	"strings"
)

func dataTime(nameList string) []*sheets.ValueRange {
	return []*sheets.ValueRange{
		{
			Range: fmt.Sprintf("'%s'!A1", nameList),
			Values: [][]interface{}{
				{"Дата записи:"},
			},
		},
		{
			Range: fmt.Sprintf("'%s'!A2:B2", nameList),
			Values: [][]interface{}{
				{"Текущая дата", "=NOW()- TIME(0, 0, 0)"},
			},
		},
		{
			Range: fmt.Sprintf("'%s'!A3:B3", nameList),
			Values: [][]interface{}{
				{"Минут прошло", "=IF(B1=\"\", \"\", ROUND((B2 - B1) * 1440,2))"},
			},
		},
	}
}

func rawDataFilterFunc() []*sheets.ValueRange {
	return []*sheets.ValueRange{
		{
			Range: "RAW_filter!A6",
			Values: [][]interface{}{
				{"=QUERY(RAW!$A4:O,\"select A,B,C,D,E,F,G,H,I,J,K,L,M,N,O\")"},
			},
		},
	}
}

func toInterface2D(data [][]string) [][]interface{} {
	result := make([][]interface{}, len(data))
	for i, row := range data {
		result[i] = make([]interface{}, len(row))
		for j := range row {
			result[i][j] = row[j]
		}
	}
	return result
}

func soupToSheet(soup models.Soup) []*sheets.ValueRange {
	return []*sheets.ValueRange{
		{
			Range: fmt.Sprintf("'%s'!A5", soup.Name),
			Values: [][]interface{}{
				{"Fixed Price:", soup.FixedPrice},
				{"Best Price:", soup.BestPrice},
				{"Best Price Link", soup.BestPriceLink},
				{"Money Supply:", soup.MoneySupply},
				{"Выборка из:", soup.AverageSize},
			},
		},
		{
			Range: fmt.Sprintf("'%s'!B1", soup.Name),
			Values: [][]interface{}{
				{soup.Date},
			},
		},
		{
			Range: fmt.Sprintf("'%s'!A10", soup.Name),
			Values: func() [][]interface{} {
				var ans [][]interface{}
				for _, InfoFilter := range soup.InfoFilters {
					// Преобразуем массив банков в строку через запятую
					banksString := strings.Join(InfoFilter.BanksName, ", ")
					ans = append(ans,
						[]interface{}{"Биржа:", InfoFilter.Exchange},
						[]interface{}{"Банки:", banksString},
						[]interface{}{"Выполненных заказов", InfoFilter.MonthOrder},
						[]interface{}{"Процент выполненных заказов", InfoFilter.MonthFinishRate},
						[]interface{}{"Макс. минимальная сумма транзакции", InfoFilter.MaxLowSingleTransAmount},
						[]interface{}{"Мин. максимальная сумма транзакции", InfoFilter.MinHighSingleTransAmount},
						[]interface{}{"Размер выборки", InfoFilter.AverageSize},
						[]interface{}{"_______________", "________________"},
					)
				}
				ans = append(ans,
					[]interface{}{
						"Exchange:", "Fiat", "Asset",
						"TradeType", "NickName", "Price",
						"MonthOrderCount", "MonthFinishRate",
						"UserType", "MaxSingleTransAmount",
						"MinSingleTransAmount", "LastQuantity",
						"Link", "PaymentMethod", "monthFinishRate"})
				ans = append(ans, toInterface2D(soup.Data)...)
				return ans
			}(),
		},
	}

}

type SheetsControl struct {
	ctx      context.Context
	sheetSrv *sheets.Service
	driveSrv *drive.Service
	cache    cache.Cache
	log      *zap.SugaredLogger
}

func New(ctx context.Context, cache cache.Cache, pathToCredentialsFile string) *SheetsControl {
	log := logger.Get()
	cred, err := googleAPI.GetCredentials(ctx, pathToCredentialsFile)
	if err != nil {
		log.Errorw("Ошибка создания SheetsControl: ", err)
		panic(err)
	}
	sheetSrv, err := googleAPI.GetSheetsService(ctx, cred)
	if err != nil {
		log.Errorw("Ошибка создания SheetsControl: ", err)
		panic(err)
	}
	driveSrv, err := googleAPI.GetDriveService(ctx, cred)
	if err != nil {
		log.Errorw("Ошибка создания SheetsControl: ", err)
		panic(err)
	}
	ans := SheetsControl{
		ctx:      ctx,
		sheetSrv: sheetSrv,
		driveSrv: driveSrv,
		cache:    cache,
		log:      log,
	}
	err = ans.update()
	if err != nil {
		log.Errorw("Ошибка создания SheetsControl: ", err)
		panic(err)
	}
	return &ans
}

func (sc *SheetsControl) update() error {
	// Получаем карту с именами и ID всех таблиц
	sheetIDMap, err := googleAPI.GetAllSheetIDByName(sc.driveSrv)
	if err != nil {
		sc.log.Errorw("Ошибка получения id таблиц: ", err)
		return err
	}

	// Добавляем информацию о таблицах в кэш
	err = sc.cache.AddToCash(sheetIDMap)
	if err != nil {
		sc.log.Errorw("Ошибка инициализации кеша таблиц: ", err)
		return err
	}

	// Проходим по всем таблицам и получаем список листов для каждой
	for fiat, spreadsheetID := range sheetIDMap {
		// Получаем список всех листов в таблице
		spreadsheet, err := sc.sheetSrv.Spreadsheets.Get(spreadsheetID).Do()
		if err != nil {
			sc.log.Warnw("Ошибка получения списка листов для таблицы",
				"fiat", fiat,
				"spreadsheetID", spreadsheetID,
				"error", err)
			continue
		}

		// Добавляем каждый лист в кэш
		for _, sheet := range spreadsheet.Sheets {
			sheetName := sheet.Properties.Title
			err = sc.cache.SetSupInCashed(fiat, sheetName)
			if err != nil {
				sc.log.Warnw("Ошибка добавления листа в кэш",
					"fiat", fiat,
					"sheetName", sheetName,
					"error", err)
			}
		}

		sc.log.Infow("Обновлены данные о листах в таблице",
			"fiat", fiat,
			"sheets_count", len(spreadsheet.Sheets))
	}

	sc.log.Info("Кэш успешно обновлен")
	return nil
}

func (sc *SheetsControl) SetSheetData(data models.SheetData) error {
	sheetID, err := sc.cache.GetIDbyFiat(data.Fiat)
	if err != nil {
		switch {
		case errors.As(err, &apperrors.ErrCacheNotFound):
			{
				_emails := settings.GetSettings().GetEmails()
				sheetID, err = googleAPI.CreateSheetWithPermission(sc.sheetSrv, sc.driveSrv, data.Fiat, &_emails)
				if err != nil {
					sc.log.Errorw("Ошибка создания таблицы: ", err)
					return err
				}
				err := sc.cache.SetIDbyFiat(data.Fiat, sheetID)
				if err != nil {
					sc.log.Errorw("Ошибка добавления  id таблицы в hash: ", err)
					return err
				}
				break
			}
		default:
			{
				sc.log.Errorw("Неизвестная ошибка", err)
				return err

			}
		}
	}
	var ans []*sheets.ValueRange
	var delAns []string
	for _, soup := range data.SoupList {
		if ok, err := sc.cache.IsSupInCashed(data.Fiat, soup.Name); err == nil {
			if !ok {
				err := googleAPI.CreateSheetList(sc.sheetSrv, sheetID, soup.Name)
				if err != nil {
					sc.log.Errorw(fmt.Sprintf("Ошибка создания листа для %s %s c ID %s ", data.Fiat, soup.Name, sheetID), err)
					return err
				}
				err = sc.cache.SetSupInCashed(data.Fiat, soup.Name)
				if err != nil {
					sc.log.Errorw(fmt.Sprintf("Ошибка добавление в cache название супа %s %s"), data.Fiat, soup.Name)
					return err
				}
				err = googleAPI.WriteToSheet(sc.sheetSrv, sheetID, dataTime(soup.Name))
				if err != nil {
					sc.log.Errorw(fmt.Sprintf("Ошибка записи данных в лист %s %s", data.Fiat, soup.Name), err)
					return err
				}
			}

		} else {
			sc.log.Errorw(fmt.Sprintf("ошибка чтения кэша %s %s", data.Fiat, soup.Name), err)
			return err
		}
		ans = append(ans, soupToSheet(soup)...)
		delAns = append(delAns, fmt.Sprintf("'%s'!A10:O100", soup.Name))
	}
	if ok, err := sc.cache.IsSupInCashed(data.Fiat, "RAW"); err == nil {
		if !ok {
			err := googleAPI.CreateSheetList(sc.sheetSrv, sheetID, "RAW")
			if err != nil {
				sc.log.Errorw(fmt.Sprintf("Ошибка создания листа для %s  %s c ID %s ", data.Fiat, "RAW", sheetID), err)
				return err
			}
			err = googleAPI.CreateSheetList(sc.sheetSrv, sheetID, "RAW_filter")
			if err != nil {
				sc.log.Errorw(fmt.Sprintf("Ошибка создания листа для %s  %s c ID %s ", data.Fiat, "RAW_filter", sheetID), err)
				return err
			}

			err = sc.cache.SetSupInCashed(data.Fiat, "RAW")
			if err != nil {
				sc.log.Errorw(fmt.Sprintf("Ошибка добавление в cache %s RAW"), data.Fiat)
				return err
			}

			err = googleAPI.WriteToSheet(sc.sheetSrv, sheetID, dataTime("RAW"))
			if err != nil {
				sc.log.Errorw(fmt.Sprintf("Ошибка записи данных в лист %s %s", data.Fiat, "RAW"), err)
				return err
			}
			idList, err := googleAPI.SheetIDByName(sc.sheetSrv, sheetID, "RAW_filter")
			if err != nil {
				sc.log.Errorw(fmt.Sprintf("Ошибка получения ID листа RAW_filter %s", data.Fiat), err)
				return err
			}
			err = googleAPI.CreateSheetFilter(sc.sheetSrv, sheetID, idList, 0, 14, 5, 4500)
			if err != nil {
				sc.log.Errorw(fmt.Sprintf("Ошибка создания фильтра в листе RAW_filter %s", data.Fiat), err)
				return err
			}

			ans = append(ans, rawDataFilterFunc()...)
		}
		rawData := []*sheets.ValueRange{
			{
				Range: fmt.Sprintf("RAW!B1"),
				Values: [][]interface{}{
					{data.RAWData.Date},
				},
			},
			{
				Range:  fmt.Sprintf("RAW!A4:Q4500"),
				Values: toInterface2D(data.RAWData.Data),
			},
		}
		ans = append(ans, rawData...)
		delAns = append(delAns, "RAW!A4:Q4500")

	} else {
		sc.log.Errorw(fmt.Sprintf("ошибка чтения кэша %s RawData", data.Fiat), err)
		return err
	}
	err = googleAPI.DeleteFromSheet(sc.sheetSrv, sheetID, delAns)
	if err != nil {
		sc.log.Errorw(fmt.Sprintf("Ошибка удаления данных в листе %s", data.Fiat), err)
		return err
	}
	err = googleAPI.WriteToSheet(sc.sheetSrv, sheetID, ans)
	if err != nil {
		sc.log.Errorw(fmt.Sprintf("Ошибка записи данных в лист %s", data.Fiat), err)
		return err
	}
	return nil
}

// DeleteSheet удаляет лист из таблицы
func (sc *SheetsControl) DeleteSheet(fiat, sheetName string) error {
	sc.log.Infow("Удаление листа из таблицы",
		"fiat", fiat,
		"sheet_name", sheetName)

	// Получаем ID таблицы из кэша
	sheetID, err := sc.cache.GetIDbyFiat(fiat)
	if err != nil {
		sc.log.Errorw("Таблица не найдена", "fiat", fiat, "error", err)
		return fmt.Errorf("таблица для %s не найдена: %v", fiat, err)
	}

	// Проверяем, что лист существует в кэше
	exists, err := sc.cache.IsSupInCashed(fiat, sheetName)
	if err != nil {
		sc.log.Errorw("Ошибка проверки листа в кэше", "fiat", fiat, "sheet_name", sheetName, "error", err)
		return fmt.Errorf("ошибка проверки листа %s: %v", sheetName, err)
	}

	if !exists {
		sc.log.Warnw("Лист не найден в кэше", "fiat", fiat, "sheet_name", sheetName)
		return fmt.Errorf("лист %s не найден в таблице %s", sheetName, fiat)
	}

	// Удаляем лист через Google Sheets API
	err = googleAPI.DeleteSheetByName(sc.sheetSrv, sheetID, sheetName)
	if err != nil {
		sc.log.Errorw("Ошибка удаления листа через API", "fiat", fiat, "sheet_name", sheetName, "error", err)
		return fmt.Errorf("не удалось удалить лист %s: %v", sheetName, err)
	}

	// Удаляем лист из кэша
	err = sc.cache.RemoveSupFromCashed(fiat, sheetName)
	if err != nil {
		sc.log.Warnw("Ошибка удаления листа из кэша", "fiat", fiat, "sheet_name", sheetName, "error", err)
		// Не возвращаем ошибку, так как лист уже удален из Google Sheets
	}

	sc.log.Infow("Лист успешно удален", "fiat", fiat, "sheet_name", sheetName)
	return nil
}

// DeleteSpreadsheet удаляет всю таблицу
func (sc *SheetsControl) DeleteSpreadsheet(fiat string) error {
	sc.log.Infow("Удаление таблицы",
		"fiat", fiat)

	// Получаем ID таблицы из кэша
	sheetID, err := sc.cache.GetIDbyFiat(fiat)
	if err != nil {
		sc.log.Errorw("Таблица не найдена", "fiat", fiat, "error", err)
		return fmt.Errorf("таблица для %s не найдена: %v", fiat, err)
	}

	// Удаляем таблицу через Google Drive API
	err = googleAPI.DeleteSpreadsheetByID(sc.driveSrv, sheetID)
	if err != nil {
		sc.log.Errorw("Ошибка удаления таблицы через API", "fiat", fiat, "error", err)
		return fmt.Errorf("не удалось удалить таблицу %s: %v", fiat, err)
	}

	// Удаляем валюту и все связанные листы из кэша
	err = sc.cache.RemoveFiatFromCache(fiat)
	if err != nil {
		sc.log.Warnw("Ошибка удаления валюты из кэша", "fiat", fiat, "error", err)
		// Не возвращаем ошибку, так как таблица уже удалена из Google Drive
	}

	sc.log.Infow("Таблица успешно удалена", "fiat", fiat)
	return nil
}
