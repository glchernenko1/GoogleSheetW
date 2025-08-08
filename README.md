# GoogleSheetW API

API для работы с Google Sheets - создание, обновление и удаление таблиц и листов.

## Запуск приложения

```bash
cd app
go run cmd/main.go
```

По умолчанию сервер запускается на порту 8888.

## API Endpoints

### 1. Установка данных в таблицу
**POST** `/api/sheets/set-data`

Отправляет данные SheetData и вызывает SetSheetData для записи в Google Sheets.

**Пример запроса:**
```json
{
  "sheet_data": {
    "fiat": "USD",
    "soup_list": [
      {
        "name": "TestSheet",
        "fixed_price": 100.5,
        "best_price": 99.8,
        "best_price_link": "https://example.com",
        "date": "2025-01-30",
        "money_supply": 10000.0,
        "average_size": 50,
        "info_filters": [
          {
            "exchange": "Binance",
            "banks_name": ["Sber", "Tinkoff"],
            "month_order": 100,
            "month_finish_rate": 95.5,
            "max_low_single_trans_amount": 1000.0,
            "min_high_single_trans_amount": 10000.0,
            "average_size": 50
          }
        ],
        "data": [
          ["Exchange", "Fiat", "Asset", "Price"],
          ["Binance", "USD", "BTC", "50000"]
        ]
      }
    ],
    "raw_data": {
      "date": "2025-01-30",
      "raw_data": [
        ["Col1", "Col2", "Col3"],
        ["Data1", "Data2", "Data3"]
      ]
    }
  }
}
```

**Пример ответа:**
```json
{
  "success": true,
  "message": "Данные успешно установлены",
  "data": null
}
```

### 2. Удаление листа из таблицы
**DELETE** `/api/sheets/{fiat}/sheet/{sheetName}`

Удаляет конкретный лист из таблицы. Параметры передаются в URL:
- `{fiat}` - название валюты (например, USD, EUR)
- `{sheetName}` - имя листа для удаления

**Пример запроса:**
```bash
DELETE /api/sheets/USD/sheet/TestSheet
```

**Пример ответа:**
```json
{
  "success": true,
  "message": "Лист успешно удален",
  "data": {
    "fiat": "USD",
    "sheet_name": "TestSheet"
  }
}
```

### 3. Удаление всей таблицы
**DELETE** `/api/sheets/{fiat}`

Удаляет всю таблицу по названию валюты. Параметр передается в URL:
- `{fiat}` - название валюты (например, USD, EUR)

**Пример запроса:**
```bash
DELETE /api/sheets/USD
```

**Пример ответа:**
```json
{
  "success": true,
  "message": "Таблица успешно удалена",
  "data": {
    "fiat": "USD"
  }
}
```

### 4. Health Check
**GET** `/health`

Проверка состояния сервиса.

**Пример ответа:**
```json
{
  "status": "ok"
}
```

## Примеры использования с curl

### Установка данных:
```bash
curl -X POST http://localhost:8888/api/sheets/set-data \
  -H "Content-Type: application/json" \
  -d '{
    "sheet_data": {
      "fiat": "USD",
      "soup_list": [{
        "name": "TestSheet",
        "fixed_price": 100.5,
        "best_price": 99.8,
        "best_price_link": "https://example.com",
        "date": "2025-01-30",
        "money_supply": 10000.0,
        "average_size": 50,
        "info_filters": [],
        "data": [["Test", "Data"]]
      }],
      "raw_data": {
        "date": "2025-01-30",
        "raw_data": [["Raw", "Data"]]
      }
    }
  }'
```

### Удаление листа:
```bash
curl -X DELETE http://localhost:8888/api/sheets/USD/sheet/TestSheet
```

### Удаление таблицы:
```bash
curl -X DELETE http://localhost:8888/api/sheets/USD
```

### Health Check:
```bash
curl -X GET http://localhost:8888/health
```

## Структура проекта

- `/cmd/main.go` - точка входа в приложение
- `/internal/app/app.go` - основное приложение и настройка роутов
- `/internal/controller/sheets_controller.go` - HTTP контроллер для обработки запросов
- `/internal/models/` - модели данных и структуры запросов/ответов
- `/internal/services/sheetsControl/` - бизнес-логика работы с Google Sheets
- `/internal/services/googleAPI/` - обертки для Google Sheets и Drive API

## Примечания

- Операции удаления используют RESTful подход с параметрами в URL
- Для полноценной работы требуется файл `google.json` с учетными данными Google API
- Логи записываются в файл `log/log.log`
- Поддерживается URL-кодирование для параметров с специальными символами

localhost:8888
