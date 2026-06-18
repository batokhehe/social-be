package donation

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// mockDonationRepo is a function-field stub implementing Repository for tests.
type mockDonationRepo struct {
	donors     map[string]DonaturRef
	categories map[string]int
	existing   map[string]struct{}
	persistErr error
	logByBatch map[string]*ImportLog

	// history reads
	importLogs    []ImportLog
	importErrors  map[string][]ImportError
	batchDonation map[string][]Donation
	userNames     map[int]string

	persisted     []ImportedDonation
	persistedErrs []ImportError
	persistedLog  *ImportLog
	logs          []ImportLog
	deletedByID   []string
}

func (m *mockDonationRepo) Create(ctx context.Context, d *Donation) (*Donation, error) { return d, nil }
func (m *mockDonationRepo) GetAll(ctx context.Context) ([]Donation, error)             { return nil, nil }
func (m *mockDonationRepo) GetByID(ctx context.Context, id int) (*Donation, error)     { return nil, nil }
func (m *mockDonationRepo) Update(ctx context.Context, d *Donation) (*Donation, error) { return d, nil }
func (m *mockDonationRepo) Delete(ctx context.Context, id int) error                   { return nil }

func (m *mockDonationRepo) FindDonaturRefsByIDs(ctx context.Context, ids []string) (map[string]DonaturRef, error) {
	out := make(map[string]DonaturRef)
	for _, id := range ids {
		if ref, ok := m.donors[id]; ok {
			out[id] = ref
		}
	}
	return out, nil
}

func (m *mockDonationRepo) LoadCategoryIndex(ctx context.Context) (map[string]int, error) {
	if m.categories == nil {
		return map[string]int{}, nil
	}
	return m.categories, nil
}

func (m *mockDonationRepo) LoadExistingDedupKeys(ctx context.Context, donaturIDs []int) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	for k := range m.existing {
		out[k] = struct{}{}
	}
	return out, nil
}

// PersistImportLocked mirrors the real repo's behaviour: it re-checks
// duplicates (against existing + within-batch) and only inserts survivors.
func (m *mockDonationRepo) PersistImportLocked(ctx context.Context, lockKey int64, donorPKs []int, candidates []ImportCandidate, errorRows []ImportError, log *ImportLog) (PersistOutcome, error) {
	if m.persistErr != nil {
		return PersistOutcome{}, m.persistErr
	}
	var skipped []DonationImportError
	seen := make(map[string]struct{})
	for _, c := range candidates {
		if _, dup := m.existing[c.DedupKey]; dup {
			skipped = append(skipped, DonationImportError{Row: c.Row, Message: "Donation already exists"})
			continue
		}
		if _, dup := seen[c.DedupKey]; dup {
			skipped = append(skipped, DonationImportError{Row: c.Row, Message: "Donation already exists"})
			continue
		}
		seen[c.DedupKey] = struct{}{}
		m.persisted = append(m.persisted, c.Donation)
	}
	m.persistedErrs = append(m.persistedErrs, errorRows...)
	log.SuccessRows = len(m.persisted)
	log.SkippedRows = len(skipped)
	m.persistedLog = log
	return PersistOutcome{Inserted: len(m.persisted), Skipped: skipped}, nil
}

func (m *mockDonationRepo) GetImportLogs(ctx context.Context, filters query.Filters, page pagination.Query, sort ImportHistorySort) ([]ImportLog, int64, error) {
	return m.importLogs, int64(len(m.importLogs)), nil
}

func (m *mockDonationRepo) GetImportErrors(ctx context.Context, batchID string) ([]ImportError, error) {
	return m.importErrors[batchID], nil
}

func (m *mockDonationRepo) GetDonationsByBatchID(ctx context.Context, batchID string) ([]Donation, error) {
	return m.batchDonation[batchID], nil
}

func (m *mockDonationRepo) GetUserNames(ctx context.Context, ids []int) (map[int]string, error) {
	out := make(map[int]string)
	for _, id := range ids {
		if n, ok := m.userNames[id]; ok {
			out[id] = n
		}
	}
	return out, nil
}

func (m *mockDonationRepo) CreateImportLog(ctx context.Context, log *ImportLog) error {
	m.logs = append(m.logs, *log)
	return nil
}

