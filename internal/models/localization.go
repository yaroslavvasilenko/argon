package models


type Localization string

const (
	HeaderLanguage  = "Accept-Language"
	KeyLanguage  = "lang"
	LanguageDefault  = "es"
	LanguageRu Localization = "ru"
	LanguageEn Localization = "en"
	LanguageEs Localization = "es"
)