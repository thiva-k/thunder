// Copyright 2025 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

const response = await fetch('/runtime.json');
const runtimeConfig = await response.json();

const config = {
    applicationID: runtimeConfig.applicationID || import.meta.env.VITE_REACT_APP_AUTH_APP_ID,
    flowEndpoint: runtimeConfig.flowEndpoint || import.meta.env.VITE_REACT_APP_SERVER_FLOW_ENDPOINT,
};

export default config;
