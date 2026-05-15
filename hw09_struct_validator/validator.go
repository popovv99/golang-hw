package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, e := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(e.Field)
		sb.WriteString(": ")
		sb.WriteString(e.Err.Error())
	}
	return sb.String()
}

func ValidateString(strValue, ruleName, ruleParam, fieldName string) (ValidationError, error) {
	switch ruleName {
	case "len":
		expectedLen, err := strconv.Atoi(ruleParam)
		if err != nil {
			return ValidationError{}, fmt.Errorf("invalid len rule: %w", err)
		}
		if len(strValue) != expectedLen {
			return ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("length mismatch: expected %d, got %d", expectedLen, len(strValue)),
			}, nil
		}
	case "regexp":
		re, err := regexp.Compile(ruleParam)
		if err != nil {
			return ValidationError{}, fmt.Errorf("invalid regexp rule: %w", err)
		}
		if !re.MatchString(strValue) {
			return ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("regexp mismatch: expected %s, got %s", ruleParam, strValue),
			}, nil
		}
	case "in":
		allowedValues := strings.Split(ruleParam, ",")
		found := false
		for _, allowed := range allowedValues {
			if allowed == strValue {
				found = true
				break
			}
		}
		if !found {
			return ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("value %s not in allowed values %v", strValue, allowedValues),
			}, nil
		}
	default:
		return ValidationError{}, fmt.Errorf("unknown validation rule for string: %s", ruleName)
	}
	return ValidationError{}, nil
}

func ValidateInt(intValue int64, ruleName, ruleParam, fieldName string) (ValidationError, error) {
	switch ruleName {
	case "min":
		minValue, err := strconv.Atoi(ruleParam)
		if err != nil {
			return ValidationError{}, fmt.Errorf("invalid min rule: %w", err)
		}
		if intValue < int64(minValue) {
			return ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("value %d is less than minimum %d", intValue, minValue),
			}, nil
		}
	case "max":
		maxValue, err := strconv.Atoi(ruleParam)
		if err != nil {
			return ValidationError{}, fmt.Errorf("invalid max rule: %w", err)
		}
		if intValue > int64(maxValue) {
			return ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("value %d is greater than maximum %d", intValue, maxValue),
			}, nil
		}
	case "in":
		allowedValues := strings.Split(ruleParam, ",")
		found := false
		for _, allowed := range allowedValues {
			if allowed == strconv.FormatInt(intValue, 10) {
				found = true
				break
			}
		}
		if !found {
			return ValidationError{
				Field: fieldName,
				Err:   fmt.Errorf("value %d not in allowed values %v", intValue, allowedValues),
			}, nil
		}
	default:
		return ValidationError{}, fmt.Errorf("unknown validation rule for int: %s", ruleName)
	}
	return ValidationError{}, nil
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Struct {
		return errors.New("input must be a struct")
	}

	var validationErrors ValidationErrors
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)
		validateTag := field.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		fieldErrors, err := validateField(field, value, validateTag)
		if err != nil {
			return err
		}
		validationErrors = append(validationErrors, fieldErrors...)
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

func validateField(field reflect.StructField, value reflect.Value, validateTag string) (ValidationErrors, error) {
	var validationErrors ValidationErrors
	rules := strings.Split(validateTag, "|")

	for _, rule := range rules {
		parts := strings.Split(rule, ":")
		if len(parts) <= 1 {
			continue
		}

		ruleName := parts[0]
		ruleParam := parts[1]
		fieldErr, err := validateRule(field, value, ruleName, ruleParam)
		if err != nil {
			return nil, err
		}
		if fieldErr.Field != "" {
			validationErrors = append(validationErrors, fieldErr)
		}
	}

	return validationErrors, nil
}

func validateRule(field reflect.StructField, value reflect.Value, ruleName, ruleParam string) (ValidationError, error) {
	switch value.Kind() { //nolint:exhaustive
	case reflect.String:
		return ValidateString(value.String(), ruleName, ruleParam, field.Name)
	case reflect.Int:
		return ValidateInt(value.Int(), ruleName, ruleParam, field.Name)
	case reflect.Slice:
		return validateSlice(field, value, ruleName, ruleParam)
	}
	return ValidationError{}, nil
}

func validateSlice(
	field reflect.StructField,
	value reflect.Value,
	ruleName, ruleParam string,
) (ValidationError, error) {
	typ := value.Type().Elem().Kind()
	for j := 0; j < value.Len(); j++ {
		item := value.Index(j)
		switch typ { //nolint:exhaustive
		case reflect.String:
			validationError, err := ValidateString(item.String(), ruleName, ruleParam, field.Name)
			if err != nil {
				return ValidationError{}, err
			}
			if validationError.Field != "" {
				return validationError, nil
			}
		case reflect.Int:
			validationError, err := ValidateInt(item.Int(), ruleName, ruleParam, field.Name)
			if err != nil {
				return ValidationError{}, err
			}
			if validationError.Field != "" {
				return validationError, nil
			}
		}
	}
	return ValidationError{}, nil
}
