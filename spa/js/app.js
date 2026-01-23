// Main application logic
class App {
    constructor() {
        this.validRoutes = ['/', '/login', '/register', '/profile'];
        this.init();
    }

    init() {
        // Ensure initial state - hide auth and profile containers by default
        const authContainer = document.getElementById('auth-container');
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const profileContainer = document.getElementById('profile-container');
        const notFoundContainer = document.getElementById('not-found-container');
        
        if (authContainer && !authContainer.classList.contains('hidden')) {
            authContainer.classList.add('hidden');
        }
        if (profileContainer && !profileContainer.classList.contains('hidden')) {
            profileContainer.classList.add('hidden');
        }
        if (notFoundContainer && !notFoundContainer.classList.contains('hidden')) {
            notFoundContainer.classList.add('hidden');
        }
        if (leaderboardContainer && leaderboardContainer.classList.contains('hidden')) {
            leaderboardContainer.classList.remove('hidden');
        }

        // Setup routing
        this.setupRouting();

        // Setup score submission form
        const scoreForm = document.getElementById('score-form');
        if (scoreForm) {
            scoreForm.addEventListener('submit', (e) => this.handleScoreSubmit(e));
        }

        // Setup profile score submission form
        const profileScoreForm = document.getElementById('profile-score-form');
        if (profileScoreForm) {
            profileScoreForm.addEventListener('submit', (e) => this.handleProfileScoreSubmit(e));
        }

        // Setup profile back button
        const profileBackBtn = document.getElementById('profile-back-btn');
        if (profileBackBtn) {
            profileBackBtn.addEventListener('click', () => {
                window.history.pushState({}, '', '/');
                this.handleRoute();
            });
        }

        // Setup auth back to leaderboard buttons
        const authBackBtn = document.getElementById('auth-back-to-leaderboard');
        if (authBackBtn) {
            authBackBtn.addEventListener('click', () => {
                window.history.pushState({}, '', '/');
                this.handleRoute();
            });
        }

        const authBackBtnRegister = document.getElementById('auth-back-to-leaderboard-register');
        if (authBackBtnRegister) {
            authBackBtnRegister.addEventListener('click', () => {
                window.history.pushState({}, '', '/');
                this.handleRoute();
            });
        }

        // Setup login prompt button
        const submitScoreBtn = document.getElementById('submit-score-btn');
        if (submitScoreBtn) {
            submitScoreBtn.addEventListener('click', () => {
                window.history.pushState({}, '', '/login');
                this.handleRoute();
            });
        }

        // Note: login-header-btn is always hidden - the score-submission-prompt has a sign-in button

        // Setup user dropdown menu
        this.setupUserDropdown();

        // Setup leaderboard limit selector
        this.setupLeaderboardLimit();

        // Start leaderboard connection immediately (leaderboard is always visible)
        const savedLimit = this.getLeaderboardLimit();
        leaderboardManager.connect(savedLimit);

        // Setup 404 page button
        const goHomeBtn = document.getElementById('go-home-btn');
        if (goHomeBtn) {
            goHomeBtn.addEventListener('click', () => {
                window.history.pushState({}, '', '/');
                this.handleRoute();
            });
        }

        // Update UI based on authentication state - call immediately and also after DOM is ready
        this.updateAuthUI();
        requestAnimationFrame(() => {
            this.updateAuthUI();
            // Call again after a short delay to ensure it runs
            setTimeout(() => {
                this.updateAuthUI();
            }, 100);
        });
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

        if (path === '/' || path === '') {
            // Default route - always show leaderboard
            this.showLeaderboard();
            // Ensure UI is updated after showing leaderboard
            setTimeout(() => {
                this.updateAuthUI();
            }, 50);
        } else if (path === '/login') {
            this.showLogin();
        } else if (path === '/register') {
            this.showRegister();
        } else if (path === '/profile') {
            // Check if user is authenticated
            const isAuthenticated = !!localStorage.getItem('accessToken');
            if (!isAuthenticated) {
                // Redirect to login if not authenticated
                window.history.pushState({}, '', '/login');
                this.handleRoute();
                return;
            }
            this.showProfile();
        } else {
            // Invalid route - show 404
            this.show404();
        }
    }

    show404() {
        const authContainer = document.getElementById('auth-container');
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const notFoundContainer = document.getElementById('not-found-container');
        const profileContainer = document.getElementById('profile-container');
        
        if (authContainer) authContainer.classList.add('hidden');
        if (leaderboardContainer) leaderboardContainer.classList.add('hidden');
        if (notFoundContainer) notFoundContainer.classList.remove('hidden');
        if (profileContainer) profileContainer.classList.add('hidden');
    }

