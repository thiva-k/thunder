-- Table to store OAuth2 authorization codes.
CREATE TABLE IDN_OAUTH2_AUTHZ_CODE (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    CODE_ID VARCHAR(36) UNIQUE NOT NULL,
    AUTHORIZATION_CODE VARCHAR(500) NOT NULL,
    CONSUMER_KEY VARCHAR(255) NOT NULL,
    CALLBACK_URL VARCHAR(500),
    AUTHZ_USER VARCHAR(255) NOT NULL,
    TIME_CREATED DATETIME NOT NULL,
    EXPIRY_TIME DATETIME NOT NULL,
    STATE VARCHAR(50) NOT NULL
);

-- Table to store scopes associated with OAuth2 authorization codes.
CREATE TABLE IDN_OAUTH2_AUTHZ_CODE_SCOPE (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
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
    user_input_data TEXT,
    runtime_data TEXT,
    authenticated_user_attributes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
    required_data TEXT,
    additional_data TEXT,
    actions TEXT,
    runtime_data TEXT,
    authenticated_user_attributes TEXT
);

-- Table to store each flow graph
CREATE TABLE graphs (
    graph_id VARCHAR PRIMARY KEY,
    type VARCHAR NOT NULL,
    nodes TEXT,
    edges TEXT,
    start_node_id VARCHAR
);

-- Table to store each node in a graph
CREATE TABLE nodes (
    node_id VARCHAR PRIMARY KEY,
    graph_id VARCHAR NOT NULL REFERENCES graphs(graph_id) ON DELETE CASCADE,
    type VARCHAR NOT NULL,
    is_start_node BOOLEAN,
    is_final_node BOOLEAN,
    next_node_list TEXT,
    previous_node_list TEXT,
    input_data TEXT,
    executor_config TEXT
);
