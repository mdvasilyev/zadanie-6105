DROP TABLE IF EXISTS bid_diff;

DROP TRIGGER IF EXISTS author_validation ON bid;

DROP FUNCTION IF EXISTS validate_author();

DROP TABLE IF EXISTS bid;

DROP TYPE IF EXISTS author_type;

DROP TYPE IF EXISTS bid_status;
