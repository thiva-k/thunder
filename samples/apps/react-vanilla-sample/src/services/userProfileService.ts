// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

import config from '../config';

export interface UserProfile {
    id: string;
    ouId?: string;
    ouHandle?: string;
    type?: string;
    display?: string;
    isReadOnly?: boolean;
    attributes: Record<string, unknown>;
}

const usersEndpoint = new URL('/users', config.flowEndpoint).toString();

const getErrorMessage = async (response: Response, fallback: string) => {
    const errorData = await response.json().catch(() => ({})) as {
        message?: { defaultValue?: string } | string;
        description?: { defaultValue?: string } | string;
    };

    if (typeof errorData.message === 'string' && errorData.message) {
        return errorData.message;
    }

    if (typeof errorData.description === 'string' && errorData.description) {
        return errorData.description;
    }

    if (typeof errorData.message === 'object' && errorData.message?.defaultValue) {
        return errorData.message.defaultValue;
    }

    if (typeof errorData.description === 'object' && errorData.description?.defaultValue) {
        return errorData.description.defaultValue;
    }

    return fallback;
};

const normalizeUserProfile = (data: Partial<UserProfile>): UserProfile => ({
    id: data.id || '',
    ouId: data.ouId,
    ouHandle: data.ouHandle,
    type: data.type,
    display: data.display,
    isReadOnly: data.isReadOnly,
    attributes: data.attributes && typeof data.attributes === 'object' ? data.attributes : {},
});

export const getCurrentUserProfile = async (token: string): Promise<UserProfile> => {
    const response = await fetch(`${usersEndpoint}/me?include=display`, {
        method: 'GET',
        headers: {
            Accept: 'application/json',
            Authorization: `Bearer ${token}`,
        },
    });

    if (!response.ok) {
        throw new Error(await getErrorMessage(response, 'Failed to load user profile.'));
    }

    return normalizeUserProfile(await response.json() as Partial<UserProfile>);
};

export const updateCurrentUserProfile = async (
    token: string,
    attributes: Record<string, unknown>
): Promise<UserProfile> => {
    const response = await fetch(`${usersEndpoint}/me`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ attributes }),
    });

    if (!response.ok) {
        throw new Error(await getErrorMessage(response, 'Failed to update user profile.'));
    }

    return normalizeUserProfile(await response.json() as Partial<UserProfile>);
};

export const updateCurrentUserPassword = async (
    token: string,
    password: string
): Promise<void> => {
    const response = await fetch(`${usersEndpoint}/me/update-credentials`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
            attributes: {
                password,
            },
        }),
    });

    if (!response.ok) {
        throw new Error(await getErrorMessage(response, 'Failed to update password.'));
    }
};
