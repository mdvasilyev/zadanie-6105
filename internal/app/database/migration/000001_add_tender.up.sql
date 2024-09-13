CREATE TYPE tender_status AS ENUM (
    'Created',
    'Published',
    'Closed'
);

CREATE TYPE service_type AS ENUM (
    'Construction',
    'Delivery',
    'Manufacture'
);

CREATE TABLE tender (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    status tender_status NOT NULL,
    service_type service_type NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    creator_username VARCHAR(100) NOT NULL REFERENCES employee(username) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tender_diff (
    id UUID NOT NULL REFERENCES tender(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL,
    status tender_status NOT NULL,
    service_type service_type NOT NULL,
    version INTEGER NOT NULL,
    organization_id UUID NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
    creator_username VARCHAR(100) NOT NULL REFERENCES employee(username) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL
);
