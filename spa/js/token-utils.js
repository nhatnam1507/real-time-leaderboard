// Token utility functions for JWT token management
// These utilities only decode tokens for expiration checking, not for extracting user data

// Decode JWT token payload (for expiration checking only)
function decodeTokenPayload(token) {
    if (!token || typeof token !== 'string') {
        return null;
    }

    try {
        const parts = token.split('.');
        if (parts.length !== 3) {
            return null;
        }

        // Decode the payload (second part)
        const payload = JSON.parse(atob(parts[1]));
        return payload;
    } catch (error) {
        return null;
    }
}

// Get token expiration time in milliseconds
function getTokenExpirationTime(token) {
    const payload = decodeTokenPayload(token);
    if (!payload || !payload.exp) {
        return null;
    }

    // JWT exp is in seconds, convert to milliseconds
    return payload.exp * 1000;
}

// Check if token is expired
function isTokenExpired(token) {
    const expirationTime = getTokenExpirationTime(token);
    if (!expirationTime) {
        return true; // Consider invalid tokens as expired
    }

    // Add a small buffer (1 second) to account for clock skew
    return Date.now() >= expirationTime - 1000;
}

// Check if token should be refreshed (e.g., within 5 minutes of expiration)
function shouldRefreshToken(token, refreshBufferMinutes = 5) {
    const expirationTime = getTokenExpirationTime(token);
    if (!expirationTime) {
        return true; // Consider invalid tokens as needing refresh
    }

    const refreshBuffer = refreshBufferMinutes * 60 * 1000; // Convert to milliseconds
    const timeUntilExpiration = expirationTime - Date.now();

    // Refresh if token expires within the buffer time
    return timeUntilExpiration <= refreshBuffer;
}

// Get time until token expiration in seconds
function getTimeUntilExpiration(token) {
    const expirationTime = getTokenExpirationTime(token);
    if (!expirationTime) {
        return 0;
    }

    const timeUntilExpiration = expirationTime - Date.now();
    return Math.max(0, Math.floor(timeUntilExpiration / 1000));
}
