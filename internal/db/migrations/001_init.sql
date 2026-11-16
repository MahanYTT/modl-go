CREATE TABLE IF NOT EXISTS servers (
    id TEXT PRIMARY KEY,
    server_name TEXT NOT NULL,
    custom_domain TEXT NOT NULL UNIQUE,
    custom_domain_override TEXT NOT NULL DEFAULT '',
    admin_email TEXT NOT NULL,
    email_verified INTEGER NOT NULL DEFAULT 1,
    email_verification_token TEXT NOT NULL DEFAULT '',
    provisioning_status TEXT NOT NULL DEFAULT 'COMPLETED',
    provisioning_notes TEXT NOT NULL DEFAULT '',
    provisioning_sign_in_token TEXT NOT NULL DEFAULT '',
    provisioning_sign_in_token_expires_at INTEGER,
    cli_setup_token TEXT NOT NULL DEFAULT '',
    api_key TEXT NOT NULL DEFAULT '',
    online_player_count INTEGER NOT NULL DEFAULT 0,
    user_count INTEGER NOT NULL DEFAULT 0,
    ticket_count INTEGER NOT NULL DEFAULT 0,
    last_stats_updated_at INTEGER,
    last_activity_at INTEGER,
    staff_permissions_updated_at INTEGER,
    punishment_types_updated_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_servers_api_key ON servers(api_key);
CREATE INDEX IF NOT EXISTS idx_servers_cli_setup_token ON servers(cli_setup_token);

CREATE TABLE IF NOT EXISTS auth_codes (
    server_id TEXT NOT NULL,
    email TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    expires_at INTEGER NOT NULL,
    PRIMARY KEY (server_id, email)
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    ip_address TEXT NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_sessions_server_email ON sessions(server_id, email);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS webauthn_credentials (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    email TEXT NOT NULL,
    credential_id TEXT NOT NULL,
    public_key_cose BLOB NOT NULL,
    signature_count INTEGER NOT NULL DEFAULT 0,
    user_handle TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    raw_json TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    last_used_at INTEGER
);
CREATE INDEX IF NOT EXISTS idx_webauthn_creds_server_email ON webauthn_credentials(server_id, email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_webauthn_creds_cred_id ON webauthn_credentials(server_id, credential_id);

CREATE TABLE IF NOT EXISTS webauthn_challenges (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    challenge_json TEXT NOT NULL,
    email TEXT NOT NULL DEFAULT '',
    expires_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS staff_roles (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    permissions_json TEXT NOT NULL DEFAULT '[]',
    is_default INTEGER NOT NULL DEFAULT 0,
    ord INTEGER NOT NULL DEFAULT 999,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_staff_roles_server ON staff_roles(server_id, ord);

CREATE TABLE IF NOT EXISTS staff (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    email TEXT NOT NULL,
    username TEXT NOT NULL DEFAULT '',
    role_id TEXT NOT NULL DEFAULT '',
    assigned_minecraft_uuid TEXT NOT NULL DEFAULT '',
    assigned_minecraft_username TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT 'en',
    date_format TEXT NOT NULL DEFAULT 'MM/dd/yyyy',
    subscribed_tickets_json TEXT NOT NULL DEFAULT '[]',
    ticket_subscription_settings_json TEXT NOT NULL DEFAULT '{}',
    two_factor_token TEXT NOT NULL DEFAULT '',
    two_factor_token_ip TEXT NOT NULL DEFAULT '',
    two_factor_token_created_at INTEGER,
    two_factor_pending_delivery INTEGER NOT NULL DEFAULT 0,
    two_factor_session_ip TEXT NOT NULL DEFAULT '',
    two_factor_session_expires_at INTEGER,
    last_seen INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE (server_id, email)
);
CREATE INDEX IF NOT EXISTS idx_staff_mc_uuid ON staff(server_id, assigned_minecraft_uuid);

CREATE TABLE IF NOT EXISTS invitations (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    email TEXT NOT NULL,
    role_id TEXT NOT NULL DEFAULT '',
    token TEXT NOT NULL,
    expires_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(server_id, token);

CREATE TABLE IF NOT EXISTS settings (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    type TEXT NOT NULL,
    data_json TEXT NOT NULL DEFAULT '{}',
    version INTEGER NOT NULL DEFAULT 1,
    updated_at INTEGER NOT NULL,
    UNIQUE (server_id, type)
);

CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    minecraft_uuid TEXT NOT NULL,
    current_username TEXT NOT NULL DEFAULT '',
    usernames_json TEXT NOT NULL DEFAULT '[]',
    notes_json TEXT NOT NULL DEFAULT '[]',
    ip_addresses_json TEXT NOT NULL DEFAULT '[]',
    data_json TEXT NOT NULL DEFAULT '{}',
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL,
    UNIQUE (server_id, minecraft_uuid)
);
CREATE INDEX IF NOT EXISTS idx_players_username ON players(server_id, current_username);
CREATE INDEX IF NOT EXISTS idx_players_last_seen ON players(server_id, last_seen DESC);

CREATE TABLE IF NOT EXISTS player_ips (
    server_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    minecraft_uuid TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL,
    PRIMARY KEY (server_id, player_id, ip_address)
);
CREATE INDEX IF NOT EXISTS idx_player_ips_lookup ON player_ips(server_id, ip_address);

CREATE TABLE IF NOT EXISTS punishments (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    player_uuid TEXT NOT NULL,
    player_username TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL,
    punishment_type_id TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    issuer_uuid TEXT NOT NULL DEFAULT '',
    issuer_name TEXT NOT NULL DEFAULT '',
    issued_at INTEGER NOT NULL,
    expires_at INTEGER,
    duration_ms INTEGER,
    pardoned INTEGER NOT NULL DEFAULT 0,
    pardoned_by TEXT NOT NULL DEFAULT '',
    pardoned_at INTEGER,
    pardon_reason TEXT NOT NULL DEFAULT '',
    points INTEGER NOT NULL DEFAULT 0,
    ip_ban INTEGER NOT NULL DEFAULT 0,
    ip_address TEXT NOT NULL DEFAULT '',
    silent INTEGER NOT NULL DEFAULT 0,
    wipe_stats INTEGER NOT NULL DEFAULT 0,
    block_alts INTEGER NOT NULL DEFAULT 0,
    notes_json TEXT NOT NULL DEFAULT '[]',
    evidence_json TEXT NOT NULL DEFAULT '[]',
    data_json TEXT NOT NULL DEFAULT '{}',
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_punishments_player ON punishments(server_id, player_uuid, issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_punishments_active ON punishments(server_id, pardoned, expires_at);
CREATE INDEX IF NOT EXISTS idx_punishments_issuer ON punishments(server_id, issuer_name, issued_at DESC);
CREATE INDEX IF NOT EXISTS idx_punishments_issued ON punishments(server_id, issued_at DESC);

CREATE TABLE IF NOT EXISTS chat_logs (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    uuid TEXT NOT NULL,
    username TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    server_name TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_chat_logs_player ON chat_logs(server_id, uuid, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_chat_logs_ts ON chat_logs(server_id, timestamp DESC);

CREATE TABLE IF NOT EXISTS command_logs (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    uuid TEXT NOT NULL,
    username TEXT NOT NULL,
    command TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    server_name TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_command_logs_player ON command_logs(server_id, uuid, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_command_logs_ts ON command_logs(server_id, timestamp DESC);

CREATE TABLE IF NOT EXISTS player_notifications (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    player_uuid TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    data_json TEXT NOT NULL DEFAULT '{}',
    delivered INTEGER NOT NULL DEFAULT 0,
    read_at INTEGER,
    created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_player_notif_pending ON player_notifications(server_id, player_uuid, delivered);

CREATE TABLE IF NOT EXISTS tickets (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    type TEXT NOT NULL,
    subject TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'OPEN',
    creator_uuid TEXT NOT NULL DEFAULT '',
    creator_name TEXT NOT NULL DEFAULT '',
    creator_avatar TEXT NOT NULL DEFAULT '',
    reported_player TEXT NOT NULL DEFAULT '',
    reported_player_uuid TEXT NOT NULL DEFAULT '',
    punishment_id TEXT NOT NULL DEFAULT '',
    tags_json TEXT NOT NULL DEFAULT '[]',
    replies_json TEXT NOT NULL DEFAULT '[]',
    notes_json TEXT NOT NULL DEFAULT '[]',
    chat_messages_json TEXT NOT NULL DEFAULT '[]',
    form_data_json TEXT NOT NULL DEFAULT '{}',
    data_json TEXT NOT NULL DEFAULT '{}',
    locked INTEGER NOT NULL DEFAULT 0,
    priority TEXT NOT NULL DEFAULT 'NORMAL',
    assigned_to_json TEXT NOT NULL DEFAULT '[]',
    appeal_workflow_status TEXT NOT NULL DEFAULT '',
    ai_analysis_json TEXT,
    email_auth_enabled INTEGER NOT NULL DEFAULT 0,
    hidden INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tickets_server_status ON tickets(server_id, status, type, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_creator ON tickets(server_id, creator_uuid, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tickets_punishment ON tickets(server_id, punishment_id);

CREATE TABLE IF NOT EXISTS ticket_verifications (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    ticket_id TEXT NOT NULL,
    token TEXT NOT NULL DEFAULT '',
    code_hash TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    expires_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ticket_verif_ticket ON ticket_verifications(server_id, ticket_id);

CREATE TABLE IF NOT EXISTS ticket_subscription_updates (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    staff_email TEXT NOT NULL,
    ticket_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    actor TEXT NOT NULL DEFAULT '',
    is_read INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sub_updates_staff ON ticket_subscription_updates(server_id, staff_email, is_read, created_at DESC);

CREATE TABLE IF NOT EXISTS homepage_cards (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    icon_color TEXT NOT NULL DEFAULT '',
    action_type TEXT NOT NULL DEFAULT '',
    action_url TEXT NOT NULL DEFAULT '',
    action_button_text TEXT NOT NULL DEFAULT '',
    category_id TEXT NOT NULL DEFAULT '',
    background_color TEXT NOT NULL DEFAULT '',
    ordinal INTEGER NOT NULL DEFAULT 0,
    is_enabled INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_homepage_cards_server ON homepage_cards(server_id, ordinal);

CREATE TABLE IF NOT EXISTS knowledgebase_categories (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    ordinal INTEGER NOT NULL DEFAULT 0,
    is_visible INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE (server_id, slug)
);

CREATE TABLE IF NOT EXISTS knowledgebase_articles (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    category_id TEXT NOT NULL DEFAULT '',
    ordinal INTEGER NOT NULL DEFAULT 0,
    is_visible INTEGER NOT NULL DEFAULT 1,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    UNIQUE (server_id, slug)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    level TEXT NOT NULL DEFAULT 'info',
    source TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    metadata_json TEXT NOT NULL DEFAULT '{}'
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_server ON audit_logs(server_id, created_at DESC);

CREATE TABLE IF NOT EXISTS server_logs (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    description TEXT NOT NULL,
    level TEXT NOT NULL DEFAULT 'info',
    source TEXT NOT NULL DEFAULT 'system',
    created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_server_logs_server ON server_logs(server_id, created_at DESC);

CREATE TABLE IF NOT EXISTS system_alerts (
    id TEXT PRIMARY KEY,
    message TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'INFO',
    audience TEXT NOT NULL DEFAULT 'ALL',
    expires_at INTEGER,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    created_by TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS metric_snapshots (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    date_ms INTEGER NOT NULL,
    active_servers INTEGER NOT NULL DEFAULT 0,
    online_players INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS server_instance_snapshots (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    date_ms INTEGER NOT NULL,
    servers_json TEXT NOT NULL DEFAULT '[]',
    created_at INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS migrations (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    progress_json TEXT NOT NULL DEFAULT '{}',
    started_at INTEGER NOT NULL,
    completed_at INTEGER,
    error TEXT NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    records_processed INTEGER NOT NULL DEFAULT 0,
    records_skipped INTEGER NOT NULL DEFAULT 0,
    total_records INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_migrations_server ON migrations(server_id, started_at DESC);

CREATE TABLE IF NOT EXISTS storage_files (
    id TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    key TEXT NOT NULL UNIQUE,
    file_name TEXT NOT NULL,
    size INTEGER NOT NULL DEFAULT 0,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    category TEXT NOT NULL DEFAULT 'media',
    created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_storage_files_server ON storage_files(server_id, created_at DESC);

CREATE TABLE IF NOT EXISTS evidence_upload_tokens (
    token_hash TEXT PRIMARY KEY,
    server_id TEXT NOT NULL,
    punishment_id TEXT NOT NULL,
    player_uuid TEXT NOT NULL,
    issuer_name TEXT NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL
);
