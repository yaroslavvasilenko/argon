package validator

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/yaroslavvasilenko/argon/config"
)

type Validator struct {
	Errors map[string][]string `json:"errors"`
}

func newValidator() *Validator {
	return &Validator{
		Errors: make(map[string][]string),
	}
}

func Validate(s interface{}) *Validator {
	v := newValidator()

	val := validator.New()

	_ = val.RegisterValidation("not_blank", NotBlank)

	_ = val.RegisterValidation("uuid_slice", func(fl validator.FieldLevel) bool {
		slice, ok := fl.Field().Interface().([]string)

		if !ok {
			return false
		}

		for _, item := range slice {
			_, err := uuid.Parse(item)
			if err != nil {
				return false
			}
		}

		return true
	})
	
	// Регистрация кастомной валидации для категорий
	_ = val.RegisterValidation("categories_validation", ValidateCategories)
	
	// Регистрация кастомной валидации для характеристик
	_ = val.RegisterValidation("characteristics_value", ValidateCharacteristicsValue)

	if err := val.StructExcept(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			index := strings.Index(e.Namespace(), ".")
			name := e.Namespace()
			attr := name[index+1:]
			desc := e.Tag()

			if e.Param() != "" {
				desc = desc + ":" + e.Param()
			}
			v.AddError(attr, desc)
		}
	}

	return v
}

func (v *Validator) AddError(attr, desc string) {
	normalizedAttr := NormalizeAttr(attr)
	v.Errors[normalizedAttr] = append(v.Errors[normalizedAttr], desc)
}

func (v *Validator) AttrErrors(attr string) []string {
	if errors, ok := v.Errors[NormalizeAttr(attr)]; ok {
		return errors
	}
	return []string{}
}

func (v *Validator) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *Validator) HasAttrErrors(attr string) bool {
	return len(v.AttrErrors(attr)) > 0
}

func (v *Validator) Error() error {
	if v.HasErrors() {
		return NewValidationError(v.Errors)
	}
	return nil
}

// Convert string to the lower camel case
func NormalizeAttr(attr string) string {
	attr = strings.ReplaceAll(attr, " ", "")
	words := strings.SplitAfter(attr, ".")

	for i, word := range words {
		word = strings.NewReplacer("[", ".", "]", "").Replace(word)
		word = strings.ToLower(word[:1]) + word[1:]
		word = strings.ReplaceAll(word, " ", "")
		words[i] = word
	}

	return strings.Join(words, "")
}

type ValidationError struct {
	Errors map[string][]string `json:"errors"`
}

func NewValidationError(errors map[string][]string) *ValidationError {
	return &ValidationError{
		Errors: errors,
	}
}

func (e *ValidationError) Error() string {
	j, _ := json.Marshal(e.Errors)
	return string(j)
}



func NotBlank(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return len(strings.Trim(strings.TrimSpace(field.String()), "\x1c\x1d\x1e\x1f")) > 0
	case reflect.Chan, reflect.Map, reflect.Slice, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface, reflect.Func:
		return !field.IsNil()
	default:
		return field.IsValid() && field.Interface() != reflect.Zero(field.Type()).Interface()
	}
}


// ValidateCategories проверяет, что категории существуют
func ValidateCategories(fl validator.FieldLevel) bool {
	categories, ok := fl.Field().Interface().([]string)
	if !ok || len(categories) == 0 {
		return false
	}

	// Получаем конфиг
	cfg := config.GetConfig()

	// Проверяем, что все категории существуют
	for _, category := range categories {
		// Проверяем наличие категории в списке доступных категорий
		if _, ok := cfg.Categories.CategoryIds[category]; !ok {
			return false
		}
	}

	return true
}
