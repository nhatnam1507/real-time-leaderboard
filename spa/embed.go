// Package spa provides embedded SPA static files.
package spa

import (
	_ "embed"
)

//go:embed index.html
var IndexHTML []byte

//go:embed js/api.js
var APIJS []byte

//go:embed js/auth.js
var AuthJS []byte

//go:embed js/leaderboard.js
var LeaderboardJS []byte

//go:embed js/app.js
var AppJS []byte

//go:embed js/token-utils.js
var TokenUtilsJS []byte
