-- Total activity duration (hours, 2dp) computed at checkout: checkout_at - checkin_at.
ALTER TABLE event_attendances
    ADD COLUMN IF NOT EXISTS total_hours NUMERIC(10, 2) NULL;

-- Backfill historical attendances that already have a valid check-in/check-out
-- pair. Rows with a NULL check-in, or a checkout earlier than the checkin
-- (bad data), are left NULL rather than storing a wrong/negative value.
UPDATE event_attendances
SET total_hours = ROUND(EXTRACT(EPOCH FROM (checkout_at - checkin_at)) / 3600.0, 2)
WHERE checkout_at IS NOT NULL
    AND checkin_at IS NOT NULL
    AND checkout_at >= checkin_at
    AND total_hours IS NULL;
