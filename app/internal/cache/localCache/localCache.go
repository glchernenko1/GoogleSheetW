package localCache

import (
	"GoogleSheetW/internal/apperrors"
	"fmt"
	"sync"
)

// MapCache реализует интерфейс Cache с использованием карт в памяти
type MapCache struct {
	fiatToID map[string]string          // ключ: fiat, значение: ID таблицы
	supMap   map[string]map[string]bool // ключ1: fiat, ключ2: supName, значение: bool (наличие)
	mu       sync.RWMutex               // мьютекс для безопасности при конкурентном доступе
}

var (
	instance *MapCache
	once     sync.Once
)

// GetInstance возвращает единственный экземпляр кэша (синглтон)
func GetInstance() *MapCache {
	once.Do(func() {
		instance = &MapCache{
			fiatToID: make(map[string]string),
			supMap:   make(map[string]map[string]bool),
		}
	})
	return instance
}

// GetIDbyFiat возвращает ID таблицы по ключу fiat
func (c *MapCache) GetIDbyFiat(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	id, exists := c.fiatToID[key]
	if !exists {
		return "", apperrors.ErrCacheNotFound
	}
	return id, nil
}

// SetIDbyFiat устанавливает ID таблицы для ключа fiat
func (c *MapCache) SetIDbyFiat(key string, value string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.fiatToID[key] = value
	return nil
}

// IsSupInCashed проверяет, есть ли лист в кэше для указанной валюты
func (c *MapCache) IsSupInCashed(fiat, supName string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	supMapForFiat, exists := c.supMap[fiat]
	if !exists {
		return false, nil
	}

	return supMapForFiat[supName], nil
}

// SetSupInCashed добавляет лист в кэш для указанной валюты
func (c *MapCache) SetSupInCashed(fiat, supName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.supMap[fiat]; !exists {
		c.supMap[fiat] = make(map[string]bool)
	}

	c.supMap[fiat][supName] = true
	return nil
}

// RemoveSupFromCashed удаляет лист из кэша для указанной валюты
func (c *MapCache) RemoveSupFromCashed(fiat, supName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	supMapForFiat, exists := c.supMap[fiat]
	if !exists {
		return fmt.Errorf("валюта %s не найдена в кэше", fiat)
	}

	delete(supMapForFiat, supName)
	return nil
}

// RemoveFiatFromCache удаляет всю валюту и связанные с ней листы из кэша
func (c *MapCache) RemoveFiatFromCache(fiat string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.fiatToID, fiat)
	delete(c.supMap, fiat)
	return nil
}

// AddToCash добавляет в кэш сразу несколько таблиц
func (c *MapCache) AddToCash(idMap map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for fiat, id := range idMap {
		c.fiatToID[fiat] = id
		// Инициализируем карту для листов, если её ещё нет
		if _, exists := c.supMap[fiat]; !exists {
			c.supMap[fiat] = make(map[string]bool)
		}
	}

	return nil
}
