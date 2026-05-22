package dto

// CSVColumnMapping maps contact field names to CSV column headers.
// An empty string means "not mapped / skip this field".
type CSVColumnMapping struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	MiddleName        string `json:"middle_name"`
	Nickname          string `json:"nickname"`
	Prefix            string `json:"prefix"`
	Suffix            string `json:"suffix"`
	Gender            string `json:"gender"`
	Birthday          string `json:"birthday"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	Company           string `json:"company"`
	JobTitle          string `json:"job_title"`
	Tags              string `json:"tags"`
	Groups            string `json:"groups"`
	Notes             string `json:"notes"`
	AddressStreet     string `json:"address_street"`
	AddressCity       string `json:"address_city"`
	AddressState      string `json:"address_state"`
	AddressPostalCode string `json:"address_postal_code"`
	AddressCountry    string `json:"address_country"`
}

type CSVImportResponse struct {
	ImportedContacts int      `json:"imported_contacts" example:"10"`
	SkippedCount     int      `json:"skipped_count"     example:"0"`
	Errors           []string `json:"errors,omitempty"`
}
