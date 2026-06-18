package donation

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"social-be/internal/pkg/apperror"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"social-be/internal/pkg/response"
	"social-be/internal/pkg/validation"

	"github.com/gin-gonic/gin"
)

// maxImportFileSize caps the uploaded spreadsheet at 10MB.
const maxImportFileSize = 10 << 20

// allowedImportExt is the set of accepted spreadsheet extensions. Note: legacy
// binary .xls is not parseable by excelize and will be rejected at parse time
// with a descriptive error; users should re-save as .xlsx.
var allowedImportExt = map[string]bool{
	".xlsx": true,
	".xls":  true,
}

// xlsxMagic is the ZIP local-file-header signature ("PK\x03\x04") that every
// .xlsx (OOXML/ZIP) file begins with. Used as a content sniff so a renamed
// non-spreadsheet file is rejected before parsing.
var xlsxMagic = []byte{0x50, 0x4B, 0x03, 0x04}

type Handler struct {
	Service        *Service
	ImportService  *ImportService
	HistoryService *ImportHistoryService
}

func NewHandler(service *Service, importService *ImportService, historyService *ImportHistoryService) *Handler {
	return &Handler{Service: service, ImportService: importService, HistoryService: historyService}
}

// importHistorySortable whitelists the columns the history list may be sorted
// by, mapping the public name to its DB column.
var importHistorySortable = map[string]string{
	"uploaded_at":  "uploaded_at",
	"filename":     "filename",
	"status":       "status",
	"total_rows":   "total_rows",
	"success_rows": "success_rows",
	"failed_rows":  "failed_rows",
	"skipped_rows": "skipped_rows",
	"period":       "period",
	"batch_id":     "batch_id",
	"id":           "id",
}

// GetAll godoc
// @Summary Get all donations
// @Description Get list of donations
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /donations [get]
func (h *Handler) GetAll(c *gin.Context) {
	items, err := h.Service.GetAll(c.Request.Context())
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_900", "failed to fetch donations"))
		return
	}
	response.Success(c, items)
}

// Create godoc
// @Summary Create donation
// @Description Create new donation
// @Tags donation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRequest true "Create donation"
// @Success 200 {object} map[string]interface{}
// @Router /donations [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	donation, err := h.Service.Create(c.Request.Context(), req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_901", "failed to create donation"))
		return
	}

	response.Success(c, donation)
}

// GetByID godoc
// @Summary Get donation by ID
// @Description Get detail donation
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation ID"
// @Success 200 {object} map[string]interface{}
// @Router /donations/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	donation, err := h.Service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusNotFound, "DB_902", "donation not found"))
		return
	}

	response.Success(c, donation)
}

// Update godoc
// @Summary Update donation
// @Description Update existing donation
// @Tags donation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation ID"
// @Param request body UpdateRequest true "Update donation"
// @Success 200 {object} map[string]interface{}
// @Router /donations/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	var req UpdateRequest
	if err := validation.BindJSON(c, &req); err != nil {
		response.AbortError(c, err)
		return
	}

	donation, err := h.Service.Update(c.Request.Context(), id, req)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_903", "failed to update donation"))
		return
	}

	response.Success(c, donation)
}

// Delete godoc
// @Summary Delete donation
// @Description Delete donation
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param id path int true "Donation ID"
// @Success 200 {object} map[string]interface{}
// @Router /donations/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid id"))
		return
	}

	if err := h.Service.Delete(c.Request.Context(), id); err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_904", "failed to delete donation"))
		return
	}

	response.Success(c, gin.H{"message": "donation deleted"})
}

