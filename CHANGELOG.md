# Changelog

All notable changes to the Meta Clash project (Go Backend & Next.js UI) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] - 2026-05-24

### Fixed
- **Deployment (Vercel Next.js Cache & Environment Variables)**: Resolved a critical deployment bug where Vercel was ignoring production environment variables (`NEXT_PUBLIC_API_URL`) and falling back to `localhost`.
  - **Root Cause**: An `.env.local` file was accidentally committed to the repository. Next.js natively merges `.env.local` during the build process, causing Vercel to silently override Dashboard variables with hardcoded `localhost` WebSocket endpoints.
  - **Resolution**: Deleted the overriding variables from `.env.local` and implemented a bulletproof hardcoded URL fallback inside `frontend/lib/config.js` to guarantee production connectivity independent of Vercel cache status.
- **WebSocket Trailing Slashes**: Implemented automatic trailing-slash stripping (`replace(/\/+$/, '')`) in `frontend/lib/config.js` to prevent double-slash WebSocket URLs (e.g., `//api/ws`). Double-slash URLs trigger HTTP 301 redirects in the Go backend, which natively breaks browser WebSocket handshakes (since WS cannot follow redirects).
- **Next.js Versioning on Vercel**: Downgraded Next.js from `15.x` to `14.2.15` in `package.json` to prevent runtime Serverless Function crashes (`FUNCTION_INVOCATION_FAILED`). Next.js 15 requires React 19, which caused fatal boot errors when attempting to run on React 18 within Vercel's Edge architecture.
- **Docker Standalone Builds**: Removed `output: 'standalone'` from `next.config.js`. While necessary for custom Docker deployments (like Render), it actively conflicts with Vercel's automated Serverless Function generation and causes the deployment to crash on execution.
- **Match History Tracking (WebSocket Auth)**: Fixed a bug where match statistics were not being linked to user profiles at the end of a game. Extracted the authenticated `UserID` from the JWT token via HTTP context in `backend/internal/ws/client.go` and applied it to in-game Player objects (`backend/internal/ws/handlers.go`), enabling correct tracking of wins and losses.
- **GitHub Language Detection**: Fixed GitHub Linguist incorrectly identifying the repository as 100% HTML by ignoring the `graphify-out` directory in a new `.gitattributes` file.

### Added
- **Multi-Universe Card Generation (5-Tier Resolver)**: Extended the card generation pipeline from anime-only to any fictional universe. New tiers:
  - **Superhero API** (`superhero_client.go`): Fetches 700+ Marvel/DC characters from the akabab CDN (free, no API key). Filters by publisher, team affiliation, and fuzzy name matching. Maps native powerstats (intelligence, strength, speed, etc.) directly to game attributes.
  - **Gemini Full-Generation** (`gemini_generator.go`): Universal fallback — asks Gemini 2.5 Flash to produce 24 character names + lore-accurate stats for any theme (Harry Potter, Star Wars, Lord of the Rings, etc.).
  - Full chain: Curated Packs → Jikan API → Superhero API → Gemini Full-Gen → Deterministic Hash.
- **Comprehensive Gitignore**: Expanded `.gitignore` to properly exclude frontend build artifacts (`.next/`, `out/`), backend Go binaries, local `.env` files, OS specific files, and agent tooling directories (`.gemini/`, `.cursor/`, etc.).

### Changed
- **Lobby UI Redesign**: Replaced the generic purple gradient and floating orb background with a dynamic, pure-CSS retro/synthwave aesthetic featuring an animated 3D grid, glowing horizon, and starry deep space to match the game's immersive, multiversal theme (`frontend/pages/index.js` and `frontend/styles/globals.css`).
- **Architecture Splitting**: Officially completed the migration away from a monolithic Render deployment. The architecture is now strictly split:
  - **Frontend**: Hosted purely statically on Vercel Edge Network (100% static HTML/JS, zero serverless functions).
  - **Backend**: Hosted as a persistent Go Docker container on Railway to entirely eliminate WebSocket cold-start latency.
- **Loading Screen UI**: Replaced development-centric loading screen facts (e.g., Go backend references) with immersive lore, gameplay rules, and "multiverse" world-building text to improve player experience (`frontend/pages/index.js`).

---

## [Unreleased] - 2026-05-23

### Fixed
- **Database (Guest & Bot UUID Hotfix)**: Added UUID validation checks to backend match-saving logic (`backend/internal/lobby/manager.go`) to prevent PostgreSQL UUID format errors when storing guest sessions and bot statistics.
- **Card UI Layout (Mobile Fit)**: Refactored card layout in `frontend/components/Card.js` to use a 2x2 grid for stats and a balanced 50% height split. This prevents attributes from getting cut off on mobile/small screen heights.
- **Card UI & Responsiveness**: Resolved card attribute clipping issues by adjusting container dimensions and padding ("Strict Boundary" scaling architecture). Fixed the active card scaling issue to ensure it does not obstruct UI text (`frontend/components/Card.js`).
- **Game UI Layout**: Eliminated the top player's card-profile overlap by implementing dynamic hand fanning (`frontend/components/PlayerSeat.js`).
- **Game Flow Snapping**: Eliminated UI "snapping" during bot-led round transitions through intelligent 15-second backend delays.
- **Auth (CORS)**: Fixed CORS issues in the registration route to properly handle preflight requests (`backend/internal/middleware/cors.go`).
- **WebSockets (Hotfix)**: Restored the `http.Hijacker` interface inside the `Logging` middleware (`backend/internal/middleware/logging.go`) to allow the Next.js frontend to successfully upgrade HTTP connections to WebSockets without throwing a 500 error.
- **Match History (Hotfix)**: Injected `UserRepository` into `LobbyManager` (`backend/internal/lobby/manager.go`) and updated `PlayRound` to correctly build and persist `models.Match` and `models.MatchPlayer` records when a lobby reaches `LobbyStateFinished`, properly updating the user's Profile page.
- **Logging Improvements**: Replaced unstructured print statements (`fmt.Printf`) with structured `slog.Error` calls in `LobbyManager` to capture faulty database writes cleanly in production logs.

