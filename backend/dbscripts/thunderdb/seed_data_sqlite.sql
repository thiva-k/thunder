-- SQLite data seeding script for Thunder database
-- This script contains initial data to populate the database tables

-- Insert sample data into the tables.
INSERT OR IGNORE INTO SP_APP (APP_NAME, APP_ID, DESCRIPTION, AUTH_FLOW_GRAPH_ID, REGISTRATION_FLOW_GRAPH_ID) 
VALUES ('Test SPA', '550e8400-e29b-41d4-a716-446655440000', 'Initial testing App', 'auth_flow_config_basic', 'registration_flow_config_basic');

INSERT OR IGNORE INTO IDN_OAUTH_CONSUMER_APPS (CONSUMER_KEY, CONSUMER_SECRET, APP_ID, CALLBACK_URIS, GRANT_TYPES, RESPONSE_TYPES, TOKEN_ENDPOINT_AUTH_METHODS)
VALUES ('client123', 'fcf730b6d95236ecd3c9fc2d92d7b6b2bb061514961aec041d6c7a7192f592e4', '550e8400-e29b-41d4-a716-446655440000', 'https://localhost:3000', 'client_credentials,authorization_code,refresh_token', 'code', 'client_secret_basic,client_secret_post');

INSERT OR IGNORE INTO SP_INBOUND_AUTH (INBOUND_AUTH_KEY, INBOUND_AUTH_TYPE, APP_ID)
VALUES ('client123', 'oauth2', '550e8400-e29b-41d4-a716-446655440000');

INSERT OR IGNORE INTO IDN_OAUTH_ALLOWED_ORIGINS (ALLOWED_ORIGINS) 
VALUES ('https://localhost:3000,https://localhost:9001,https://localhost:9090');

-- Insert sample organization units (must be inserted in order due to foreign key constraints)
INSERT OR IGNORE INTO ORGANIZATION_UNIT (OU_ID, PARENT_ID, HANDLE, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT)
VALUES
('456e8400-e29b-41d4-a716-446655440001', NULL, 'root', 'Root Organization', 'Root organization unit', datetime('now'), datetime('now')),
('456e8400-e29b-41d4-a716-446655440002', '456e8400-e29b-41d4-a716-446655440001', 'engineering', 'Engineering', 'Engineering department', datetime('now'), datetime('now')),
('456e8400-e29b-41d4-a716-446655440003', '456e8400-e29b-41d4-a716-446655440001', 'sales', 'Sales', 'Sales department', datetime('now'), datetime('now')),
('456e8400-e29b-41d4-a716-446655440004', '456e8400-e29b-41d4-a716-446655440002', 'frontend', 'Frontend Team', 'Frontend development team', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO USER (USER_ID, OU_ID, TYPE, ATTRIBUTES, CREATED_AT, UPDATED_AT)
VALUES (
'550e8400-e29b-41d4-a716-446655440000', '456e8400-e29b-41d4-a716-446655440001', 'person',
'{"age": 30, "roles": ["admin", "user"], "address": {"city": "Colombo", "zip": "00100"}}',
datetime('now'), datetime('now')
);

INSERT OR IGNORE INTO IDP (IDP_ID, NAME, DESCRIPTION, CREATED_AT, UPDATED_AT)
VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Local', 'Local Identity Provider', datetime('now'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440001', 'Github', 'Login with Github', datetime('now'), datetime('now')),
('550e8400-e29b-41d4-a716-446655440002', 'Google', 'Login with Google', datetime('now'), datetime('now'));

INSERT OR IGNORE INTO IDP_PROPERTY (IDP_ID, PROPERTY_NAME, PROPERTY_VALUE, IS_SECRET)
VALUES
('550e8400-e29b-41d4-a716-446655440001', 'client_id', 'client1', '0'),
('550e8400-e29b-41d4-a716-446655440001', 'client_secret', 'secret1', '1'),
('550e8400-e29b-41d4-a716-446655440001', 'redirect_uri', 'https://localhost:3000', '0'),
('550e8400-e29b-41d4-a716-446655440001', 'scopes', 'user:email,read:user', '0'),
('550e8400-e29b-41d4-a716-446655440002', 'client_id', 'client2', '0'),
('550e8400-e29b-41d4-a716-446655440002', 'client_secret', 'secret2', '1'),
('550e8400-e29b-41d4-a716-446655440002', 'redirect_uri', 'https://localhost:3000', '0'),
('550e8400-e29b-41d4-a716-446655440002', 'scopes', 'openid,email,profile', '0');