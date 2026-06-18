-- Attachment owner type: 0 = reporter, 1 = respondent.
ALTER TABLE speak_attachments
    ADD COLUMN IF NOT EXISTS type SMALLINT NOT NULL DEFAULT 0;
