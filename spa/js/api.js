// API client for communicating with the backend
const API_BASE_URL = '/api/v1';

class API {
    constructor() {
        this.accessToken = localStorage.getItem('accessToken');
        this.refreshToken = localStorage.getItem('refreshToken');
    }

    async request(endpoint, options = {}) {
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
                    // Try to refresh token
                    const refreshed = await this.refreshAccessToken();
                    if (refreshed) {
                        // Retry original request
                        headers['Authorization'] = `Bearer ${this.accessToken}`;
                        const retryResponse = await fetch(url, { ...config, headers });
                        const retryData = await retryResponse.json();
                        return retryData;
                    }
                }
                throw new Error(data.message || data.error?.message || 'Request failed');
            }

            return data;
        } catch (error) {
            throw error;
        }
    }

    async refreshAccessToken() {
        if (!this.refreshToken) {
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

            if (response.ok && data.data) {
                this.accessToken = data.data.access_token;
                this.refreshToken = data.data.refresh_token;
                localStorage.setItem('accessToken', this.accessToken);
                localStorage.setItem('refreshToken', this.refreshToken);
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
