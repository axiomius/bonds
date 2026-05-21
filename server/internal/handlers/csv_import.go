package handlers

import (
	"bytes"
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/naiba/bonds/internal/dto"
	"github.com/naiba/bonds/internal/middleware"
	"github.com/naiba/bonds/internal/services"
	"github.com/naiba/bonds/pkg/response"
)

var _ dto.CSVImportResponse

type CSVImportHandler struct {
	svc *services.CSVImportService
}

func NewCSVImportHandler(svc *services.CSVImportService) *CSVImportHandler {
	return &CSVImportHandler{svc: svc}
}

// Import godoc
//
//	@Summary		Import contacts from a CSV file
//	@Description	Import contacts from a CSV file with a user-defined column mapping
//	@Tags			Vault Settings
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			vault_id	path		string	true	"Vault ID"
//	@Param			file		formData	file	true	"CSV file"
//	@Param			mapping		formData	string	true	"JSON column mapping"
//	@Success		200			{object}	response.APIResponse{data=dto.CSVImportResponse}
//	@Failure		400			{object}	response.APIResponse
//	@Failure		500			{object}	response.APIResponse
//	@Router			/vaults/{vault_id}/settings/import/csv [post]
func (h *CSVImportHandler) Import(c echo.Context) error {
	vaultID := c.Param("vault_id")
	userID := middleware.GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "err.file_required", nil)
	}

	src, err := file.Open()
	if err != nil {
		return response.InternalError(c, "err.failed_to_read_file")
	}
	defer src.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(src); err != nil {
		return response.InternalError(c, "err.failed_to_read_file")
	}

	mappingJSON := c.FormValue("mapping")
	var mapping dto.CSVColumnMapping
	if mappingJSON != "" {
		if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
			return response.BadRequest(c, "err.invalid_csv_mapping", nil)
		}
	}

	result, err := h.svc.Import(vaultID, userID, buf.Bytes(), mapping)
	if err != nil {
		return response.InternalError(c, "err.failed_to_import_csv")
	}

	return response.OK(c, result)
}
