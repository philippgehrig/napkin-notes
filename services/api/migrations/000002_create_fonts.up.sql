CREATE TYPE font_status AS ENUM ('pending', 'processing', 'ready', 'failed');

CREATE TABLE fonts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    file_path VARCHAR(512) NOT NULL DEFAULT '',
    status font_status NOT NULL DEFAULT 'pending',
    template_scan_path VARCHAR(512) NOT NULL DEFAULT '',
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
