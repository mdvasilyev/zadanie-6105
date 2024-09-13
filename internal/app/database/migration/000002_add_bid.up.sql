CREATE TYPE bid_status AS ENUM (
    'Created',
    'Published',
    'Cancelled'
);

CREATE TYPE author_type AS ENUM (
    'Organization',
    'User'
);

CREATE TABLE bid (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    status bid_status NOT NULL,
    tender_id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
    author_type author_type NOT NULL,
    author_id UUID NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE FUNCTION validate_author() RETURNS TRIGGER AS $$
BEGIN
    IF (NEW.author_type = 'User') THEN
        IF NOT EXISTS (SELECT 1 FROM employee WHERE id = NEW.author_id) THEN
            RAISE EXCEPTION 'Invalid author_id for User';
        END IF;
    ELSIF (NEW.author_type = 'Organization') THEN
        IF NOT EXISTS (SELECT 1 FROM organization WHERE id = NEW.author_id) THEN
            RAISE EXCEPTION 'Invalid author_id for Organization';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER author_validation BEFORE INSERT OR UPDATE ON bid
FOR EACH ROW EXECUTE FUNCTION validate_author();

CREATE TABLE bid_diff (
    id UUID NOT NULL REFERENCES bid(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    status bid_status NOT NULL,
    tender_id UUID REFERENCES tender(id) ON DELETE CASCADE,
    author_type author_type NOT NULL,
    author_id UUID NOT NULL,
    version INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL
);
