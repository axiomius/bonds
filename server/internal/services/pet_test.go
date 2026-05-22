package services

import (
	"testing"

	"github.com/naiba/bonds/internal/dto"
	"github.com/naiba/bonds/internal/testutil"
)

func setupPetTest(t *testing.T) (*PetService, string, string) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	cfg := testutil.TestJWTConfig()
	authSvc := NewAuthService(db, cfg)
	vaultSvc := NewVaultService(db)

	resp, err := authSvc.Register(dto.RegisterRequest{
		FirstName: "Test",
		LastName:  "User",
		Email:     "pet-test@example.com",
		Password:  "password123",
	}, "en")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	vault, err := vaultSvc.CreateVault(resp.User.AccountID, resp.User.ID, dto.CreateVaultRequest{Name: "Test Vault"}, "en")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	contactSvc := NewContactService(db)
	contact, err := contactSvc.CreateContact(vault.ID, resp.User.ID, dto.CreateContactRequest{FirstName: "John"})
	if err != nil {
		t.Fatalf("CreateContact failed: %v", err)
	}

	return NewPetService(db), contact.ID, vault.ID
}

func setupPetTestWithDB(t *testing.T) (*PetService, *AuthService, *VaultService, string, string, string) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	cfg := testutil.TestJWTConfig()
	authSvc := NewAuthService(db, cfg)
	vaultSvc := NewVaultService(db)

	resp, err := authSvc.Register(dto.RegisterRequest{
		FirstName: "Test",
		LastName:  "User",
		Email:     "pet-test@example.com",
		Password:  "password123",
	}, "en")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	vault, err := vaultSvc.CreateVault(resp.User.AccountID, resp.User.ID, dto.CreateVaultRequest{Name: "Test Vault"}, "en")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}

	contactSvc := NewContactService(db)
	contact, err := contactSvc.CreateContact(vault.ID, resp.User.ID, dto.CreateContactRequest{FirstName: "John"})
	if err != nil {
		t.Fatalf("CreateContact failed: %v", err)
	}

	return NewPetService(db), authSvc, vaultSvc, contact.ID, vault.ID, resp.User.AccountID
}

