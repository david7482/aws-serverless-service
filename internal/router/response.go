package router

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/david7482/aws-serverless-service/internal/domain"
)

type ErrorCategory string

const (
	ErrorCategoryParameter = ErrorCategory("PARAMETER_ERROR")
	ErrorCategoryResource  = ErrorCategory("RESOURCE_ERROR")
	ErrorCategoryInternal  = ErrorCategory("INTERNAL_ERROR")
	ErrorCategoryExternal  = ErrorCategory("EXTERNAL_ERROR")
	ErrorCategoryUnknown   = ErrorCategory("UNKNOWN_ERROR")
)

type ErrorMessage struct {
	Category ErrorCategory `json:"category"`
	Message  string        `json:"message"`
}

func respondWithJSON(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, payload)
}

func respondWithoutBody(c *gin.Context, code int) {
	c.Status(code)
}

func respondWithError(c *gin.Context, err error) {
	code, category, msg := parseError(err)

	ctx := c.Request.Context()
	zerolog.Ctx(ctx).Error().Err(err).Msg(msg)
	payload := ErrorMessage{
		Category: category,
		Message:  msg,
	}
	c.AbortWithStatusJSON(code, payload)
}

func parseError(err error) (int, ErrorCategory, string) {
	// Handle InternalError
	var internalProcessError domain.InternalError
	if valid := errors.As(err, &internalProcessError); valid {
		return http.StatusInternalServerError, ErrorCategoryInternal, internalProcessError.ClientMsg()
	}

	// Handle ExternalError
	var remoteProcessError domain.ExternalError
	if valid := errors.As(err, &remoteProcessError); valid {
		return remoteProcessError.StatusCode(), ErrorCategoryExternal, remoteProcessError.ClientMsg()
	}

	// Handle ResourceNotFoundError
	var resourceNotFoundError domain.ResourceNotFoundError
	if valid := errors.As(err, &resourceNotFoundError); valid {
		return http.StatusNotFound, ErrorCategoryResource, resourceNotFoundError.ClientMsg()
	}

	// Handle ParameterError
	var parameterError domain.ParameterError
	if valid := errors.As(err, &parameterError); valid {
		return http.StatusBadRequest, ErrorCategoryParameter, parameterError.ClientMsg()
	}

	// Return default status code and category
	return http.StatusInternalServerError, ErrorCategoryUnknown, "unknown internal error"
}
