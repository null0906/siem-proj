CREATE TABLE identity_users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_email      TEXT NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    mfa_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    account_status  TEXT NOT NULL DEFAULT 'active'
                    CHECK (account_status IN ('active', 'suspended', 'dormant')),
    last_login      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_privileged   BOOLEAN NOT NULL DEFAULT FALSE,
    groups          TEXT NOT NULL DEFAULT '',
    sso_apps_count  INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source_file_id  UUID REFERENCES source_files(id) ON DELETE SET NULL,
    UNIQUE (user_email, source_file_id)
);

CREATE INDEX idx_identity_users_status ON identity_users(account_status);
CREATE INDEX idx_identity_users_last_login ON identity_users(last_login DESC);
CREATE INDEX idx_identity_users_mfa ON identity_users(mfa_enabled);
CREATE INDEX idx_identity_users_privileged ON identity_users(is_privileged);
