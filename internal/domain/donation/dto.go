package donation

type CreateRequest struct {
	DonaturID          int     `json:"donatur_id" validate:"required,min=1"`
	DonaturGroupID     int     `json:"donatur_group_id" validate:"required,min=1"`
	AreaID             int     `json:"area_id" validate:"required,min=1"`
	DonationCategoryID int     `json:"donation_category_id" validate:"required,min=1"`
	Currency           string  `json:"currency" validate:"required,max=10"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	OtherItems         string  `json:"other_items" validate:"omitempty,max=500"`
}

// ImportDonationRequest is a single donation row parsed from the uploaded Excel
// sheet. Values are kept as raw strings here because lookups, normalisation and
// amount parsing happen during the import so that one bad value only fails its
// own row.
type ImportDonationRequest struct {
	Row          int    `json:"row"`             // 1-based Excel row number (incl. header)
	DonaturID    string `json:"donatur_id"`      // column B - "ID Donatur"
	DonaturName  string `json:"donatur_name"`    // column C - "Nama Donatur"
	CategoryName string `json:"jenis_sumbangan"` // column D - "Jenis Sumbangan"
	Note         string `json:"catatan"`         // column E - "Catatan"
	Amount       string `json:"jumlah"`          // column F - "Jumlah"
}

// DonationImportError describes a row that could not be imported (validation
// failure) or that was skipped (duplicate warning).
type DonationImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportOptions carries request-level context into the import service.
type ImportOptions struct {
	Filename   string
	UploadedBy *int
	DryRun     bool
}

// DonationImportResult is the summary returned after processing the whole file.
// Rows are processed independently: a single failure never affects the others.
type DonationImportResult struct {
	BatchID        string                `json:"batch_id,omitempty"`
	DryRun         bool                  `json:"dry_run"`
	Status         string                `json:"status"`
	Filename       string                `json:"filename,omitempty"`
	KomisariatID   string                `json:"komisariat_id,omitempty"`
	KomisariatName string                `json:"komisariat_name,omitempty"`
	Period         string                `json:"period,omitempty"`
	TotalRows      int                   `json:"total_rows"`
	SuccessRows    int                   `json:"success_rows"`
	FailedRows     int                   `json:"failed_rows"`
	SkippedRows    int                   `json:"skipped_rows"`
	Errors         []DonationImportError `json:"errors"`
	Warnings       []DonationImportError `json:"warnings"`
}

type UpdateRequest struct {
	DonaturID          int     `json:"donatur_id" validate:"required,min=1"`
	DonaturGroupID     int     `json:"donatur_group_id" validate:"required,min=1"`
	AreaID             int     `json:"area_id" validate:"required,min=1"`
	DonationCategoryID int     `json:"donation_category_id" validate:"required,min=1"`
	Currency           string  `json:"currency" validate:"required,max=10"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	OtherItems         string  `json:"other_items" validate:"omitempty,max=500"`
}
