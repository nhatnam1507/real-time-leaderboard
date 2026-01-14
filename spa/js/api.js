// API client for communicating with the backend
const API_BASE_URL = '/api/v1';

class API {
    constructor() {
        this.accessToken = localStorage.getItem('accessToken');
        this.refreshToken = localStorage.getItem('refreshToken');
        this.refreshPromise = null; // Track ongoing refresh to prevent multiple simultaneous refreshes
        this.pendingRequests = []; // Queue for requests waiting for token refresh
        this.refreshBufferMinutes = 5; // Refresh token if it expires within 5 minutes
    }

    // Update tokens from localStorage (call this when tokens are updated elsewhere)
    updateTokens() {
        this.accessToken = localStorage.getItem('accessToken');
        this.refreshToken = localStorage.getItem('refreshToken');
    }

    async request(endpoint, options = {}) {
        // Update tokens in case they were changed
        this.updateTokens();

        // Check if token needs refresh before making request
        if (this.accessToken && shouldRefreshToken(this.accessToken, this.refreshBufferMinutes)) {
            // Proactively refresh token
            const refreshed = await this.ensureValidToken();
            if (!refreshed) {
                // Token refresh failed, clear tokens and throw error
                this.clearTokens();
                throw new Error('Authentication failed. Please log in again.');
            }
        }

        // If there's an ongoing refresh, wait for it
        if (this.refreshPromise) {
            await this.refreshPromise;
        }

        const url = `${API_BASE_URL}${endpoint}`;
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers,
        };

        if (this.accessToken) {
            headers['Authorization'] = `Bearer ${this.accessToken}`;
        }

        const config = {
            ...options,
            headers,
        };

        try {
            const response = await fetch(url, config);
            const data = await response.json();

            if (!response.ok) {
                if (response.status === 401 && this.refreshToken) {
                    // Token expired or invalid, try to refresh
                    const refreshed = await this.ensureValidToken();
                    if (refreshed) {
                        // Retry original request with new token
                        headers['Authorization'] = `Bearer ${this.accessToken}`;
                        const retryResponse = await fetch(url, { ...config, headers });
                        const retryData = await retryResponse.json();
                        
                        if (!retryResponse.ok) {
                            // Still failed after refresh, might be refresh token expired
                            if (retryResponse.status === 401) {
                                this.clearTokens();
                                throw new Error('Session expired. Please log in again.');
                            }
                            throw new Error(retryData.message || retryData.error?.message || 'Request failed');
                        }
                        
                        return retryData;
                    } else {
                        // Refresh failed, clear tokens
                        this.clearTokens();
                        throw new Error('Session expired. Please log in again.');
                    }
                }
                throw new Error(data.message || data.error?.message || 'Request failed');
            }

            return data;
        } catch (error) {
            // If it's already our custom error, rethrow it
            if (error.message && (error.message.includes('Session expired') || error.message.includes('Authentication failed'))) {
                throw error;
            }
            // Otherwise wrap it
            throw new Error(error.message || 'Request failed');
        }
    }

    // Ensure token is valid, refresh if needed
    async ensureValidToken() {
        // If already refreshing, wait for that to complete
        if (this.refreshPromise) {
            try {
                await this.refreshPromise;
                return true;
            } catch (error) {
                return false;
            }
        }

        // Check if refresh is needed
        if (!this.refreshToken) {
            return false;
        }

        // Check if access token is expired or about to expire
        const needsRefresh = !this.accessToken || 
                            isTokenExpired(this.accessToken) || 
                            shouldRefreshToken(this.accessToken, this.refreshBufferMinutes);

        if (!needsRefresh) {
            return true; // Token is still valid
        }

        // Start refresh process
        this.refreshPromise = this.refreshAccessToken();
        
        try {
            const result = await this.refreshPromise;
            return result;
        } catch (error) {
            return false;
        } finally {
            this.refreshPromise = null;
        }
    }

    async refreshAccessToken() {
        if (!this.refreshToken) {
            return false;
        }

        // Check if refresh token itself is expired
        if (isTokenExpired(this.refreshToken)) {
            return false;
        }

        try {
            const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    refresh_token: this.refreshToken,
                }),
            });

            const data = await response.json();

            if (response.ok && data.data && data.data.token) {
                const tokenPair = data.data.token;
                this.setTokens(tokenPair.access_token, tokenPair.refresh_token);
                
                // Notify auth manager if it exists (fire-and-forget, don't block refresh)
                if (typeof authManager !== 'undefined' && authManager.onTokenRefreshed) {
                    authManager.onTokenRefreshed().catch(() => {
                        // Silently handle errors - user info fetch failure shouldn't break refresh
                    });
                }
                
                return true;
            }

            return false;
        } catch (error) {
            return false;
        }
    }

    setTokens(accessToken, refreshToken) {
        this.accessToken = accessToken;
        this.refreshToken = refreshToken;
        if (accessToken) {
            localStorage.setItem('accessToken', accessToken);
        }
        if (refreshToken) {
            localStorage.setItem('refreshToken', refreshToken);
        }
    }

    clearTokens() {
        this.accessToken = null;
        this.refreshToken = null;
        localStorage.removeItem('accessToken');
        localStorage.removeItem('refreshToken');
    }

    // Auth endpoints
    async register(username, email, password) {
        return this.request('/auth/register', {
            method: 'POST',
            body: JSON.stringify({ username, email, password }),
        });
    }

    async login(username, password) {
        return this.request('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password }),
        });
    }

    // Leaderboard endpoints
    async submitScore(score) {
        return this.request('/leaderboard/score', {
            method: 'PUT',
            body: JSON.stringify({ score }),
        });
    }
}

const api = new API();