    showLogin() {
        const authContainer = document.getElementById('auth-container');
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const notFoundContainer = document.getElementById('not-found-container');
        const profileContainer = document.getElementById('profile-container');
        const loginForm = document.getElementById('login-form');
        const registerForm = document.getElementById('register-form');
        
        if (authContainer) authContainer.classList.remove('hidden');
        if (leaderboardContainer) leaderboardContainer.classList.add('hidden');
        if (notFoundContainer) notFoundContainer.classList.add('hidden');
        if (profileContainer) profileContainer.classList.add('hidden');
        if (loginForm) loginForm.classList.remove('hidden');
        if (registerForm) registerForm.classList.add('hidden');
        
        // Clear forms
        const loginFormElement = document.getElementById('login-form-element');
        const registerFormElement = document.getElementById('register-form-element');
        if (loginFormElement) loginFormElement.reset();
        if (registerFormElement) registerFormElement.reset();
        
        // Clear errors
        const loginError = document.getElementById('login-error');
        const registerError = document.getElementById('register-error');
        if (loginError) loginError.classList.add('hidden');
        if (registerError) registerError.classList.add('hidden');
    }

    showRegister() {
        const authContainer = document.getElementById('auth-container');
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const notFoundContainer = document.getElementById('not-found-container');
        const profileContainer = document.getElementById('profile-container');
        const loginForm = document.getElementById('login-form');
        const registerForm = document.getElementById('register-form');
        
        if (authContainer) authContainer.classList.remove('hidden');
        if (leaderboardContainer) leaderboardContainer.classList.add('hidden');
        if (notFoundContainer) notFoundContainer.classList.add('hidden');
        if (profileContainer) profileContainer.classList.add('hidden');
        if (loginForm) loginForm.classList.add('hidden');
        if (registerForm) registerForm.classList.remove('hidden');
        
        // Clear errors
        const loginError = document.getElementById('login-error');
        const registerError = document.getElementById('register-error');
        if (loginError) loginError.classList.add('hidden');
        if (registerError) registerError.classList.add('hidden');
    }

    showLeaderboard() {
        const authContainer = document.getElementById('auth-container');
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const notFoundContainer = document.getElementById('not-found-container');
        const profileContainer = document.getElementById('profile-container');
        
        if (authContainer) authContainer.classList.add('hidden');
        if (leaderboardContainer) leaderboardContainer.classList.remove('hidden');
        if (notFoundContainer) notFoundContainer.classList.add('hidden');
        if (profileContainer) profileContainer.classList.add('hidden');
        
        // Update UI immediately
        this.updateAuthUI();
        
        // Update UI again after DOM is ready
        requestAnimationFrame(() => {
            this.updateAuthUI();
            // Also call it again after a short delay to ensure it runs
            setTimeout(() => {
                this.updateAuthUI();
            }, 100);
        });
    }

    async showProfile() {
        const authContainer = document.getElementById('auth-container');
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const notFoundContainer = document.getElementById('not-found-container');
        const profileContainer = document.getElementById('profile-container');
        
        if (authContainer) authContainer.classList.add('hidden');
        if (leaderboardContainer) leaderboardContainer.classList.add('hidden');
        if (notFoundContainer) notFoundContainer.classList.add('hidden');
        if (profileContainer) profileContainer.classList.remove('hidden');
        
        // Load profile data
        await this.loadProfileData();
    }

    async loadProfileData() {
        const currentUser = authManager.getCurrentUser();
        if (!currentUser) {
            return;
        }

        // Update profile info
        const profileAvatarText = document.getElementById('profile-avatar-text');
        const profileUsername = document.getElementById('profile-username');
        const profileEmail = document.getElementById('profile-email');
        
        if (profileAvatarText && currentUser.username) {
            profileAvatarText.textContent = currentUser.username.charAt(0).toUpperCase();
        }
        if (profileUsername) {
            profileUsername.textContent = currentUser.username || '-';
        }
        if (profileEmail) {
            profileEmail.textContent = currentUser.email || '-';
        }

        // Fetch user's current rank and score from leaderboard via SSE
        try {
            // Create a temporary EventSource to get leaderboard data
            const leaderboardData = await this.fetchLeaderboardData();
            
            if (leaderboardData && leaderboardData.entries) {
                const entries = leaderboardData.entries;
                // Find user entry by matching user_id
                const userEntry = entries.find(entry => entry.user_id === currentUser.id);
                
                const profileRank = document.getElementById('profile-rank');
                const profileScore = document.getElementById('profile-score');
                const profileTotalPlayers = document.getElementById('profile-total-players');
                
                if (userEntry) {
                    if (profileRank) {
                        profileRank.textContent = userEntry.rank || '-';
                    }
                    if (profileScore) {
                        profileScore.textContent = userEntry.score?.toLocaleString() || '0';
                    }
                } else {
                    // User not in top players, show defaults
                    if (profileRank) profileRank.textContent = '-';
                    if (profileScore) profileScore.textContent = '0';
                }
                
                if (profileTotalPlayers) {
                    // Total is now in pagination meta field, fallback to entries length
                    profileTotalPlayers.textContent = leaderboardData.meta?.total || entries.length;
                }
            }
        } catch (error) {
            console.error('Failed to load profile data:', error);
            // Set defaults
            const profileRank = document.getElementById('profile-rank');
            const profileScore = document.getElementById('profile-score');
            const profileTotalPlayers = document.getElementById('profile-total-players');
            
            if (profileRank) profileRank.textContent = '-';
            if (profileScore) profileScore.textContent = '0';
            if (profileTotalPlayers) profileTotalPlayers.textContent = '-';
        }
    }

