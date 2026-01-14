// Authentication management
class AuthManager {
    constructor() {
        this.currentUser = null;
        this.init();
    }

    init() {
        // Load user info from localStorage if available
        this.loadUserInfoFromStorage();

        // Validate user info on startup
        this.validateUserInfo();

        // Setup event listeners
        this.setupEventListeners();
    }

    // Load user info from localStorage
    loadUserInfoFromStorage() {
        try {
            const userInfoStr = localStorage.getItem('userInfo');
            if (userInfoStr) {
                this.currentUser = JSON.parse(userInfoStr);
            }
        } catch (error) {
            // Invalid user info in storage, clear it
            localStorage.removeItem('userInfo');
            this.currentUser = null;
        }
    }

    // Validate user info - check if token exists and is valid
    async validateUserInfo() {
        const accessToken = localStorage.getItem('accessToken');
        
        if (!accessToken) {
            // No token, clear user info
            this.currentUser = null;
            localStorage.removeItem('userInfo');
            return;
        }

        // Check if token is expired
        if (isTokenExpired(accessToken)) {
            // Token expired, try to refresh
            const refreshed = await api.ensureValidToken();
            if (!refreshed) {
                // Refresh failed, clear everything
                this.clearUserInfo();
                return;
            }
        }

        // If we have user info but no token, or token is invalid, clear user info
        if (this.currentUser && !accessToken) {
            this.clearUserInfo();
        }
    }

    // Set user info (called after successful login/register)
    setUserInfo(user) {
        this.currentUser = {
            id: user.id,
            username: user.username,
            email: user.email,
        };
        localStorage.setItem('userInfo', JSON.stringify(this.currentUser));
    }

    // Clear user info
    clearUserInfo() {
        this.currentUser = null;
        localStorage.removeItem('userInfo');
    }

    // Fetch user info from backend (for /auth/me endpoint)
    async fetchUserInfo() {
        try {
            const response = await api.request('/auth/me');
            if (response.success && response.data) {
                this.setUserInfo(response.data);
                return this.currentUser;
            }
            return null;
        } catch (error) {
            // If endpoint doesn't exist or fails, that's okay
            // We'll rely on user info from login/register responses
            return null;
        }
    }

    // Called when tokens are refreshed
    async onTokenRefreshed() {
        // Optionally fetch fresh user info after token refresh
        // This ensures user info is up-to-date after token refresh
        try {
            await this.fetchUserInfo();
        } catch (error) {
            // If /auth/me endpoint doesn't exist or fails, that's okay
            // We'll keep existing user info from localStorage
        }
    }

    setupEventListeners() {
        // Login form
        const loginForm = document.getElementById('login-form-element');
        if (loginForm) {
            loginForm.addEventListener('submit', (e) => this.handleLogin(e));
        }

        // Register form
        const registerForm = document.getElementById('register-form-element');
        if (registerForm) {
            registerForm.addEventListener('submit', (e) => this.handleRegister(e));
        }

        // Toggle between login and register
        const showRegister = document.getElementById('show-register');
        const showLogin = document.getElementById('show-login');
        if (showRegister) {
            showRegister.addEventListener('click', () => {
                window.history.pushState({}, '', '/register');
                window.dispatchEvent(new PopStateEvent('popstate'));
            });
        }
        if (showLogin) {
            showLogin.addEventListener('click', () => {
                window.history.pushState({}, '', '/login');
                window.dispatchEvent(new PopStateEvent('popstate'));
            });
        }

        // Logout button
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => this.handleLogout());
        }
    }

    async handleLogin(e) {
        e.preventDefault();
        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;
        const errorDiv = document.getElementById('login-error');

        // Clear previous errors
        errorDiv.classList.add('hidden');
        errorDiv.textContent = '';

        try {
            const response = await api.login(username, password);
            
            if (response.success && response.data) {
                // Extract tokens from nested structure: data.token.access_token
                const tokenPair = response.data.token || {};
                const user = response.data.user || {};
                
                // Store tokens first - this is the source of truth for authentication
                api.setTokens(tokenPair.access_token, tokenPair.refresh_token);
                
                // Store user info using the setter method
                this.setUserInfo(user);
                
                // Navigate back to leaderboard
                window.history.pushState({}, '', '/');
                
                // Update UI immediately after token is stored
                if (app && typeof app.handleRoute === 'function') {
                    app.handleRoute();
                }
                
                // Update UI
                if (app && typeof app.updateAuthUI === 'function') {
                    app.updateAuthUI();
                }
            } else {
                throw new Error(response.message || 'Login failed');
            }
        } catch (error) {
            errorDiv.textContent = error.message || 'Login failed. Please try again.';
            errorDiv.classList.remove('hidden');
        }
    }

    async handleRegister(e) {
        e.preventDefault();
        const username = document.getElementById('register-username').value;
        const email = document.getElementById('register-email').value;
        const password = document.getElementById('register-password').value;
        const errorDiv = document.getElementById('register-error');

        // Clear previous errors
        errorDiv.classList.add('hidden');
        errorDiv.textContent = '';

        try {
            const response = await api.register(username, email, password);
            
            if (response.success && response.data) {
                // Extract tokens from nested structure: data.token.access_token
                const tokenPair = response.data.token || {};
                const user = response.data.user || {};
                
                // Store tokens first - this is the source of truth for authentication
                api.setTokens(tokenPair.access_token, tokenPair.refresh_token);
                
                // Store user info using the setter method
                this.setUserInfo(user);
                
                // Navigate back to leaderboard
                window.history.pushState({}, '', '/');
                
                // Update UI immediately after token is stored
                if (app && typeof app.handleRoute === 'function') {
                    app.handleRoute();
                }
                
                // Update UI
                if (app && typeof app.updateAuthUI === 'function') {
                    app.updateAuthUI();
                }
            } else {
                throw new Error(response.message || 'Registration failed');
            }
        } catch (error) {
            errorDiv.textContent = error.message || 'Registration failed. Please try again.';
            errorDiv.classList.remove('hidden');
        }
    }

    handleLogout() {
        api.clearTokens();
        this.clearUserInfo();
        window.history.pushState({}, '', '/');
        window.dispatchEvent(new PopStateEvent('popstate'));
        // Update UI after logout
        if (app && typeof app.updateAuthUI === 'function') {
            app.updateAuthUI();
        }
    }

    getCurrentUser() {
        return this.currentUser;
    }
}

const authManager = new AuthManager();
