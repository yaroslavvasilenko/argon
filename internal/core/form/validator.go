package form

import (
	"encoding/json"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Errors map[string][]string `json:"errors"`
}

func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string][]string),
	}
}

func validate(s interface{}) *Validator {
	v := NewValidator()

	val := validator.New()

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
	// normalizedAttr := NormalizeAttr(attr)
	v.Errors[attr] = append(v.Errors[attr], desc)
}

func (v *Validator) AttrErrors(attr string) []string {
	if errors, ok := v.Errors[attr]; ok {
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