func (m *mockDonationRepo) GetImportLogByBatchID(ctx context.Context, batchID string) (*ImportLog, error) {
	if m.logByBatch == nil {
		return &ImportLog{BatchID: batchID, Status: StatusCompleted}, nil
	}
	return m.logByBatch[batchID], nil
}

func (m *mockDonationRepo) SoftDeleteByBatchID(ctx context.Context, batchID string, deletedBy *int) (int64, error) {
	m.deletedByID = append(m.deletedByID, batchID)
	return 3, nil
}

// buildXLSX creates an in-memory .xlsx: 3 metadata rows, a blank line, the
// column header row, then the data rows.
func buildXLSX(t *testing.T, dataRows [][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	setRow := func(ref string, vals []string) {
		cells := make([]interface{}, len(vals))
		for i, v := range vals {
			cells[i] = v
		}
		if err := f.SetSheetRow(sheet, ref, &cells); err != nil {
			t.Fatalf("set row %s: %v", ref, err)
		}
	}

	setRow("A1", []string{"ID Komisariat", "WG58522-1"})
	setRow("A2", []string{"Nama Komisariat", "YULIANTI"})
	setRow("A3", []string{"Periode", "2025"})
	setRow("A5", []string{"No", "ID Donatur", "Nama Donatur", "Jenis Sumbangan", "Catatan", "Jumlah"})
	for i, row := range dataRows {
		ref, _ := excelize.CoordinatesToCellName(1, i+6)
		setRow(ref, row)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write xlsx: %v", err)
	}
	return buf.Bytes()
}

func baseRepo() *mockDonationRepo {
	group := 7
	return &mockDonationRepo{
		donors: map[string]DonaturRef{
			"D2053534": {DonaturID: "D2053534", ID: 10, DonaturGroupID: &group},
			"D0000001": {DonaturID: "D0000001", ID: 11},
		},
		categories: map[string]int{"amal": 3, "pembangunan": 4},
		existing:   map[string]struct{}{},
	}
}

func TestImport_MixedRows(t *testing.T) {
	repo := baseRepo()
	svc := NewImportService(repo)

	xlsx := buildXLSX(t, [][]string{
		{"1", "D2053534", "Adrian", "Amal", "Desember 2025", "20.000"}, // ok
		{"2", "D0000001", "Budi", "AMAL", "", "1.899.000"},             // ok, no group, empty note
		{"3", "D999999", "Ghost", "Amal", "x", "10.000"},               // donor missing
		{"4", "D2053534", "Adrian", "ABC", "y", "5.000"},               // category missing
		{"5", "", "NoID", "Amal", "z", "1.000"},                        // missing ID Donatur
		{"6", "D2053534", "Adrian", "", "z", "1.000"},                  // missing category
	})

	res, err := svc.Import(context.Background(), bytes.NewReader(xlsx), ImportOptions{Filename: "d.xlsx"})
	if err != nil {
		t.Fatalf("Import returned error: %v", err)
	}

	if res.TotalRows != 6 || res.SuccessRows != 2 || res.FailedRows != 4 {
		t.Fatalf("got total=%d success=%d failed=%d, want 6/2/4", res.TotalRows, res.SuccessRows, res.FailedRows)
	}
	if res.KomisariatID != "WG58522-1" || res.KomisariatName != "YULIANTI" || res.Period != "2025" {
		t.Errorf("header metadata not parsed: %+v", res)
	}
	if res.Status != StatusCompletedWithError {
		t.Errorf("status = %q, want %q", res.Status, StatusCompletedWithError)
	}
	if res.BatchID == "" {
		t.Error("expected batch id on real import")
	}
	if len(repo.persisted) != 2 {
		t.Fatalf("persisted = %d, want 2", len(repo.persisted))
	}
	if repo.persisted[0].Amount != 20000 || repo.persisted[1].Amount != 1899000 {
		t.Errorf("amounts = %v / %v", repo.persisted[0].Amount, repo.persisted[1].Amount)
	}
	if repo.persisted[0].DonaturGroupID == nil || *repo.persisted[0].DonaturGroupID != 7 {
		t.Error("row1 group not enriched from donor")
	}
	if repo.persisted[1].DonaturGroupID != nil {
		t.Error("row2 group should be nil")
	}
	if repo.persisted[0].ImportBatchID != res.BatchID {
		t.Error("donation not tagged with batch id")
	}
	if repo.persistedLog == nil || repo.persistedLog.SuccessRows != 2 {
		t.Errorf("import log not persisted correctly: %+v", repo.persistedLog)
	}
}

func TestImport_DuplicateInFileAndDB(t *testing.T) {
	repo := baseRepo()
	// Pre-existing identical donation: donor 10, category 3, note "note", 10000.
	repo.existing[dedupKey(10, 3, "note", 10000)] = struct{}{}
	svc := NewImportService(repo)

	xlsx := buildXLSX(t, [][]string{
		{"1", "D2053534", "Adrian", "Amal", "note", "10.000"}, // dup vs DB -> skip
		{"2", "D2053534", "Adrian", "Amal", "fresh", "5.000"}, // ok
		{"3", "D2053534", "Adrian", "Amal", "fresh", "5.000"}, // dup within file -> skip
	})

	res, err := svc.Import(context.Background(), bytes.NewReader(xlsx), ImportOptions{})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}
	if res.SuccessRows != 1 || res.SkippedRows != 2 || res.FailedRows != 0 {
		t.Errorf("got success=%d skipped=%d failed=%d, want 1/2/0", res.SuccessRows, res.SkippedRows, res.FailedRows)
	}
	if len(res.Warnings) != 2 {
		t.Errorf("expected 2 duplicate warnings, got %d", len(res.Warnings))
	}
}

