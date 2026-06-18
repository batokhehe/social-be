-- Make donation_import_logs an immutable audit trail: the identifying/audit
-- columns can never change after insert, and rows can never be deleted. Only
-- the mutable lifecycle columns (status, and the row counters written once at
-- creation) may be updated -- in practice only `status` (e.g. -> rolled_back).
CREATE OR REPLACE FUNCTION donation_import_logs_guard()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'donation_import_logs are immutable and cannot be deleted';
    END IF;

    IF NEW.batch_id        IS DISTINCT FROM OLD.batch_id
        OR NEW.filename        IS DISTINCT FROM OLD.filename
        OR NEW.uploaded_by     IS DISTINCT FROM OLD.uploaded_by
        OR NEW.uploaded_at     IS DISTINCT FROM OLD.uploaded_at
        OR NEW.komisariat_id   IS DISTINCT FROM OLD.komisariat_id
        OR NEW.komisariat_name IS DISTINCT FROM OLD.komisariat_name
        OR NEW.period          IS DISTINCT FROM OLD.period
        OR NEW.total_rows      IS DISTINCT FROM OLD.total_rows
        OR NEW.success_rows    IS DISTINCT FROM OLD.success_rows
        OR NEW.failed_rows     IS DISTINCT FROM OLD.failed_rows
        OR NEW.skipped_rows    IS DISTINCT FROM OLD.skipped_rows THEN
        RAISE EXCEPTION 'donation_import_logs audit columns are immutable; only status may change';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_donation_import_logs_guard
    BEFORE UPDATE OR DELETE ON donation_import_logs
    FOR EACH ROW EXECUTE FUNCTION donation_import_logs_guard();
