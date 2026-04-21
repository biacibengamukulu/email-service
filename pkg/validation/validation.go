package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()

	Validate.RegisterValidation("phone_number_za", func(fl validator.FieldLevel) bool {
		re := regexp.MustCompile(`^(?:\+27|27|0)[6-8][0-9]{8}$`)
		return re.MatchString(fl.Field().String())
	})
}