func TestImport_DryRunWritesNothing(t *testing.T) {
	repo := baseRepo()
	svc := NewImportService(repo)
	xlsx := buildXLSX(t, [][]string{{"1", "D2053534", "A", "Amal", "n", "10.000"}})

	res, err := svc.Import(context.Background(), bytes.NewReader(xlsx), ImportOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}
	if res.Status != StatusDryRun || !res.DryRun {
		t.Errorf("expected dry_run status, got %q", res.Status)
	}
	if res.SuccessRows != 1 {
		t.Errorf("dry run should still report success=1, got %d", res.SuccessRows)
	}
	if res.BatchID != "" {
		t.Errorf("dry run should not assign a batch id")
	}
	if len(repo.persisted) != 0 || repo.persistedLog != nil {
		t.Errorf("dry run must not write any data")
	}
}

func TestImport_PersistFailureRecordsFailedLog(t *testing.T) {
	repo := baseRepo()
	repo.persistErr = context.DeadlineExceeded
	svc := NewImportService(repo)
	xlsx := buildXLSX(t, [][]string{{"1", "D2053534", "A", "Amal", "n", "10.000"}})

	_, err := svc.Import(context.Background(), bytes.NewReader(xlsx), ImportOptions{})
	if err == nil {
		t.Fatal("expected error when persist fails")
	}
	if len(repo.logs) != 1 || repo.logs[0].Status != StatusFailed {
		t.Errorf("expected a failed audit log, got %+v", repo.logs)
	}
}

func TestImport_InvalidTemplateRejected(t *testing.T) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	cells := []interface{}{"No", "Donor Code", "Name", "Type", "Memo", "Total"} // wrong headers
	_ = f.SetSheetRow(sheet, "A1", &cells)
	var buf bytes.Buffer
	_ = f.Write(&buf)

	svc := NewImportService(baseRepo())
	_, err := svc.Import(context.Background(), bytes.NewReader(buf.Bytes()), ImportOptions{})
	if err == nil || !strings.Contains(err.Error(), "invalid template") {
		t.Fatalf("expected invalid template error, got %v", err)
	}
}

func TestRollbackBatch(t *testing.T) {
	repo := baseRepo()
	svc := NewImportService(repo)
	actor := 99
	n, err := svc.RollbackBatch(context.Background(), "batch-123", &actor)
	if err != nil || n != 3 {
		t.Fatalf("rollback = %d, %v; want 3, nil", n, err)
	}
	if len(repo.deletedByID) != 1 || repo.deletedByID[0] != "batch-123" {
		t.Errorf("rollback did not target the batch id")
	}
}