### Added
- **Dynamic Config (Runtime API Injection)**: Created an API-driven configuration endpoint `/api/config` and wrapper utilities (`frontend/lib/config.js`, `frontend/pages/api/config.js`) to dynamically resolve backend and WebSocket URLs in the browser at runtime using container environment variables, removing hardcoded endpoints.
- **Central API Request Client**: Created `frontend/lib/api.js` to centralize all REST requests, default headers, bearer token injection, and unified HTTP 401 session clearance.
- **Environment Templates**: Added a safe, template-driven `.env.example` in the root and `frontend/.env.local` for local development setup.
- **Unit Testing**: Implemented a comprehensive test suite for `game.Engine` (`backend/internal/game/engine_test.go`) and `lobby.Manager` (`backend/internal/lobby/manager_test.go`) to test round resolution, card dealing, bot execution, and match saving, preventing regression in CI/CD pipelines.
- **Deployment**: Added Render.com deployment configuration (`render.yaml`) and adjusted the Next.js package build to properly build from the `frontend` subdirectory to resolve Render deployment errors.
- **Game Assets**: Added high-quality One Piece starter base deck card images (`frontend/public/images/onepiece/`) and a new dynamic favicon (`frontend/public/favicon.svg`).
- **Environment**: Upgraded system compiler to Go 1.26.
- **Environment**: Installed `air` (`github.com/air-verse/air`) for Go hot-reloading.
- **`backend/`**: Initialized Go module (`github.com/nikhilsaxena04/meta_clash/backend`).
- **AI Tooling**: Configured multi-agent workflow profiles in `.claude/` and `.agents/`, and added `.mcp.json` integration.
- **Code Visualization**: Integrated Graphify (`graphify-out/`) to map project architecture autonomously. (Recently updated to track the finalized architecture).
- **Documentation**: Established `README.md` and `CHANGELOG.md` to maintain AI context windows across sessions.
- **Models**: Created `internal/models/` (`card.go`, `lobby.go`, `player.go`, `user.go`) establishing domain boundaries and interfaces.
- **Config**: Implemented `internal/config/config.go` for environment variable loading.
- **Game & Networking**: Built `internal/game/` for deterministic combat and `internal/ws/` / `internal/lobby/` for real-time multiplayer via WebSockets.
- **Game Content**: Introduced starter base decks (`backend/internal/game/packs/onepiece.go` and `pokemon.go`).
- **Entrypoint**: Bootstrapped `cmd/server/main.go` for executing the backend server.
- **Database**: Added `internal/db/postgres.go` and `internal/db/queries.go` for PostgreSQL connection pooling, auto-migrations, and repository pattern implementation.
- **Auth**: Added `internal/auth/` containing JWT lifecycle management, register/login REST handlers, and authorization middleware.
- **Middleware**: Added `internal/middleware/` for structured HTTP logging (`slog`), panic recovery, and CORS configurations.
- **Dynamic Card Generation (Phase 3)**: Implemented Jikan API integration (`backend/internal/game/jikan_client.go`) with an in-memory LRU cache to prevent rate-limiting, and deterministic `hashbot` fallback logic for guaranteed stable stats.
- **DevOps & CI/CD (Phase 4)**: Created multi-stage Dockerfiles for Go and Next.js (standalone output). Added `docker-compose.yml` for unified local orchestration and a `.github/workflows/ci.yml` pipeline for automated tests and linting.

### Changed
- **Animation Enhancements**: Maintained and polished Framer Motion animations during gameplay, guaranteeing smooth state transitions and premium visual hierarchy.
- **Backend Config**: Updated backend server logic and Jikan Client configurations for robustness and seamless rate limit handling.
- **Architecture**: Officially began migration from the previous Node.js/JS implementation to a strongly typed Go backend. Created `feature/go-backend-rewrite` branch.
- **Frontend Networking**: Removed Node.js socket server (`pages/api/socket.js`) and implemented a native WebSocket client (`lib/ws.js`) to connect the React UI directly to the Go server.
- **Server**: Updated `cmd/server/main.go` to wire PostgreSQL, Auth endpoints, and HTTP middleware together.
- **Frontend Migration (Phase 2)**: Moved all Next.js code to `frontend/`. Implemented premium glassmorphism Auth UI (`login.js`, `register.js`, `profile.js`). Updated `ws.js` and `index.js` to securely pass JWT tokens.

### Pending (Future Enhancements)
- Migrate in-memory `LobbyStore` to Redis for horizontal scalability.
- AI-generated character images for the Gemini full-generation fallback path.
- Add game sound effects and more robust animations to the React frontend.
- Deploy to AWS/GCP using the provided Docker containers.