package models


type Localization string

const (
	HeaderLanguage  = "Accept-Language"
	KeyLanguage  Localization = "lang"
	LanguageDefault Localization = "es"
	LanguageRu Localization = "ru"
	LanguageEn Localization = "en"
	LanguageEs Localization = "es"
)

var LocalMap = map[Localization]struct{}{
	LanguageRu: {},
	LanguageEn: {},
	LanguageEs: {},
}