func TestRollback_NotFound(t *testing.T) {
	repo := baseRepo()
	repo.logByBatch = map[string]*ImportLog{} // batch absent -> nil log
	svc := NewImportService(repo)
	actor := 1
	_, err := svc.RollbackBatch(context.Background(), "missing", &actor)
	if err != ErrBatchNotFound {
		t.Fatalf("got %v, want ErrBatchNotFound", err)
	}
	if len(repo.deletedByID) != 0 {
		t.Error("must not delete anything for an unknown batch")
	}
}

func TestRollback_AlreadyRolledBack(t *testing.T) {
	repo := baseRepo()
	repo.logByBatch = map[string]*ImportLog{"b1": {BatchID: "b1", Status: StatusRolledBack}}
	svc := NewImportService(repo)
	actor := 1
	_, err := svc.RollbackBatch(context.Background(), "b1", &actor)
	if err != ErrBatchAlreadyRolledBk {
		t.Fatalf("got %v, want ErrBatchAlreadyRolledBk", err)
	}
	if len(repo.deletedByID) != 0 {
		t.Error("must not re-delete an already rolled-back batch")
	}
}

func TestFileLockKey_StableAndDistinct(t *testing.T) {
	a := fileLockKey(headerMeta{KomisariatID: "WG58522-1", Period: "2025"}, "d.xlsx")
	b := fileLockKey(headerMeta{KomisariatID: "WG58522-1", Period: "2025"}, "d.xlsx")
	c := fileLockKey(headerMeta{KomisariatID: "WG58522-2", Period: "2025"}, "d.xlsx")
	if a != b {
		t.Error("same file identity must produce the same lock key")
	}
	if a == c {
		t.Error("different files must produce different lock keys")
	}
}

