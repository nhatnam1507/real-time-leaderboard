// Leaderboard real-time updates via SSE
class LeaderboardManager {
    constructor() {
        this.eventSource = null;
        this.isConnected = false;
        this.leaderboardEntries = new Map(); // Map<userID, entry>
        this.totalPlayers = 0;
        this.limit = 10;
    }

    async loadInitialLeaderboard(limit = 10) {
        try {
            const response = await fetch(`/api/v1/leaderboard?limit=${limit}&offset=0`);
            const data = await response.json();
            
            if (data.success && data.data) {
                this.updateLeaderboard(data.data, data.meta);
            }
        } catch (error) {
            console.error('Error loading initial leaderboard:', error);
        }
    }

    async connect(limit = 10) {
        this.limit = limit;

        // Load initial leaderboard
        await this.loadInitialLeaderboard(limit);

        // Only connect stream if not already connected
        if (this.eventSource && this.isConnected) {
            // Stream already connected, just return
            return;
        }

        // Disconnect existing connection first and wait for it to close
        await this.disconnect();

        // Small delay to ensure old connection is fully closed
        await new Promise(resolve => setTimeout(resolve, 100));

        // Connect to SSE stream for delta updates (independent of limit)
        const url = `/api/v1/leaderboard/stream`;
        this.eventSource = new EventSource(url);

        // Check initial readyState - EventSource.OPEN = 1
        if (this.eventSource.readyState === EventSource.OPEN) {
            this.isConnected = true;
            this.updateConnectionStatus(true);
        } else {
            // Connection is still opening, will be updated in onopen
            this.isConnected = false;
        }

        this.eventSource.onopen = () => {
            this.isConnected = true;
            this.updateConnectionStatus(true);
        };

        this.eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.success && data.data) {
                    this.handleEntryUpdate(data.data);
                }
            } catch (error) {
                console.error('Error parsing leaderboard data:', error);
            }
        };

        this.eventSource.onerror = (error) => {
            // Only update status if connection is actually closed
            if (this.eventSource && this.eventSource.readyState === EventSource.CLOSED) {
                console.error('SSE connection error:', error);
                this.isConnected = false;
                this.updateConnectionStatus(false);
                
                // Reload leaderboard from API on disconnect
                this.loadInitialLeaderboard(this.limit);
                
                // Try to reconnect after a delay
                setTimeout(() => {
                    if (!this.isConnected) {
                        // Reconnect stream (not dependent on limit)
                        this.reconnectStream();
                    }
                }, 3000);
            }
        };
    }

    // Reconnect stream only (without calling /leaderboard API)
    async reconnectStream() {
        await this.disconnect();
        await new Promise(resolve => setTimeout(resolve, 100));

        const url = `/api/v1/leaderboard/stream`;
        this.eventSource = new EventSource(url);

        if (this.eventSource.readyState === EventSource.OPEN) {
            this.isConnected = true;
            this.updateConnectionStatus(true);
        } else {
            this.isConnected = false;
        }

        this.eventSource.onopen = () => {
            this.isConnected = true;
            this.updateConnectionStatus(true);
        };

        this.eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.success && data.data) {
                    this.handleEntryUpdate(data.data);
                }
            } catch (error) {
                console.error('Error parsing leaderboard data:', error);
            }
        };

        this.eventSource.onerror = (error) => {
            if (this.eventSource && this.eventSource.readyState === EventSource.CLOSED) {
                console.error('SSE connection error:', error);
                this.isConnected = false;
                this.updateConnectionStatus(false);
                
                setTimeout(() => {
                    if (!this.isConnected) {
                        this.reconnectStream();
                    }
                }, 3000);
            }
        };
    }

    // Update limit - calls /leaderboard API but keeps stream open
    async updateLimit(newLimit) {
        this.limit = newLimit;
        // Call /leaderboard API with new limit
        await this.loadInitialLeaderboard(newLimit);
        // Stream stays open - no need to reconnect
    }

    disconnect() {
        return new Promise((resolve) => {
            if (this.eventSource) {
                // Remove all event listeners to prevent callbacks during close
                this.eventSource.onopen = null;
                this.eventSource.onmessage = null;
                this.eventSource.onerror = null;
                
                // Close the connection
                this.eventSource.close();
                this.eventSource = null;
                this.isConnected = false;
                // Don't update status to disconnected here - we're about to reconnect
                // The status will be updated when the new connection opens
            }
            // Small delay to ensure connection is fully closed
            setTimeout(resolve, 50);
        });
    }

    updateConnectionStatus(connected) {
        const statusEl = document.getElementById('connection-status');
        if (statusEl) {
            const indicator = statusEl.querySelector('span');
            const text = statusEl.querySelectorAll('span')[1];
            
            if (connected) {
                indicator.className = 'w-2 h-2 rounded-full bg-green-500 animate-pulse';
                if (text) text.textContent = 'Live';
            } else {
                indicator.className = 'w-2 h-2 rounded-full bg-red-500';
                if (text) text.textContent = 'Disconnected';
            }
        }
    }

    handleEntryUpdate(entry) {
        // Always update the entry in local state (stream provides all updates)
        // The renderLeaderboard() will filter based on current limit
        this.leaderboardEntries.set(entry.user_id, entry);
        
        // Rebuild leaderboard from local state (will filter to top N)
        this.renderLeaderboard();
    }

    updateLeaderboard(data, meta) {
        const entries = Array.isArray(data) ? data : [];
        const total = meta?.total || 0;

        // Update local state
        this.leaderboardEntries.clear();
        entries.forEach(entry => {
            this.leaderboardEntries.set(entry.user_id, entry);
        });

        // Update total from pagination meta
        this.totalPlayers = total;

        // Render leaderboard
        this.renderLeaderboard();
    }

    renderLeaderboard() {
        // Convert map to array and sort by score descending (highest first)
        // Then recalculate ranks based on sorted position
        const allEntries = Array.from(this.leaderboardEntries.values())
            .sort((a, b) => {
                // Primary sort: by score descending
                if (b.score !== a.score) {
                    return b.score - a.score;
                }
                // Secondary sort: by user_id for consistency when scores are equal
                return a.user_id.localeCompare(b.user_id);
            });

        // Get top N entries with recalculated ranks
        // Keep all entries in local state - stream provides all updates
        // Just filter when rendering based on current limit
        const entries = allEntries
            .slice(0, this.limit)
            .map((entry, index) => {
                // Recalculate rank based on sorted position (1-indexed)
                return {
                    ...entry,
                    rank: index + 1
                };
            });

        const total = this.totalPlayers || allEntries.length;

        // Update stats
        const totalPlayersEl = document.getElementById('total-players');
        if (totalPlayersEl) {
            totalPlayersEl.textContent = total;
        }

        // Show/hide loading and empty states
        const loadingEl = document.getElementById('leaderboard-loading');
        const emptyEl = document.getElementById('leaderboard-empty');
        const contentEl = document.getElementById('leaderboard-content');
        const bodyEl = document.getElementById('leaderboard-body');

        if (loadingEl) loadingEl.classList.add('hidden');
        
        if (entries.length === 0) {
            if (emptyEl) emptyEl.classList.remove('hidden');
            if (contentEl) contentEl.classList.add('hidden');
            return;
        }

        if (emptyEl) emptyEl.classList.add('hidden');
        if (contentEl) contentEl.classList.remove('hidden');

        // Clear existing rows
        if (bodyEl) {
            bodyEl.innerHTML = '';
        }

        // Add entries
        entries.forEach((entry, index) => {
            const row = this.createLeaderboardRow(entry, index);
            if (bodyEl) {
                bodyEl.appendChild(row);
            }
        });
    }

    createLeaderboardRow(entry, index) {
        const row = document.createElement('tr');
        row.className = 'leaderboard-row border-b border-purple-500/20';

        // Rank cell
        const rankCell = document.createElement('td');
        rankCell.className = 'py-4 px-4';
        
        const rank = entry.rank || (index + 1);
        
        // Create rank badge
        const rankBadge = document.createElement('div');
        rankBadge.className = 'rank-badge';
        
        // Add special styling for top 3
        if (rank === 1) {
            rankBadge.classList.add('top-1');
        } else if (rank === 2) {
            rankBadge.classList.add('top-2');
        } else if (rank === 3) {
            rankBadge.classList.add('top-3');
        }
        
        rankBadge.textContent = rank;
        rankCell.appendChild(rankBadge);
        row.appendChild(rankCell);

        // Username cell
        const usernameCell = document.createElement('td');
        usernameCell.className = 'py-4 px-4 text-sm text-slate-200 font-medium';
        usernameCell.textContent = entry.username || 'Unknown';
        row.appendChild(usernameCell);

        // Score cell
        const scoreCell = document.createElement('td');
        scoreCell.className = 'py-4 px-4 text-sm font-bold text-purple-300 text-right';
        scoreCell.textContent = entry.score?.toLocaleString() || '0';
        row.appendChild(scoreCell);

        return row;
    }
}

const leaderboardManager = new LeaderboardManager();