// Import godoc
// @Summary Import donations from Excel
// @Description Upload an Excel (.xlsx) file of donation records. Each row is
// @Description validated and created independently; one bad row does not stop the rest.
// @Description Use dry_run=true to validate without writing any data.
// @Tags donation
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Excel file (.xlsx)"
// @Param dry_run query bool false "Validate only, do not insert"
// @Success 200 {object} map[string]interface{}
// @Router /donations/import [post]
func (h *Handler) Import(c *gin.Context) {
	actor, ok := actorID(c)
	if !ok {
		return
	}

	// Cap the request body before any buffering to defend against oversized
	// uploads (the multipart reader will error past the limit).
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImportFileSize)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, "file is required (max 10MB)"))
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedImportExt[ext] {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, "unsupported file type: only .xlsx and .xls are allowed"))
		return
	}
	if fileHeader.Size > maxImportFileSize {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, "file size exceeds 10MB limit"))
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, apperror.CodeInternal, "failed to open uploaded file"))
		return
	}
	defer src.Close()

	// Content sniff: confirm the bytes actually look like a ZIP/OOXML file
	// before handing them to the parser (defends against renamed files).
	header := make([]byte, len(xlsxMagic))
	n, _ := io.ReadFull(src, header)
	if n < len(xlsxMagic) || !bytes.Equal(header, xlsxMagic) {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, "file is not a valid .xlsx workbook"))
		return
	}
	reader := io.MultiReader(bytes.NewReader(header), src)

	opts := ImportOptions{
		Filename:   filepath.Base(fileHeader.Filename),
		UploadedBy: &actor,
		DryRun:     c.Query("dry_run") == "true",
	}

	result, err := h.ImportService.Import(c.Request.Context(), reader, opts)
	if err != nil {
		// Template/parse errors carry an uploader-friendly message.
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidBody, err.Error()))
		return
	}

	response.Success(c, result)
}

// Rollback godoc
// @Summary Rollback an import batch
// @Description Delete every donation created by a given import batch.
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} map[string]interface{}
// @Router /donations/import/{batch_id} [delete]
func (h *Handler) Rollback(c *gin.Context) {
	actor, ok := actorID(c)
	if !ok {
		return
	}

	batchID := strings.TrimSpace(c.Param("batch_id"))
	if batchID == "" {
		response.AbortError(c, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "batch_id is required"))
		return
	}

	deleted, err := h.ImportService.RollbackBatch(c.Request.Context(), batchID, &actor)
	if err != nil {
		switch {
		case errors.Is(err, ErrBatchNotFound):
			response.AbortError(c, apperror.New(http.StatusNotFound, "DB_902", "import batch not found"))
		case errors.Is(err, ErrBatchAlreadyRolledBk):
			response.AbortError(c, apperror.New(http.StatusConflict, "DB_907", "import batch already rolled back"))
		default:
			response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_906", "failed to rollback import batch"))
		}
		return
	}

	response.Success(c, gin.H{"batch_id": batchID, "deleted_rows": deleted})
}

// actorID extracts the authenticated user id set by the auth middleware.
func actorID(c *gin.Context) (int, bool) {
	value, ok := c.Get("user_id")
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "user not found"))
		return 0, false
	}
	id, ok := value.(int)
	if !ok {
		response.AbortError(c, apperror.New(http.StatusUnauthorized, apperror.CodeActorNotFound, "invalid user"))
		return 0, false
	}
	return id, true
}

// ListImports godoc
// @Summary List import history
// @Description Filtered, sorted, paginated import history (SuperAdmin/Admin).
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param sort query string false "Sort column"
// @Param order query string false "asc or desc"
// @Param batch_id query string false "Filter by batch id"
// @Param filename query string false "Filter by filename (partial)"
// @Param status query string false "Filter by status"
// @Param uploaded_by query int false "Filter by uploader user id"
// @Param period query string false "Filter by period"
// @Param uploaded_at_from query string false "Uploaded from (date)"
// @Param uploaded_at_to query string false "Uploaded to (date)"
// @Success 200 {object} map[string]interface{}
// @Router /donations/imports [get]
func (h *Handler) ListImports(c *gin.Context) {
	page, appErr := pagination.FromGin(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	filters, appErr := query.ParseFilters(c.Request.URL.Query(), ImportLog{})
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}
	// These should match exactly, not partially.
	exactFilter(&filters, "status")
	exactFilter(&filters, "batch_id")
	exactFilter(&filters, "period")

	sort, appErr := parseImportSort(c)
	if appErr != nil {
		response.AbortError(c, appErr)
		return
	}

	items, meta, err := h.HistoryService.List(c.Request.Context(), filters, page, sort)
	if err != nil {
		response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_910", "failed to fetch import history"))
		return
	}

	response.SuccessWithPagination(c, items, meta)
}

