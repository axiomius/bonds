package mcp

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/naiba/bonds/internal/models"
	"gorm.io/gorm"
)

type ResourceFetcher struct {
	db           *gorm.DB
	vaultService VaultAccessChecker
}

type FetchResourceArgs struct {
	URI string `json:"uri"`
}

func NewResourceFetcher(db *gorm.DB, vaultService VaultAccessChecker) *ResourceFetcher {
	return &ResourceFetcher{db: db, vaultService: vaultService}
}

func (f *ResourceFetcher) Fetch(userID string, args FetchResourceArgs) (interface{}, error) {
	parsed, err := parseBondsURI(args.URI)
	if err != nil {
		return nil, err
	}
	switch parsed.kind {
	case "vault":
		var vault models.Vault
		if err := f.db.First(&vault, "id = ?", parsed.id).Error; err != nil {
			return nil, err
		}
		if err := f.vaultService.CheckUserVaultAccess(userID, vault.ID, models.PermissionViewer); err != nil {
			return nil, err
		}
		return vault, nil
	case "contact":
		var contact models.Contact
		if err := f.db.Preload("ContactInformations").Preload("Notes").Preload("ImportantDates").Where("listed = ?", true).First(&contact, "id = ?", parsed.id).Error; err != nil {
			return nil, err
		}
		if err := f.vaultService.CheckUserVaultAccess(userID, contact.VaultID, models.PermissionViewer); err != nil {
			return nil, err
		}
		return contact, nil
	case "note":
		var note models.Note
		if err := f.db.Preload("Contact").First(&note, "id = ?", parsed.id).Error; err != nil {
			return nil, err
		}
		if !note.Contact.Listed {
			return nil, gorm.ErrRecordNotFound
		}
		if err := f.vaultService.CheckUserVaultAccess(userID, note.VaultID, models.PermissionViewer); err != nil {
			return nil, err
		}
		return note, nil
	case "task":
		var task models.ContactTask
		if err := f.db.First(&task, "id = ?", parsed.id).Error; err != nil {
			return nil, err
		}
		if err := f.vaultService.CheckUserVaultAccess(userID, task.VaultID, models.PermissionViewer); err != nil {
			return nil, err
		}
		visible, err := f.taskVisible(task.ID, task.VaultID)
		if err != nil {
			return nil, err
		}
		if !visible {
			return nil, gorm.ErrRecordNotFound
		}
		return task, nil
	case "reminder":
		var reminder models.ContactReminder
		if err := f.db.Preload("Contact").First(&reminder, "id = ?", parsed.id).Error; err != nil {
			return nil, err
		}
		if !reminder.Contact.Listed {
			return nil, gorm.ErrRecordNotFound
		}
		if err := f.vaultService.CheckUserVaultAccess(userID, reminder.Contact.VaultID, models.PermissionViewer); err != nil {
			return nil, err
		}
		return reminder, nil
	case "important-date":
		var date models.ContactImportantDate
		if err := f.db.Preload("Contact").First(&date, "id = ?", parsed.id).Error; err != nil {
			return nil, err
		}
		if !date.Contact.Listed {
			return nil, gorm.ErrRecordNotFound
		}
		if err := f.vaultService.CheckUserVaultAccess(userID, date.Contact.VaultID, models.PermissionViewer); err != nil {
			return nil, err
		}
		return date, nil
	}
	return nil, fmt.Errorf("unsupported resource kind: %s", parsed.kind)
}

func (f *ResourceFetcher) taskVisible(taskID uint, vaultID string) (bool, error) {
	var assignments int64
	if err := f.db.Table("task_contacts").Where("contact_task_id = ?", taskID).Count(&assignments).Error; err != nil {
		return false, err
	}
	if assignments == 0 {
		return true, nil
	}
	var visibleAssignments int64
	err := f.db.Table("task_contacts").
		Joins("JOIN contacts ON contacts.id = task_contacts.contact_id").
		Where("task_contacts.contact_task_id = ? AND contacts.vault_id = ? AND contacts.listed = ? AND contacts.deleted_at IS NULL", taskID, vaultID, true).
		Count(&visibleAssignments).Error
	return visibleAssignments > 0, err
}

type bondsURI struct {
	kind string
	id   string
}

func parseBondsURI(raw string) (bondsURI, error) {
	uri, err := url.Parse(raw)
	if err != nil {
		return bondsURI{}, err
	}
	if uri.Scheme != "bonds" {
		return bondsURI{}, fmt.Errorf("resource URI must use bonds:// scheme")
	}
	kind := uri.Host
	id := strings.Trim(uri.Path, "/")
	if kind == "" || id == "" {
		return bondsURI{}, fmt.Errorf("resource URI must include kind and id")
	}
	return bondsURI{kind: kind, id: id}, nil
}
