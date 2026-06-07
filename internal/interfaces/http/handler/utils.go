package handler

import (
	"fmt"
	"net/http"

	domain_errors "github.com/pelDev/health-connect/internal/domain/errors"
	"github.com/pelDev/health-connect/internal/utils"
)

func handleError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case domain_errors.AppError:
		utils.RespondWithJson(
			w,
			e.GetHTTPCode(),
			map[string]any{
				"message": e.GetDetail(),
				"error":   e.Error(),
				"code":    e.GetCode(),
			},
		)
	default:
		utils.RespondWithJson(
			w,
			http.StatusInternalServerError,
			map[string]any{
				"message": fmt.Sprintf("error occurred: %s", e.Error()),
			},
		)
	}
}