// GetImportDetail godoc
// @Summary Get import detail
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} map[string]interface{}
// @Router /donations/imports/{batch_id} [get]
func (h *Handler) GetImportDetail(c *gin.Context) {
	batchID := strings.TrimSpace(c.Param("batch_id"))
	detail, err := h.HistoryService.Detail(c.Request.Context(), batchID)
	if err != nil {
		h.abortHistoryErr(c, err, "failed to fetch import detail")
		return
	}
	response.Success(c, detail)
}

// GetImportErrors godoc
// @Summary Get failed rows for an import batch
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} map[string]interface{}
// @Router /donations/imports/{batch_id}/errors [get]
func (h *Handler) GetImportErrors(c *gin.Context) {
	batchID := strings.TrimSpace(c.Param("batch_id"))
	rows, err := h.HistoryService.Errors(c.Request.Context(), batchID)
	if err != nil {
		h.abortHistoryErr(c, err, "failed to fetch import errors")
		return
	}
	response.Success(c, rows)
}

// GetImportDonations godoc
// @Summary Get donations created by an import batch
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} map[string]interface{}
// @Router /donations/imports/{batch_id}/donations [get]
func (h *Handler) GetImportDonations(c *gin.Context) {
	batchID := strings.TrimSpace(c.Param("batch_id"))
	donations, err := h.HistoryService.Donations(c.Request.Context(), batchID)
	if err != nil {
		h.abortHistoryErr(c, err, "failed to fetch batch donations")
		return
	}
	response.Success(c, donations)
}

// GetImportRollback godoc
// @Summary Get rollback info for an import batch
// @Tags donation
// @Produce json
// @Security BearerAuth
// @Param batch_id path string true "Import batch id"
// @Success 200 {object} map[string]interface{}
// @Router /donations/imports/{batch_id}/rollback [get]
func (h *Handler) GetImportRollback(c *gin.Context) {
	batchID := strings.TrimSpace(c.Param("batch_id"))
	info, err := h.HistoryService.Rollback(c.Request.Context(), batchID)
	if err != nil {
		h.abortHistoryErr(c, err, "failed to fetch rollback info")
		return
	}
	response.Success(c, info)
}

// ExportImportErrors godoc
// @Summary Download the error report (.xlsx) for an import batch
// @Tags donation
// @Produce application/octet-stream
// @Security BearerAuth
// @Param batch_id path string true "Import batch id"
// @Success 200 {file} binary
// @Router /donations/imports/{batch_id}/errors/export [get]
func (h *Handler) ExportImportErrors(c *gin.Context) {
	batchID := strings.TrimSpace(c.Param("batch_id"))
	data, filename, err := h.HistoryService.ExportErrors(c.Request.Context(), batchID)
	if err != nil {
		h.abortHistoryErr(c, err, "failed to export import errors")
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// abortHistoryErr maps history-service errors to HTTP responses.
func (h *Handler) abortHistoryErr(c *gin.Context, err error, msg string) {
	if errors.Is(err, ErrBatchNotFound) {
		response.AbortError(c, apperror.New(http.StatusNotFound, "DB_902", "import batch not found"))
		return
	}
	response.AbortError(c, apperror.Wrap(err, http.StatusInternalServerError, "DB_911", msg))
}

// exactFilter promotes a string filter from partial (ILIKE) to exact match.
func exactFilter(filters *query.Filters, field string) {
	if v, ok := filters.Like[field]; ok {
		delete(filters.Like, field)
		filters.Equals[field] = v
	}
}

// parseImportSort validates the sort/order query params against the whitelist.
func parseImportSort(c *gin.Context) (ImportHistorySort, *apperror.AppError) {
	sortParam := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort", "uploaded_at")))
	column, ok := importHistorySortable[sortParam]
	if !ok {
		return ImportHistorySort{}, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "invalid sort column")
	}

	order := strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc")))
	if order != "asc" && order != "desc" {
		return ImportHistorySort{}, apperror.New(http.StatusBadRequest, apperror.CodeInvalidParam, "order must be asc or desc")
	}

	return ImportHistorySort{Column: column, Order: order}, nil
}
