// Leaderboard real-time updates via SSE
class LeaderboardManager {
    constructor() {
        this.eventSource = null;
        this.isConnected = false;
    }

    connect(limit = 100) {
        if (this.eventSource) {
            this.disconnect();
        }

        const url = `/api/v1/leaderboard/stream?limit=${limit}`;
        this.eventSource = new EventSource(url);

        this.eventSource.onopen = () => {
            this.isConnected = true;
            this.updateConnectionStatus(true);
        };

        this.eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (data.success && data.data) {
                    this.updateLeaderboard(data.data);
                }
            } catch (error) {
                console.error('Error parsing leaderboard data:', error);
            }
        };

        this.eventSource.onerror = (error) => {
            console.error('SSE connection error:', error);
            this.isConnected = false;
            this.updateConnectionStatus(false);
            
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

    updateLeaderboard(data) {
        const entries = data.entries || [];
        const total = data.total || 0;

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
        row.className = 'border-b border-slate-100 hover:bg-slate-50 transition-colors duration-150';

        // Rank cell
        const rankCell = document.createElement('td');
        rankCell.className = 'py-4 px-4 text-sm font-semibold text-slate-700';
        
        // Add medal emoji for top 3
        const rank = entry.rank || (index + 1);
        let rankText = String(rank);
        if (rank === 1) {
            rankText = '🥇 ' + rankText;
        } else if (rank === 2) {
            rankText = '🥈 ' + rankText;
        } else if (rank === 3) {
            rankText = '🥉 ' + rankText;
        }
        rankCell.textContent = rankText;
        row.appendChild(rankCell);

        // Username cell
        const usernameCell = document.createElement('td');
        usernameCell.className = 'py-4 px-4 text-sm text-slate-900';
        usernameCell.textContent = entry.username || 'Unknown';
        row.appendChild(usernameCell);

        // Score cell
        const scoreCell = document.createElement('td');
        scoreCell.className = 'py-4 px-4 text-sm font-semibold text-slate-900 text-right';
        scoreCell.textContent = entry.score?.toLocaleString() || '0';
        row.appendChild(scoreCell);

        return row;
    }
}

const leaderboardManager = new LeaderboardManager();
