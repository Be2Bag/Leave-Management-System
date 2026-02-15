package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate ตรวจสอบ struct ตาม validation tags
func (v *Validator) Validate(s any) []string {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	// แปลง validation errors เป็นข้อความที่อ่านเข้าใจง่าย
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []string{err.Error()}
	}

	errs := make([]string, 0, len(validationErrors))
	for _, e := range validationErrors {
		errs = append(errs, formatValidationError(e))
	}

	return errs
}

// formatValidationError แปลง validation error เป็นข้อความภาษาอังกฤษที่อ่านง่าย
func formatValidationError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
