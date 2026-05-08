package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	InvalidTag struct {
		Field string `validate:"unknown:rule"`
	}

	InvalidRegexp struct {
		Field string `validate:"regexp:[invalid"`
	}

	InvalidParam struct {
		Field int `validate:"min:abc"`
	}
)

func TestValidate(t *testing.T) {
	tests := getTestCases()

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			if tt.expectedErr == nil {
				assertNoError(t, err)
			} else {
				assertError(t, err, tt.expectedErr)
			}
		})
	}
}

func getTestCases() []struct {
	in          interface{}
	expectedErr error
} {
	return []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "123",
				Name:   "John",
				Age:    25,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"1234567890"},
			},
			expectedErr: ValidationErrors{
				{
					Field: "ID",
					Err:   fmt.Errorf("invalid ID"),
				},
				{
					Field: "Phones",
					Err:   fmt.Errorf("invalid Phone"),
				},
			},
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    15,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{
					Field: "Age",
					Err:   fmt.Errorf("invalid Age"),
				},
			},
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    55,
				Email:  "john@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{
					Field: "Age",
					Err:   fmt.Errorf("invalid Age"),
				},
			},
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "invalid-email",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{
					Field: "Email",
					Err:   fmt.Errorf("invalid Email"),
				},
			},
		},
		{
			in: User{
				ID:     "123456789012345678901234567890123456",
				Name:   "John",
				Age:    25,
				Email:  "john@example.com",
				Role:   "guest",
				Phones: []string{"12345678901"},
			},
			expectedErr: ValidationErrors{
				{
					Field: "Role",
					Err:   fmt.Errorf("invalid Role"),
				},
			},
		},
		{
			in: Response{
				Code: 403,
			},
			expectedErr: ValidationErrors{
				{
					Field: "Code",
					Err:   fmt.Errorf("invalid Code"),
				},
			},
		},
		{
			in: Response{
				Code: 200,
			},
			expectedErr: nil,
		},
		{
			in: App{
				Version: "1234",
			},
			expectedErr: ValidationErrors{
				{
					Field: "Version",
					Err:   fmt.Errorf("invalid Version"),
				},
			},
		},
		{
			in:          "not a struct",
			expectedErr: errors.New("input must be a struct"),
		},
		{
			in:          InvalidTag{Field: "test"},
			expectedErr: errors.New("unknown validation rule"),
		},
		{
			in:          InvalidRegexp{Field: "test"},
			expectedErr: errors.New("invalid regexp"),
		},
		{
			in:          InvalidParam{Field: 10},
			expectedErr: errors.New("invalid min rule"),
		},
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	var validationErrors ValidationErrors
	if errors.As(err, &validationErrors) && len(validationErrors) > 0 {
		t.Errorf("expected no error, got validation errors: %v", err)
		return
	}
	t.Errorf("expected no error, got error: %v", err)
}

func assertError(t *testing.T, err error, expectedErr error) {
	t.Helper()
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	var validationErrors ValidationErrors
	if errors.As(err, &validationErrors) {
		var expectedErrs ValidationErrors
		if errors.As(expectedErr, &expectedErrs) {
			if len(validationErrors) != len(expectedErrs) {
				t.Errorf("expected %d validation errors, got %d",
					len(expectedErrs), len(validationErrors))
			}
		}
		return
	}
	expectedErrStr := expectedErr.Error()
	if !strings.Contains(err.Error(), expectedErrStr) {
		t.Errorf("expected error containing %q, got %q",
			expectedErrStr, err.Error())
	}
}
