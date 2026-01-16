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
        if (this.eventSource) {
            this.disconnect();
        }

        this.limit = limit;

        // Load initial leaderboard
        await this.loadInitialLeaderboard(limit);

        // Connect to SSE stream for delta updates
        const url = `/api/v1/leaderboard/stream`;
        this.eventSource = new EventSource(url);

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
            console.error('SSE connection error:', error);
            this.isConnected = false;
            this.updateConnectionStatus(false);
            
            // Reload leaderboard from API on disconnect
            this.loadInitialLeaderboard(limit);
            
            // Try to reconnect after a delay
            setTimeout(() => {
                if (!this.isConnected) {
                    this.connect(limit);
                }
            }, 3000);
        };
    }

    disconnect() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
            this.isConnected = false;
            this.updateConnectionStatus(false);
        }
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
        // Use the rank from the event message (authoritative from backend)
        // If rank is greater than limit, entry is outside top N - remove it
        if (entry.rank > this.limit) {
            // Entry is outside top N, remove it from local state
            this.leaderboardEntries.delete(entry.user_id);
            
            // Reload leaderboard to get the current top N from backend
            // This ensures we have the correct top N after an entry falls out
            this.loadInitialLeaderboard(this.limit);
            return;
        }
        
        // Entry is within top N, update it in local state
        this.leaderboardEntries.set(entry.user_id, entry);
        
        // Rebuild leaderboard from local state
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
        const entries = allEntries
            .slice(0, this.limit)
            .map((entry, index) => {
                // Recalculate rank based on sorted position (1-indexed)
                return {
                    ...entry,
                    rank: index + 1
                };
            });

        // Remove entries that are outside the top N from local state
        // This ensures entries that fall out of top N are removed
        const topNUserIDs = new Set(entries.map(e => e.user_id));
        for (const [userID] of this.leaderboardEntries) {
            if (!topNUserIDs.has(userID)) {
                this.leaderboardEntries.delete(userID);
            }
        }

        const total = this.totalPlayers || this.leaderboardEntries.size;

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