    fetchLeaderboardData() {
        return new Promise((resolve, reject) => {
            const eventSource = new EventSource('/api/v1/leaderboard/stream?limit=100');
            const timeout = setTimeout(() => {
                eventSource.close();
                reject(new Error('Timeout fetching leaderboard data'));
            }, 5000);

            eventSource.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    if (data.success && data.data) {
                        clearTimeout(timeout);
                        eventSource.close();
                        resolve(data.data);
                    }
                } catch (error) {
                    clearTimeout(timeout);
                    eventSource.close();
                    reject(error);
                }
            };

            eventSource.onerror = (error) => {
                clearTimeout(timeout);
                eventSource.close();
                reject(error);
            };
        });
    }

    async handleProfileScoreSubmit(e) {
        e.preventDefault();
        
        const scoreInput = document.getElementById('profile-score-input');
        const score = parseInt(scoreInput.value, 10);
        const errorDiv = document.getElementById('profile-score-error');
        const successDiv = document.getElementById('profile-score-success');

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
                
                // Reload profile data to update rank and score
                await this.loadProfileData();
                
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

    updateAuthUI() {
        // Rely solely on access token in localStorage to determine authentication status
        // This is the single source of truth - if token exists, user is authenticated
        const accessToken = localStorage.getItem('accessToken');
        const isAuthenticated = !!accessToken;

        // Get all elements first
        const leaderboardContainer = document.getElementById('leaderboard-container');
        const loginHeaderBtn = document.getElementById('login-header-btn');
        const userDropdown = document.getElementById('user-dropdown');
        const scoreSubmissionCard = document.getElementById('score-submission-card');
        const scoreSubmissionPrompt = document.getElementById('score-submission-prompt');

        // Only update if leaderboard container exists
        if (!leaderboardContainer) {
            return;
        }
        
        // Don't update if leaderboard container is hidden (we're on login/register page)
        if (leaderboardContainer.classList.contains('hidden')) {
            return;
        }

        // Always hide the header Sign In button - the score-submission-prompt already has one
        if (loginHeaderBtn) {
            loginHeaderBtn.classList.add('hidden');
            loginHeaderBtn.style.setProperty('display', 'none', 'important');
            loginHeaderBtn.hidden = true;
        }

        // Get current user info from AuthManager (single source of truth)
        const currentUser = authManager.getCurrentUser();
        const username = currentUser ? (currentUser.username || '') : '';
        const email = currentUser ? (currentUser.email || '') : '';

        if (isAuthenticated) {
            // User is authenticated (token exists in localStorage) - show user dropdown
            if (userDropdown) {
                userDropdown.classList.remove('hidden');
                userDropdown.style.setProperty('display', '', 'important');
                userDropdown.hidden = false;
                
                // Update user info in dropdown
                const userNameEl = document.getElementById('user-name');
                const userAvatarText = document.getElementById('user-avatar-text');
                const dropdownUserName = document.getElementById('dropdown-user-name');
                const dropdownUserEmail = document.getElementById('dropdown-user-email');
                
                if (userNameEl && username) {
                    userNameEl.textContent = username;
                }
                if (userAvatarText && username) {
                    // Show first letter of username
                    userAvatarText.textContent = username.charAt(0).toUpperCase();
                }
                if (dropdownUserName && username) {
                    dropdownUserName.textContent = username;
                }
                if (dropdownUserEmail && email) {
                    dropdownUserEmail.textContent = email;
                } else if (dropdownUserEmail) {
                    dropdownUserEmail.textContent = '';
                }
            }
            
            // User has token - show score submission form, hide prompt
            // The presence of accessToken in localStorage determines this
            if (scoreSubmissionCard) {
                scoreSubmissionCard.classList.remove('hidden');
                scoreSubmissionCard.style.setProperty('display', '', 'important');
                scoreSubmissionCard.hidden = false;
            }
            if (scoreSubmissionPrompt) {
                scoreSubmissionPrompt.classList.add('hidden');
                scoreSubmissionPrompt.style.setProperty('display', 'none', 'important');
                scoreSubmissionPrompt.hidden = true;
            }
        } else {
            // User is not authenticated (no token in localStorage) - hide user dropdown
            if (userDropdown) {
                userDropdown.classList.add('hidden');
                userDropdown.style.setProperty('display', 'none', 'important');
                userDropdown.hidden = true;
                // Close dropdown menu if open
                const dropdownMenu = document.getElementById('user-dropdown-menu');
                if (dropdownMenu) {
                    dropdownMenu.classList.add('hidden');
                }
            }
            
            // No token in localStorage - hide score submission form, show prompt
            // The absence of accessToken in localStorage determines this
            // The prompt contains a "Sign In to Submit Score" button
            if (scoreSubmissionCard) {
                scoreSubmissionCard.classList.add('hidden');
                scoreSubmissionCard.style.setProperty('display', 'none', 'important');
                scoreSubmissionCard.hidden = true;
            }
            if (scoreSubmissionPrompt) {
                scoreSubmissionPrompt.classList.remove('hidden');
                scoreSubmissionPrompt.style.setProperty('display', '', 'important');
                scoreSubmissionPrompt.hidden = false;
            }
        }
    }

    setupUserDropdown() {
        const userMenuButton = document.getElementById('user-menu-button');
        const userDropdownMenu = document.getElementById('user-dropdown-menu');
        const profileLink = document.getElementById('profile-link');
        const updateScoreLink = document.getElementById('update-score-link');
        const logoutBtn = document.getElementById('logout-btn');

        // Toggle dropdown menu
        if (userMenuButton && userDropdownMenu) {
            userMenuButton.addEventListener('click', (e) => {
                e.stopPropagation();
                const isOpen = !userDropdownMenu.classList.contains('hidden');
                if (isOpen) {
                    userDropdownMenu.classList.add('hidden');
                    userMenuButton.setAttribute('aria-expanded', 'false');
                } else {
                    userDropdownMenu.classList.remove('hidden');
                    userMenuButton.setAttribute('aria-expanded', 'true');
                }
            });
        }

        // Close dropdown when clicking outside
        document.addEventListener('click', (e) => {
            if (userDropdownMenu && userMenuButton) {
                if (!userDropdownMenu.contains(e.target) && !userMenuButton.contains(e.target)) {
                    userDropdownMenu.classList.add('hidden');
                    if (userMenuButton) {
                        userMenuButton.setAttribute('aria-expanded', 'false');
                    }
                }
            }
        });

        // Profile link - navigate to profile page
        if (profileLink) {
            profileLink.addEventListener('click', (e) => {
                e.preventDefault();
                if (userDropdownMenu) {
                    userDropdownMenu.classList.add('hidden');
                }
                window.history.pushState({}, '', '/profile');
                this.handleRoute();
            });
        }

        // Logout button
        if (logoutBtn) {
            logoutBtn.addEventListener('click', (e) => {
                e.preventDefault();
                if (userDropdownMenu) {
                    userDropdownMenu.classList.add('hidden');
                }
                if (authManager && typeof authManager.handleLogout === 'function') {
                    authManager.handleLogout();
                }
            });
        }
    }

    setupLeaderboardLimit() {
        const limitSelector = document.getElementById('leaderboard-limit');
        if (!limitSelector) {
            return;
        }

        // Load saved limit from localStorage
        const savedLimit = this.getLeaderboardLimit();
        limitSelector.value = savedLimit;

        // Handle limit changes
        limitSelector.addEventListener('change', async (e) => {
            const newLimit = parseInt(e.target.value, 10);
            if (!isNaN(newLimit) && newLimit > 0) {
                // Save to localStorage
                this.setLeaderboardLimit(newLimit);
                
                // Update limit - calls /leaderboard API with new pagination
                // Stream stays open (independent)
                await leaderboardManager.updateLimit(newLimit);
            }
        });
    }

    getLeaderboardLimit() {
        const saved = localStorage.getItem('leaderboardLimit');
        if (saved) {
            const limit = parseInt(saved, 10);
            // Valid options: 5, 10, 50, 100
            const validLimits = [5, 10, 50, 100];
            if (!isNaN(limit) && limit > 0 && validLimits.includes(limit)) {
                return limit;
            }
        }
        return 10; // Default limit
    }

    setLeaderboardLimit(limit) {
        localStorage.setItem('leaderboardLimit', String(limit));
    }

    async handleScoreSubmit(e) {
        e.preventDefault();
        
        // Check if user is authenticated
        const isAuthenticated = !!localStorage.getItem('accessToken');
        if (!isAuthenticated) {
            // Redirect to login
            window.history.pushState({}, '', '/login');
            this.handleRoute();
            return;
        }

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
