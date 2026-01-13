// Authentication management
class AuthManager {
    constructor() {
        this.currentUser = null;
        this.init();
    }

    init() {
        // Check if user is already logged in
        const accessToken = localStorage.getItem('accessToken');
        if (accessToken) {
            // Try to decode token to get user info (simple base64 decode)
            try {
                const payload = JSON.parse(atob(accessToken.split('.')[1]));
                this.currentUser = {
                    id: payload.user_id,
                    username: payload.username,
                };
            } catch (error) {
                // Invalid token, clear it
                api.clearTokens();
            }
        }

        // Setup event listeners
        this.setupEventListeners();
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
                api.setTokens(response.data.access_token, response.data.refresh_token);
                this.currentUser = {
                    id: response.data.user_id,
                    username: response.data.username,
                };
                window.history.pushState({}, '', '/');
                window.dispatchEvent(new PopStateEvent('popstate'));
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
                api.setTokens(response.data.access_token, response.data.refresh_token);
                this.currentUser = {
                    id: response.data.user_id,
                    username: response.data.username,
                };
                window.history.pushState({}, '', '/');
                window.dispatchEvent(new PopStateEvent('popstate'));
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
        this.currentUser = null;
        window.history.pushState({}, '', '/');
        window.dispatchEvent(new PopStateEvent('popstate'));
    }

    getCurrentUser() {
        return this.currentUser;
    }
}

const authManager = new AuthManager();
