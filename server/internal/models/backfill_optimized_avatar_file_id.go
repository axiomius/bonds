package models

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/sunshineplan/imgconv"
	"gorm.io/gorm"
)

// BackfillOptimizedAvatarFileID populates the optimized_avatar_file_id column for existing contacts.
// If an avatar is present and it is already in the jpeg format, it will be used as the optimized avatar.
// Otherwise, a new optimized avatar will be generated from the original file.
func BackfillOptimizedAvatarFileID(db *gorm.DB, uploadDir *string) error {
	return db.Transaction(func(tx *gorm.DB) error {

		var contacts []Contact

		if err := tx.
			Where("optimized_avatar_file_id IS NULL AND file_id IS NOT NULL").
			Find(&contacts).Error; err != nil {
			return err
		}

		for _, contact := range contacts {
			var file File
			if err := tx.First(&file, "id = ?", *contact.FileID).Error; err != nil {
				continue
			}

			// If already JPEG → just link it
			if file.MimeType == "image/jpeg" {
				if err := tx.Model(&Contact{}).
					Where("id = ?", contact.ID).
					Update("optimized_avatar_file_id", contact.FileID).Error; err != nil {
					return err
				}
				continue
			}

			filePath := filepath.Join(*uploadDir, file.UUID)

			src, err := os.Open(filePath)
			if err != nil {
				continue
			}

			img, err := imgconv.Decode(src)
			src.Close()

			if err != nil {
				continue
			}

			var jpegAvatarData bytes.Buffer
			if err := imgconv.Write(&jpegAvatarData, img, &imgconv.FormatOption{
				Format: imgconv.JPEG,
			}); err != nil {
				continue
			}

			fileJPEG := File{
				VaultID:  file.VaultID,
				UUID:     uuid.New().String(),
				Name:     file.Name + ".jpg",
				MimeType: "image/jpeg",
				Type:     file.Type,
				Size:     jpegAvatarData.Len(),
			}

			if err := tx.Create(&fileJPEG).Error; err != nil {
				continue
			}

			if err := tx.Model(&Contact{}).
				Where("id = ?", contact.ID).
				Update("optimized_avatar_file_id", fileJPEG.ID).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
