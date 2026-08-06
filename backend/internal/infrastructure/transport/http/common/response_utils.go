package common

import "github.com/go-playground/validator/v10"

//	{
//	  "message": "validation error",
//	  "code": "VALIDATION_ERROR",
//	  "request_id": "...",
//	  "errors": {
//	    "email": [
//	      {"code": "FIELD_REQUIRED", "param": ""}
//	    ],
//	    "password": [
//	      {"code": "FIELD_MIN_VALUE", "param": "8"}
//	    ],
//	    "status": [
//	      {"code": "ISSUE_STATUS_INVALID", "param": ""}
//	    ]
//	  }
//	}
func mapValidationErrorsIntoResponse(err validator.ValidationErrors) ValidationErrors {
	errors := make(ValidationErrors)
	for _, fieldErr := range err {
		code := getValidationCode(fieldErr.Tag())
		param := fieldErr.Param() // параметр (например, "8" для min=8)
		errors[fieldErr.Field()] = append(errors[fieldErr.Field()], ValidationError{
			Code:  code,
			Param: param,
		})
	}
	return errors
}
