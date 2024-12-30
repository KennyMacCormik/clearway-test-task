package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

func validate(c Config) error {
	val := validator.New(validator.WithRequiredStructEnabled())

	err := val.Struct(c)
	if err != nil {
		return err
	}

	return nil
}

func handleValidatorError(c Config, err error) string {
	// Expected way to handle error according to module docs
	// https://github.com/go-playground/validator?tab=readme-ov-file#error-return-value
	valErr := err.(validator.ValidationErrors)
	errStr := ""

	for _, v := range valErr {
		tag, env := reflectActualTag(c, v.StructField())
		if tag == "" {
			tag = "err reflect tag"
		}
		errStr += fmt.Sprintf("env '%s' value '%s' invalid, '%s' expected; ",
			env, v.Value(), tag)
	}
	errStr = strings.Trim(errStr, " ")

	return errStr
}

func reflectActualTag(c Config, sf string) (string, string) {
	ref := reflect.TypeOf(c)

	for i := 0; i < ref.NumField(); i++ {
		fieldName := ref.Field(i).Name
		field, _ := ref.FieldByName(fieldName)
		if field.Type.Name() != "bool" {
			for j := 0; j < field.Type.NumField(); j++ {
				intFieldName := field.Type.Field(j)
				if intFieldName.Name == sf {
					return intFieldName.Tag.Get("validate"), intFieldName.Tag.Get("env")
				}
			}
		}
	}

	return "", ""
}
