package settings

import (
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
		setting = &Settings{
			emails: []string{"glchernenko1@gmail.com"},
		}
	})
	return setting
}
