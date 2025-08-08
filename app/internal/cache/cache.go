package cache

type Cache interface {
	GetIDbyFiat(key string) (string, error) // if "not found" return apperrors "not found"
	SetIDbyFiat(key string, value string) error
	IsSupInCashed(fiat, supName string) (bool, error)
	SetSupInCashed(fiat, supName string) error
	AddToCash(idMap map[string]string) error
	RemoveSupFromCashed(fiat, supName string) error
	RemoveFiatFromCache(fiat string) error
}
