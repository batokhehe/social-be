-- Indexes supporting GET /dashboard/volunteer. All are partial on
-- deleted_at IS NULL to match the query predicates and keep them small.
-- NOTE: on very large existing tables, create these CONCURRENTLY by hand to
-- avoid write locks; golang-migrate wraps each file in a transaction so
-- CONCURRENTLY cannot be used here.

-- Activity hours per volunteer over a time window (also helps /dashboard/home
-- top-volunteers and the applied/completed event lists). The most selective
-- filter (volunteer_id) currently has no index at all.
CREATE INDEX IF NOT EXISTS idx_event_attendances_volunteer_checkout
    ON event_attendances (volunteer_id, checkout_at)
    WHERE deleted_at IS NULL;

-- Resolve the donatur groups a volunteer owns. volunteer_id holds volunteers.id
-- as text and is matched via btrim(), so an expression index is required.
CREATE INDEX IF NOT EXISTS idx_master_donatur_groups_volunteer
    ON master_donatur_groups (volunteer_id)
    WHERE deleted_at IS NULL;

-- Group -> donors traversal (My Donors, Collected, monthly donor_count).
CREATE INDEX IF NOT EXISTS idx_master_donaturs_group
    ON master_donaturs (id_group_donatur)
    WHERE deleted_at IS NULL;

-- Donors that are themselves a volunteer (My Donations). id_vis_volunteer holds
-- volunteers.id as text and is matched via btrim(), so an expression index.
CREATE INDEX IF NOT EXISTS idx_master_donaturs_vis_volunteer
    ON master_donaturs (id_vis_volunteer)
    WHERE deleted_at IS NULL;
