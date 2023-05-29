package v1

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

var InvalidRequestExceptionMessage = "Bad Request"
var UnauthorizedExceptionMessage = "Unauthorized"
var ForbiddenExceptionMessage = "Forbidden"
var InternalServerExceptionMessage = "Internal Server Error"

type Error struct {
	Message string `json:"message"`
}

type RestError struct {
	// json:"-" omits the field from marshalling
	Status int `json:"-"`
	Error
}

func BadRequestError() *RestError {
	return &RestError{
		Status: http.StatusBadRequest,
		Error: Error{
			Message: InvalidRequestExceptionMessage,
		},
	}
}

func UnauthorizedError() *RestError {
	return &RestError{
		Status: http.StatusUnauthorized,
		Error: Error{
			Message: UnauthorizedExceptionMessage,
		},
	}
}

func ForbiddenError() *RestError {
	return &RestError{
		Status: http.StatusForbidden,
		Error: Error{
			Message: ForbiddenExceptionMessage,
		},
	}
}

func InternalServerError() *RestError {
	return &RestError{
		Status: http.StatusInternalServerError,
		Error: Error{
			Message: InternalServerExceptionMessage,
		},
	}
}

func parseBindingError(err error) *RestError {
	fieldErrors := make(map[string]string)
	for _, v := range err.(validator.ValidationErrors) {
		fieldErrors[toCamelCase(v.Field())] = v.Tag()
	}
	logrus.Errorf("Failed to bind to query: %+v", fieldErrors)
	return BadRequestError()
}