func TestCreatePet(t *testing.T) {
	svc, contactID, vaultID := setupPetTest(t)

	pet, err := svc.Create(contactID, vaultID, dto.CreatePetRequest{
		PetCategoryID: 1,
		Name:          "Buddy",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if pet.Name != "Buddy" {
		t.Errorf("Expected name 'Buddy', got '%s'", pet.Name)
	}
	if pet.PetCategoryID != 1 {
		t.Errorf("Expected pet_category_id 1, got %d", pet.PetCategoryID)
	}
	if pet.ContactID != contactID {
		t.Errorf("Expected contact_id '%s', got '%s'", contactID, pet.ContactID)
	}
	if pet.ID == 0 {
		t.Error("Expected pet ID to be non-zero")
	}
	if pet.PetCategoryName != "Dog" {
		t.Errorf("Expected pet_category_name 'Dog', got '%s'", pet.PetCategoryName)
	}
}

func TestListPets(t *testing.T) {
	svc, contactID, vaultID := setupPetTest(t)

	_, err := svc.Create(contactID, vaultID, dto.CreatePetRequest{PetCategoryID: 1, Name: "Pet 1"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	_, err = svc.Create(contactID, vaultID, dto.CreatePetRequest{PetCategoryID: 2, Name: "Pet 2"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	pets, err := svc.List(contactID, vaultID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(pets) != 2 {
		t.Errorf("Expected 2 pets, got %d", len(pets))
	}
	if pets[0].PetCategoryName != "Cat" && pets[0].PetCategoryName != "Dog" {
		t.Errorf("Expected pet category name on list response, got '%s'", pets[0].PetCategoryName)
	}
}

func TestListCategories(t *testing.T) {
	svc, _, _, _, _, accountID := setupPetTestWithDB(t)

	categories, err := svc.ListCategories(accountID)
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if len(categories) == 0 {
		t.Fatal("Expected seeded pet categories")
	}
	if categories[0].Name != "Dog" {
		t.Errorf("Expected first category name 'Dog', got '%s'", categories[0].Name)
	}
}

func TestUpdatePet(t *testing.T) {
	svc, contactID, vaultID := setupPetTest(t)

	created, err := svc.Create(contactID, vaultID, dto.CreatePetRequest{PetCategoryID: 1, Name: "Original"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	updated, err := svc.Update(created.ID, contactID, vaultID, dto.UpdatePetRequest{
		PetCategoryID: 2,
		Name:          "Updated",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got '%s'", updated.Name)
	}
	if updated.PetCategoryID != 2 {
		t.Errorf("Expected pet_category_id 2, got %d", updated.PetCategoryID)
	}
	if updated.PetCategoryName != "Cat" {
		t.Errorf("Expected pet_category_name 'Cat', got '%s'", updated.PetCategoryName)
	}
}

func TestCreatePetRejectsCrossAccountCategory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.TestJWTConfig()
	authSvc := NewAuthService(db, cfg)
	vaultSvc := NewVaultService(db)

	first, err := authSvc.Register(dto.RegisterRequest{FirstName: "One", LastName: "User", Email: "one@example.com", Password: "password123"}, "en")
	if err != nil {
		t.Fatalf("Register first account failed: %v", err)
	}
	second, err := authSvc.Register(dto.RegisterRequest{FirstName: "Two", LastName: "User", Email: "two@example.com", Password: "password123"}, "en")
	if err != nil {
		t.Fatalf("Register second account failed: %v", err)
	}

	vault, err := vaultSvc.CreateVault(first.User.AccountID, first.User.ID, dto.CreateVaultRequest{Name: "Vault"}, "en")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}
	contactSvc := NewContactService(db)
	contact, err := contactSvc.CreateContact(vault.ID, first.User.ID, dto.CreateContactRequest{FirstName: "John"})
	if err != nil {
		t.Fatalf("CreateContact failed: %v", err)
	}
	var otherCategory struct{ ID uint }
	if err := db.Table("pet_categories").Where("account_id = ?", second.User.AccountID).Order("id ASC").Select("id").First(&otherCategory).Error; err != nil {
		t.Fatalf("failed to load second account pet category: %v", err)
	}

	petSvc := NewPetService(db)
	_, err = petSvc.Create(contact.ID, vault.ID, dto.CreatePetRequest{PetCategoryID: 1, Name: "Buddy"})
	if err != nil {
		t.Fatalf("Create pet with own category failed: %v", err)
	}

	_, err = petSvc.Create(contact.ID, vault.ID, dto.CreatePetRequest{PetCategoryID: 9999, Name: "Missing"})
	if err != ErrPetCategoryNotFound {
		t.Fatalf("Expected ErrPetCategoryNotFound for missing category, got %v", err)
	}

	if first.User.AccountID == second.User.AccountID {
		t.Fatal("expected distinct accounts")
	}
	_, err = petSvc.Create(contact.ID, vault.ID, dto.CreatePetRequest{PetCategoryID: otherCategory.ID, Name: "Cross-account"})
	if err != ErrPetCategoryNotFound {
		t.Fatalf("Expected ErrPetCategoryNotFound for cross-account category, got %v", err)
	}
}

func TestUpdatePetRejectsCrossAccountCategory(t *testing.T) {
	db := testutil.SetupTestDB(t)
	cfg := testutil.TestJWTConfig()
	authSvc := NewAuthService(db, cfg)
	vaultSvc := NewVaultService(db)

	first, err := authSvc.Register(dto.RegisterRequest{FirstName: "One", LastName: "User", Email: "update-one@example.com", Password: "password123"}, "en")
	if err != nil {
		t.Fatalf("Register first account failed: %v", err)
	}
	second, err := authSvc.Register(dto.RegisterRequest{FirstName: "Two", LastName: "User", Email: "update-two@example.com", Password: "password123"}, "en")
	if err != nil {
		t.Fatalf("Register second account failed: %v", err)
	}

	vault, err := vaultSvc.CreateVault(first.User.AccountID, first.User.ID, dto.CreateVaultRequest{Name: "Vault"}, "en")
	if err != nil {
		t.Fatalf("CreateVault failed: %v", err)
	}
	contactSvc := NewContactService(db)
	contact, err := contactSvc.CreateContact(vault.ID, first.User.ID, dto.CreateContactRequest{FirstName: "John"})
	if err != nil {
		t.Fatalf("CreateContact failed: %v", err)
	}
	var otherCategory struct{ ID uint }
	if err := db.Table("pet_categories").Where("account_id = ?", second.User.AccountID).Order("id ASC").Select("id").First(&otherCategory).Error; err != nil {
		t.Fatalf("failed to load second account pet category: %v", err)
	}

	petSvc := NewPetService(db)
	created, err := petSvc.Create(contact.ID, vault.ID, dto.CreatePetRequest{PetCategoryID: 1, Name: "Buddy"})
	if err != nil {
		t.Fatalf("Create pet failed: %v", err)
	}

	_, err = petSvc.Update(created.ID, contact.ID, vault.ID, dto.UpdatePetRequest{PetCategoryID: otherCategory.ID, Name: "Updated"})
	if err != ErrPetCategoryNotFound {
		t.Fatalf("Expected ErrPetCategoryNotFound for cross-account update category, got %v", err)
	}

	_, err = petSvc.Update(created.ID, contact.ID, vault.ID, dto.UpdatePetRequest{PetCategoryID: 1, Name: "Updated"})
	if err != nil {
		t.Fatalf("Expected own category to continue working on update, got %v", err)
	}
}

func TestDeletePet(t *testing.T) {
	svc, contactID, vaultID := setupPetTest(t)

	created, err := svc.Create(contactID, vaultID, dto.CreatePetRequest{PetCategoryID: 1, Name: "To delete"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if err := svc.Delete(created.ID, contactID, vaultID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	pets, err := svc.List(contactID, vaultID)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(pets) != 0 {
		t.Errorf("Expected 0 pets after delete, got %d", len(pets))
	}
}

func TestPetNotFound(t *testing.T) {
	svc, contactID, vaultID := setupPetTest(t)

	_, err := svc.Update(9999, contactID, vaultID, dto.UpdatePetRequest{PetCategoryID: 1, Name: "nope"})
	if err != ErrPetNotFound {
		t.Errorf("Expected ErrPetNotFound, got %v", err)
	}

	err = svc.Delete(9999, contactID, vaultID)
	if err != ErrPetNotFound {
		t.Errorf("Expected ErrPetNotFound, got %v", err)
	}
}
