package handlers

import (
	"bytes"
	"net/http"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/naiba/bonds/internal/dto"
	"github.com/naiba/bonds/internal/middleware"
	"github.com/naiba/bonds/internal/models"
	"github.com/naiba/bonds/internal/services"
	"github.com/naiba/bonds/internal/utils"
	"github.com/naiba/bonds/pkg/avatar"
	"github.com/naiba/bonds/pkg/response"
	"github.com/sunshineplan/imgconv"
	"gorm.io/gorm"
)

var _ dto.VaultFileResponse

type AvatarHandler struct {
	db               *gorm.DB
	vaultFileService *services.VaultFileService
}

func NewAvatarHandler(db *gorm.DB, vaultFileService *services.VaultFileService) *AvatarHandler {
	return &AvatarHandler{db: db, vaultFileService: vaultFileService}
}

// GetAvatar godoc
//
//	@Summary		Get contact avatar
//	@Description	Return contact avatar image or generated initials
//	@Tags			contacts
//	@Produce		png
//	@Security		BearerAuth
//	@Param			vault_id	path		string	true	"Vault ID"
//	@Param			contact_id	path		string	true	"Contact ID"
//	@Success		200			{file}		file
//	@Failure		404			{object}	response.APIResponse
//	@Router			/vaults/{vault_id}/contacts/{contact_id}/avatar [get]
func (h *AvatarHandler) GetAvatar(c echo.Context) error {
	contactID := c.Param("contact_id")

	var contact models.Contact
	if err := h.db.First(&contact, "id = ?", contactID).Error; err != nil {
		return response.NotFound(c, "err.contact_not_found")
	}

	if contact.FileID != nil {
		var file models.File
		if err := h.db.First(&file, *contact.FileID).Error; err == nil {
			filePath := filepath.Join(h.vaultFileService.UploadDir(), file.UUID)
			return c.File(filePath)
		}
	}

	name := utils.BuildContactName(&contact)
	pngData := avatar.GenerateInitials(name, 128)

	return c.Blob(http.StatusOK, "image/png", pngData)
}

// UpdateAvatar godoc
//
//	@Summary		Update contact avatar
//	@Description	Upload a new avatar image for a contact
//	@Tags			contacts
//	@Accept			mpfd
//	@Produce		json
//	@Security		BearerAuth
//	@Param			vault_id	path		string	true	"Vault ID"
//	@Param			contact_id	path		string	true	"Contact ID"
//	@Param			file		formData	file	true	"Avatar image file"
//	@Success		200			{object}	response.APIResponse{data=dto.VaultFileResponse}
//	@Failure		400			{object}	response.APIResponse
//	@Failure		404			{object}	response.APIResponse
//	@Failure		500			{object}	response.APIResponse
//	@Router			/vaults/{vault_id}/contacts/{contact_id}/avatar [put]
func (h *AvatarHandler) UpdateAvatar(c echo.Context) error {
	contactID := c.Param("contact_id")
	vaultID := c.Param("vault_id")

	var contact models.Contact
	if err := h.db.Where("id = ? AND vault_id = ?", contactID, vaultID).First(&contact).Error; err != nil {
		return response.NotFound(c, "err.contact_not_found")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "err.file_required", nil)
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" && mimeType != "image/webp" {
		return response.BadRequest(c, "err.file_type_not_allowed", nil)
	}

	authorID := middleware.GetUserID(c)

	// If the mime type is not jpeg, we store a separate version of the avatar in the jpeg format for compatibility with
	// Apple Contacts, which only supports jpeg avatars. We can still store the original file as well, so that we can display
	// it in the app and allow users to download it in the original format if they want.
	// In the future it might also be worth scaling the avatar to a reasonable size if it's too large, but for now we'll just
	// convert it to jpeg without resizing.
	if mimeType != "image/jpeg" {
		// Open the uploaded file for reading
		src, err := fileHeader.Open()
		if err != nil {
			return response.InternalError(c, "err.failed_to_read_file")
		}
		defer src.Close()
		// Now, convert to jpeg as Apple Contacts only supports jpeg avatars.
		img, err := imgconv.Decode(src)
		if err != nil {
			return response.InternalError(c, "err.failed_to_decode_file")
		}
		var jpegAvatarData bytes.Buffer
		if err := imgconv.Write(&jpegAvatarData, img, &imgconv.FormatOption{
			Format: imgconv.JPEG,
		}); err != nil {
			return response.InternalError(c, "err.failed_to_convert_file")
		}
		// Upload the jpeg version as well, with a different filename and type.
		jpegAvatar, err := h.vaultFileService.Upload(vaultID, contactID, authorID, "avatar", fileHeader.Filename+".jpg", "image/jpeg", int64(jpegAvatarData.Len()), &jpegAvatarData)
		if err != nil {
			return response.InternalError(c, "err.failed_to_upload_file")
		}
		contact.OptimizedAvatarFileID = &jpegAvatar.ID
	}

	src, err := fileHeader.Open()
	if err != nil {
		return response.InternalError(c, "err.failed_to_read_file")
	}
	defer src.Close()
	file, err := h.vaultFileService.Upload(vaultID, contactID, authorID, "avatar", fileHeader.Filename, mimeType, fileHeader.Size, src)
	if err != nil {
		return response.InternalError(c, "err.failed_to_upload_file")
	}

	contact.FileID = &file.ID
	if contact.OptimizedAvatarFileID == nil {
		// This covers the case where the uploaded file is already a jpeg, so we don't need to create a separate jpeg version. In this case we can just use the original file as the avatar.
		contact.OptimizedAvatarFileID = &file.ID
	}

	// Ensure that clients receive the updated avatar.
	now := time.Now()
	contact.LastUpdatedAt = &now

	if err := h.db.Save(&contact).Error; err != nil {
		return response.InternalError(c, "err.failed_to_update_avatar")
	}

	return response.OK(c, file)
}

// DeleteAvatar godoc
//
//	@Summary		Delete contact avatar
//	@Description	Remove the avatar from a contact
//	@Tags			contacts
//	@Produce		json
//	@Security		BearerAuth
//	@Param			vault_id	path	string	true	"Vault ID"
//	@Param			contact_id	path	string	true	"Contact ID"
//	@Success		204			"No Content"
//	@Failure		404			{object}	response.APIResponse
//	@Failure		500			{object}	response.APIResponse
//	@Router			/vaults/{vault_id}/contacts/{contact_id}/avatar [delete]
func (h *AvatarHandler) DeleteAvatar(c echo.Context) error {
	contactID := c.Param("contact_id")
	vaultID := c.Param("vault_id")

	var contact models.Contact
	if err := h.db.Where("id = ? AND vault_id = ?", contactID, vaultID).First(&contact).Error; err != nil {
		return response.NotFound(c, "err.contact_not_found")
	}

	// TODO: Shouldn't this delete the avatar (and now also the jpeg version) from storage as well?

	contact.FileID = nil
	contact.OptimizedAvatarFileID = nil
	if err := h.db.Save(&contact).Error; err != nil {
		return response.InternalError(c, "err.failed_to_delete_avatar")
	}

	return response.NoContent(c)
}
