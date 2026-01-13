// Main application logic
class App {
    constructor() {
        this.validRoutes = ['/', '/login', '/register'];
        this.init();
    }

    init() {
        // Setup routing
        this.setupRouting();

        // Setup score submission form
        const scoreForm = document.getElementById('score-form');
        if (scoreForm) {
            scoreForm.addEventListener('submit', (e) => this.handleScoreSubmit(e));
        }

        // Start leaderboard connection when dashboard is shown
        this.observeDashboard();

        // Setup 404 page button
        const goHomeBtn = document.getElementById('go-home-btn');
        if (goHomeBtn) {
            goHomeBtn.addEventListener('click', () => {
                window.history.pushState({}, '', '/');
                this.handleRoute();
            });
        }
    }

    setupRouting() {
        // Handle initial route
        this.handleRoute();

        // Handle browser back/forward buttons
        window.addEventListener('popstate', () => {
            this.handleRoute();
        });
    }

    handleRoute() {
        const path = window.location.pathname;

        // Don't handle /api or /docs routes (these are backend routes)
        if (path.startsWith('/api') || path.startsWith('/docs') || path.startsWith('/js')) {
            return;
        }

        // Check if user is authenticated
        const isAuthenticated = !!localStorage.getItem('accessToken');

        if (path === '/' || path === '') {
            if (isAuthenticated) {
                this.showDashboard();
            } else {
                this.showLogin();
            }
        } else if (path === '/login') {
            this.showLogin();
        } else if (path === '/register') {
            this.showRegister();
        } else {
            // Invalid route - show 404
            this.show404();
        }
    }

    show404() {
        document.getElementById('auth-container').classList.add('hidden');
        document.getElementById('dashboard-container').classList.add('hidden');
        document.getElementById('not-found-container').classList.remove('hidden');
    }

    showLogin() {
        document.getElementById('auth-container').classList.remove('hidden');
        document.getElementById('dashboard-container').classList.add('hidden');
        document.getElementById('not-found-container').classList.add('hidden');
        document.getElementById('login-form').classList.remove('hidden');
        document.getElementById('register-form').classList.add('hidden');
        
        // Clear forms
        const loginForm = document.getElementById('login-form-element');
        const registerForm = document.getElementById('register-form-element');
        if (loginForm) loginForm.reset();
        if (registerForm) registerForm.reset();
        
        // Clear errors
        const loginError = document.getElementById('login-error');
        const registerError = document.getElementById('register-error');
        if (loginError) loginError.classList.add('hidden');
        if (registerError) registerError.classList.add('hidden');
    }

    showRegister() {
        document.getElementById('auth-container').classList.remove('hidden');
        document.getElementById('dashboard-container').classList.add('hidden');
        document.getElementById('not-found-container').classList.add('hidden');
        document.getElementById('login-form').classList.add('hidden');
        document.getElementById('register-form').classList.remove('hidden');
        
        // Clear errors
        const loginError = document.getElementById('login-error');
        const registerError = document.getElementById('register-error');
        if (loginError) loginError.classList.add('hidden');
        if (registerError) registerError.classList.add('hidden');
    }

    showDashboard() {
        document.getElementById('auth-container').classList.add('hidden');
        document.getElementById('dashboard-container').classList.remove('hidden');
        document.getElementById('not-found-container').classList.add('hidden');
        
        // Update user info
        const userInfo = document.getElementById('user-info');
        const currentUser = authManager.getCurrentUser();
        if (userInfo && currentUser) {
            userInfo.textContent = currentUser.username;
        }
    }

    observeDashboard() {
        // Use MutationObserver to detect when dashboard becomes visible
        const dashboard = document.getElementById('dashboard-container');
        if (!dashboard) return;

        const observer = new MutationObserver(() => {
            if (!dashboard.classList.contains('hidden')) {
                // Dashboard is visible, connect to leaderboard
                leaderboardManager.connect(100);
            } else {
                // Dashboard is hidden, disconnect
                leaderboardManager.disconnect();
            }
        });

        observer.observe(dashboard, {
            attributes: true,
            attributeFilter: ['class'],
        });

        // Also check initial state
        if (!dashboard.classList.contains('hidden')) {
            leaderboardManager.connect(100);
        }
    }

    async handleScoreSubmit(e) {
        e.preventDefault();
        const scoreInput = document.getElementById('score-input');
        const score = parseInt(scoreInput.value, 10);
        const errorDiv = document.getElementById('score-error');
        const successDiv = document.getElementById('score-success');

        // Clear previous messages
        errorDiv.classList.add('hidden');
        successDiv.classList.add('hidden');
        errorDiv.textContent = '';
        successDiv.textContent = '';

        if (isNaN(score) || score < 0) {
            errorDiv.textContent = 'Please enter a valid score (0 or greater)';
            errorDiv.classList.remove('hidden');
            return;
        }

        try {
            const response = await api.submitScore(score);
            
            if (response.success) {
                successDiv.textContent = `Score updated successfully! Your new score is ${response.data?.score || score}.`;
                successDiv.classList.remove('hidden');
                
                // Clear input
                scoreInput.value = '';
                
                // Hide success message after 3 seconds
                setTimeout(() => {
                    successDiv.classList.add('hidden');
                }, 3000);
            } else {
                throw new Error(response.message || 'Failed to update score');
            }
        } catch (error) {
            errorDiv.textContent = error.message || 'Failed to update score. Please try again.';
            errorDiv.classList.remove('hidden');
        }
    }
}

// Initialize app - must be initialized after auth.js
// This ensures routing is set up before auth manager tries to navigate
let app;
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        app = new App();
    });
} else {
    app = new App();
}
