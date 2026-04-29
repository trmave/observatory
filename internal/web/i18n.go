package web

import (
	"encoding/json"
	"os"
	"strings"
)

type Translations map[string]map[string]string

var translations Translations

func LoadTranslations(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &translations)
}

func GetLanguage(acceptLanguage string) string {
	if strings.Contains(acceptLanguage, "en") {
		return "en"
	}
	return "es" // Default
}

func (t Translations) Get(lang, key string) string {
	if val, ok := t[lang][key]; ok {
		return val
	}
	return t["es"][key] // Fallback
}
