package settings

import (
	"GoogleSheetW/internal/config"
	"sync"
)

var (
	once    sync.Once
	setting *Settings
)

type Settings struct {
	emails []string
}

func (s *Settings) GetEmails() []string {
	return s.emails
}

func GetSettings() *Settings {
	once.Do(func() {
		cfg := config.GetConfig()
		emails := cfg.GetEmails()

		// Если email не найдены ни в переменной окружения, ни в файле,
		// используем резервный email
		if len(emails) == 0 {
			emails = []string{"glchernenko1@gmail.com"}
		}

		setting = &Settings{
			emails: emails,
		}
	})
	return setting
}