func TestParseAmount(t *testing.T) {
	cases := []struct {
		in      string
		want    float64
		wantErr bool
	}{
		{"20.000", 20000, false},
		{"1.899.000", 1899000, false},
		{"10.000.000", 10000000, false},
		{"Rp 500.000", 500000, false},
		{"20000", 20000, false},
		{"", 0, false},
		{"   ", 0, false},
		{"abc", 0, true},
	}
	for _, c := range cases {
		got, err := parseAmount(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseAmount(%q) expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseAmount(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseAmount(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRouteRegistrationNoConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("router panicked registering routes: %v", r)
		}
	}()
	r := gin.New()
	g := r.Group("/donations")
	g.GET("", func(*gin.Context) {})
	g.GET(":id", func(*gin.Context) {})
	g.POST("", func(*gin.Context) {})
	g.POST("import", func(*gin.Context) {})
	g.DELETE("import/:batch_id", func(*gin.Context) {})
	g.GET("imports", func(*gin.Context) {})
	g.GET("imports/:batch_id", func(*gin.Context) {})
	g.GET("imports/:batch_id/errors", func(*gin.Context) {})
	g.GET("imports/:batch_id/errors/export", func(*gin.Context) {})
	g.GET("imports/:batch_id/donations", func(*gin.Context) {})
	g.GET("imports/:batch_id/rollback", func(*gin.Context) {})
	g.PUT(":id", func(*gin.Context) {})
	g.DELETE(":id", func(*gin.Context) {})
}

func TestImport_PersistsFailedRows(t *testing.T) {
	repo := baseRepo()
	svc := NewImportService(repo)
	xlsx := buildXLSX(t, [][]string{
		{"1", "D2053534", "Adrian", "Amal", "n", "20.000"}, // ok
		{"2", "D999999", "Ghost", "Amal", "x", "10.000"},   // donor missing -> failed row
		{"3", "D2053534", "Adrian", "ABC", "y", "5.000"},   // category missing -> failed row
	})
	_, err := svc.Import(context.Background(), bytes.NewReader(xlsx), ImportOptions{Filename: "d.xlsx"})
	if err != nil {
		t.Fatalf("Import error: %v", err)
	}
	if len(repo.persistedErrs) != 2 {
		t.Fatalf("expected 2 persisted error rows, got %d", len(repo.persistedErrs))
	}
	for _, e := range repo.persistedErrs {
		if e.BatchID == "" || e.RowNumber == 0 || e.ErrorMessage == "" {
			t.Errorf("error row not fully populated: %+v", e)
		}
	}
}

func TestHistory_ListEnrichesNames(t *testing.T) {
	uploader := 5
	repo := baseRepo()
	repo.userNames = map[int]string{5: "Budi Admin"}
	repo.importLogs = []ImportLog{
		{BatchID: "b1", Filename: "f.xlsx", UploadedBy: &uploader, Status: StatusCompleted, TotalRows: 10},
	}
	svc := NewImportHistoryService(repo)

	items, meta, err := svc.List(context.Background(), query.Filters{}, pagination.Query{Page: 1, Limit: 10}, ImportHistorySort{Column: "uploaded_at", Order: "desc"})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if meta.Total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 item, got %d (total %d)", len(items), meta.Total)
	}
	if items[0].UploadedByName != "Budi Admin" {
		t.Errorf("uploader name not enriched: %q", items[0].UploadedByName)
	}
}

func TestHistory_DetailNotFound(t *testing.T) {
	repo := baseRepo()
	repo.logByBatch = map[string]*ImportLog{} // unknown batch -> nil
	svc := NewImportHistoryService(repo)
	_, err := svc.Detail(context.Background(), "missing")
	if err != ErrBatchNotFound {
		t.Fatalf("got %v, want ErrBatchNotFound", err)
	}
}

func TestHistory_DetailSummaryAndRollback(t *testing.T) {
	uploader, rollbacker := 5, 9
	now := time.Now()
	repo := baseRepo()
	repo.userNames = map[int]string{5: "Uploader", 9: "Rollbacker"}
	repo.logByBatch = map[string]*ImportLog{
		"b1": {
			BatchID: "b1", Filename: "f.xlsx", UploadedBy: &uploader,
			Status: StatusRolledBack, TotalRows: 10, SuccessRows: 8, FailedRows: 2,
			RolledBackBy: &rollbacker, RolledBackAt: &now,
		},
	}
	svc := NewImportHistoryService(repo)

	detail, err := svc.Detail(context.Background(), "b1")
	if err != nil {
		t.Fatalf("Detail error: %v", err)
	}
	if detail.Summary.TotalRows != 10 || detail.Summary.SuccessRows != 8 || detail.Summary.FailedRows != 2 {
		t.Errorf("summary wrong: %+v", detail.Summary)
	}
	if detail.UploadedByName != "Uploader" {
		t.Errorf("uploader name = %q", detail.UploadedByName)
	}
	if !detail.Rollback.RolledBack || detail.Rollback.RolledBackByName != "Rollbacker" {
		t.Errorf("rollback info wrong: %+v", detail.Rollback)
	}
}

func TestHistory_ExportErrorsXLSX(t *testing.T) {
	repo := baseRepo()
	repo.logByBatch = map[string]*ImportLog{"b1": {BatchID: "b1", Filename: "f.xlsx", Status: StatusCompletedWithError}}
	repo.importErrors = map[string][]ImportError{
		"b1": {
			{BatchID: "b1", RowNumber: 3, DonaturID: "D999", CategoryName: "Amal", ErrorMessage: "Donatur D999 not found"},
		},
	}
	svc := NewImportHistoryService(repo)

	data, filename, err := svc.ExportErrors(context.Background(), "b1")
	if err != nil {
		t.Fatalf("ExportErrors error: %v", err)
	}
	if !strings.HasSuffix(filename, ".xlsx") {
		t.Errorf("filename = %q", filename)
	}
	// Re-open the produced file and verify the header + one row.
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("produced file not a valid xlsx: %v", err)
	}
	rows, _ := f.GetRows(f.GetSheetName(0))
	if len(rows) != 2 {
		t.Fatalf("expected header + 1 row, got %d rows", len(rows))
	}
	if rows[0][0] != "Row Number" || rows[0][3] != "Error" {
		t.Errorf("unexpected header: %v", rows[0])
	}
	if rows[1][1] != "D999" {
		t.Errorf("unexpected data row: %v", rows[1])
	}
}
