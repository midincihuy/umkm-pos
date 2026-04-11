package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(s interface{}) []ValidationError {
	var errs []ValidationError
	if err := validate.Struct(s); err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			errs = append(errs, ValidationError{
				Field:   e.Field(),
				Message: msgForTag(e.Tag(), e.Param()),
			})
		}
	}
	return errs
}

func msgForTag(tag, param string) string {
	switch tag {
	case "required":
		return "field ini wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		return "nilai minimal adalah " + param
	case "max":
		return "nilai maksimal adalah " + param
	case "gt":
		return "nilai harus lebih besar dari " + param
	case "oneof":
		return "nilai harus salah satu dari: " + param
	}
	return "validasi gagal"
}
