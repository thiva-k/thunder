-- Table to store OAuth2 authorization codes.
CREATE TABLE IDN_OAUTH2_AUTHZ_CODE (
    ID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    CODE_ID VARCHAR(36) UNIQUE NOT NULL,
    AUTHORIZATION_CODE VARCHAR(500) NOT NULL,
    CONSUMER_KEY VARCHAR(255) NOT NULL,
    CALLBACK_URL VARCHAR(500),
    AUTHZ_USER VARCHAR(255) NOT NULL,
    TIME_CREATED TIMESTAMP NOT NULL,
    EXPIRY_TIME TIMESTAMP NOT NULL,
    STATE VARCHAR(50) NOT NULL
);

-- Table to store scopes associated with OAuth2 authorization codes.
CREATE TABLE IDN_OAUTH2_AUTHZ_CODE_SCOPE (
    ID INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    CODE_ID VARCHAR(36) NOT NULL REFERENCES IDN_OAUTH2_AUTHZ_CODE(CODE_ID) ON DELETE CASCADE,
    SCOPE VARCHAR(255) NOT NULL,
    UNIQUE (CODE_ID, SCOPE)
);

-- Table to store flow context for each flow instance
CREATE TABLE flow_context (
    flow_id VARCHAR PRIMARY KEY,
    flow_type VARCHAR NOT NULL,
    app_id VARCHAR NOT NULL,
    current_node_id VARCHAR NOT NULL,
    current_action_id VARCHAR,
    graph_id VARCHAR NOT NULL,
    is_authenticated BOOLEAN,
    authenticated_user_id VARCHAR,
    user_input_data JSONB,
    runtime_data JSONB,
    authenticated_user_attributes JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table to store the current node response for each flow
CREATE TABLE flow_current_node_response (
    flow_id VARCHAR PRIMARY KEY REFERENCES flow_context(flow_id) ON DELETE CASCADE,
    status VARCHAR,
    type VARCHAR,
    failure_reason VARCHAR,
    redirect_url VARCHAR,
    next_node_id VARCHAR,
    assertion VARCHAR,
    authenticated_user_id VARCHAR,
    is_authenticated BOOLEAN,
    required_data JSONB,
    additional_data JSONB,
    actions JSONB,
    runtime_data JSONB,
    authenticated_user_attributes JSONB
);

-- Table to store each flow graph
CREATE TABLE graphs (
    graph_id VARCHAR PRIMARY KEY,
    type VARCHAR NOT NULL,
    nodes JSONB,
    edges JSONB,
    start_node_id VARCHAR
);

-- Table to store each node in a graph
CREATE TABLE nodes (
    node_id VARCHAR PRIMARY KEY,
    graph_id VARCHAR NOT NULL REFERENCES graphs(graph_id) ON DELETE CASCADE,
    type VARCHAR NOT NULL,
    is_start_node BOOLEAN,
    is_final_node BOOLEAN,
    next_node_list JSONB,
    previous_node_list JSONB,
    input_data JSONB,
    executor_config JSONB
);
