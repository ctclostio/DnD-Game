# D&D Game Improvement Implementation Plan

This plan outlines the steps to address high-priority pending improvements, with a strong emphasis on modular design. This initial plan focuses on items identified in "Week 2 Focus" and "Quick Wins" from the `PENDING_IMPROVEMENTS.md` document.

## Phase 1: Backend Foundation Enhancements

This phase focuses on improving the stability, observability, and robustness of the backend systems.

### Task 1.1: Database Index Optimization
*   **Objective & Deliverables:**
    *   Objective: Improve database query performance for frequently accessed data.
    *   Deliverables:
        *   SQL migration scripts for adding new indexes (targeting `characters.user_id`, `game_sessions.status`, `game_participants.session_id`).
        *   Updated database schema documentation reflecting new indexes.
        *   Performance comparison (before/after) for key queries utilizing these indexes.
*   **Target Module(s):**
    *   Database schema (via migration scripts).
    *   Potentially `backend/internal/database/` if specific query adjustments are needed.
*   **Modularity Enhancement:**
    *   **Encapsulation:** Indexing strategy is encapsulated within the database layer.
    *   **Separation of Concerns:** Data persistence performance is optimized independently of business logic.
    *   **Minimized Dependencies:** Application code should not require changes.
*   **Prerequisites:**
    *   Access to the current database schema.
    *   Confirmation of critical query patterns.
*   **Acceptance Criteria:**
    *   Indexes are successfully applied.
    *   Measurable improvement (e.g., >20% reduction in execution time) for targeted queries.
    *   No regressions in application functionality.

### Task 1.2: Implement Graceful Shutdown
*   **Objective & Deliverables:**
    *   Objective: Ensure the application can shut down cleanly.
    *   Deliverables:
        *   Modified application entry point to handle OS signals (SIGINT, SIGTERM).
        *   Logic to coordinate shutdown of services.
        *   Logging indicating shutdown stages.
*   **Target Module(s):**
    *   Application main package/entry point (e.g., `backend/server/main.go`).
    *   Interfaces for services requiring explicit cleanup (e.g., `Shutdowner` interface).
*   **Modularity Enhancement:**
    *   **API Design:** Services implement a common `Shutdowner` interface.
    *   **Encapsulation:** Centralized shutdown orchestration; modules manage own cleanup.
    *   **Separation of Concerns:** Resource-specific cleanup logic within the owning module.
*   **Prerequisites:**
    *   Identification of all resources requiring explicit cleanup.
*   **Acceptance Criteria:**
    *   Application exits cleanly (exit code 0) on SIGINT/SIGTERM.
    *   In-flight requests given timeout to complete.
    *   Resources (DB connections, etc.) confirmed released.
    *   Logs show shutdown sequence and success/failure.

### Task 1.3: Add Health Check Endpoint & Dependencies
*   **Objective & Deliverables:**
    *   Objective: Provide an HTTP endpoint (e.g., `/healthz`) reporting application and dependency health.
    *   Deliverables:
        *   New HTTP handler for `/healthz`.
        *   Modular health check functions for critical dependencies.
        *   JSON response detailing overall and component statuses.
*   **Target Module(s):**
    *   New `health` package (e.g., `backend/internal/health/`).
    *   Modules for critical dependencies (e.g., `backend/internal/database/connection.go`) expose health check methods.
*   **Modularity Enhancement:**
    *   **API Design:** Modules implement `HealthCheckable` interface. Health handler aggregates results.
    *   **Separation of Concerns:** Health endpoint orchestrates; modules own health assessment.
    *   **Encapsulation:** Modules encapsulate their health check logic.
*   **Prerequisites:**
    *   Defined list of critical dependencies to monitor.
*   **Acceptance Criteria:**
    *   `/healthz` returns HTTP 200 when healthy.
    *   `/healthz` returns HTTP 503 if unhealthy.
    *   JSON response accurately reflects component statuses.

### Task 1.4: Add Environment Validation
*   **Objective & Deliverables:**
    *   Objective: Validate required environment variables at startup.
    *   Deliverables:
        *   Enhanced configuration loading module for validation.
        *   Application exits with clear error if config is missing/invalid.
*   **Target Module(s):**
    *   Configuration loading module (e.g., `config` package).
*   **Modularity Enhancement:**
    *   **Encapsulation:** Config loading/validation in a dedicated module.
    *   **Separation of Concerns:** Config management distinct from business logic.
    *   **Clear Boundaries:** Modules depend on `Config` struct/interface, not directly on env vars.
*   **Prerequisites:**
    *   Comprehensive list of environment variables and validation rules.
*   **Acceptance Criteria:**
    *   App fails to start with descriptive error if required env var is missing/invalid.
    *   App starts successfully with correct config.

### Task 1.5: Performance Monitoring Setup (Prometheus Integration)
*   **Objective & Deliverables:**
    *   Objective: Integrate Prometheus for collecting and exposing application metrics.
    *   Deliverables:
        *   Prometheus client library integrated.
        *   `/metrics` HTTP endpoint exposing metrics.
        *   Implementation of initial custom metrics (API request duration, WebSocket connections, AI latency, DB query performance, Cache hit rates).
*   **Target Module(s):**
    *   New `metrics` package (e.g., `backend/internal/metrics/`).
    *   Instrumentation points in `backend/internal/middleware/`, `backend/internal/services/`, `backend/internal/database/`, WebSocket logic.
*   **Modularity Enhancement:**
    *   **API Design:** `metrics` module provides API for other modules to register/update metrics.
    *   **Separation of Concerns:** Metrics collection setup centralized; instrumentation as cross-cutting concern.
    *   **Encapsulation:** Choice of monitoring system (Prometheus) encapsulated.
*   **Prerequisites:**
    *   Selection of Go Prometheus client library.
    *   Detailed specification of initial metrics.
*   **Acceptance Criteria:**
    *   `/metrics` endpoint available and serves Prometheus format data.
    *   Specified custom metrics collected and exposed correctly.
    *   Metrics can be scraped by Prometheus.
    *   Minimal performance overhead from metrics collection.

### Task 1.6: Request Correlation IDs (Backend Refinement/Verification)
*   **Objective & Deliverables:**
    *   Objective: Verify and ensure every request has a unique, consistently logged, and propagated correlation ID.
    *   Deliverables:
        *   Verified/enhanced middleware for correlation IDs.
        *   Ensured context propagation mechanism.
        *   Confirmation that all logs for a request flow contain the same correlation ID.
*   **Target Module(s):**
    *   `backend/internal/middleware/`.
    *   Logging infrastructure.
    *   Context utility functions.
*   **Modularity Enhancement:**
    *   **Separation of Concerns:** Correlation ID handling as a cross-cutting concern.
    *   **API Design:** `context.Context` for implicit propagation; utilities for clean access.
    *   **Encapsulation:** Business logic decoupled from correlation ID specifics.
*   **Prerequisites:**
    *   Review of existing structured logging and correlation ID implementation.
*   **Acceptance Criteria:**
    *   Correlation IDs present and consistent in all logs for a request flow.
    *   New ID generated if incoming request lacks one.

## Phase 2: Testing and Developer Experience

This phase focuses on improving test coverage and the tools available to developers.

### Task 2.1: Integration Test Implementation (Backend)
*   **Objective & Deliverables:**
    *   Objective: Develop integration tests for key API endpoints and business flows (Auth, Character CRUD, WebSocket, Game Session Lifecycle).
    *   Deliverables:
        *   Suite of integration tests.
        *   Test environment setup.
        *   Integration into CI/CD pipeline.
*   **Target Module(s):**
    *   New top-level `tests/integration` directory in `backend`.
    *   Test helper utilities.
*   **Modularity Enhancement:**
    *   **Black-Box Testing:** Tests interact via public APIs, reinforcing contracts.
    *   **Separation of Concerns:** Test code separate from application code.
    *   **Testable APIs:** Writing tests often improves API design for testability.
*   **Prerequisites:**
    *   Stable unit tests.
    *   Test data management strategy.
    *   Clear API specifications.
*   **Acceptance Criteria:**
    *   Tests cover critical API flows.
    *   Tests run reliably and repeatably.
    *   Clear feedback on failures.
    *   Test database/state properly managed.

### Task 2.2: API Documentation with Error Codes
*   **Objective & Deliverables:**
    *   Objective: Generate comprehensive API documentation including error codes.
    *   Deliverables:
        *   API documentation in standard format (e.g., OpenAPI/Swagger).
        *   Documentation includes all custom error codes and meanings.
        *   Documentation easily accessible.
*   **Target Module(s):**
    *   Annotations in `backend/internal/handlers/` (if using `swaggo/swag`).
    *   Source of truth for error codes from `AppError` definitions.
    *   `docs` directory for documentation.
*   **Modularity Enhancement:**
    *   **Clear API Contracts:** Fundamental to modularity.
    *   **Decoupling:** Clients develop against documented API.
    *   **Discoverability & Usability:** Easier API understanding and use.
*   **Prerequisites:**
    *   Stable standardized error handling system.
    *   Relatively stable API endpoints.
    *   Decision on documentation tool/format.
*   **Acceptance Criteria:**
    *   Documentation covers all public backend endpoints.
    *   Details request/response formats, status codes.
    *   All custom error codes listed with meanings.
    *   Documentation is accurate.

## Phase 3: Frontend Refinement

This phase includes quick wins for improving the frontend codebase.

### Task 3.1: Fix TypeScript `any` Types
*   **Objective & Deliverables:**
    *   Objective: Enhance frontend code quality by replacing `any` types with specific TypeScript types.
    *   Deliverables:
        *   Incremental PRs reducing `any` usage in `frontend/src/`.
        *   Definition of new interfaces/types in `frontend/src/types/` or alongside modules.
        *   Potentially stricter TypeScript compiler options.
*   **Target Module(s):**
    *   All `.ts`, `.tsx` files in `frontend/src/`.
    *   Shared type definitions in `frontend/src/types/`.
*   **Modularity Enhancement:**
    *   **Clear Interfaces:** Specific types clarify contracts between frontend parts.
    *   **Encapsulation:** Components become more robust with explicit data contracts.
    *   **Reduced Coupling:** Stronger typing reduces implicit dependencies.
*   **Prerequisites:**
    *   Developer familiarity with frontend codebase and TypeScript.
    *   Strategy for incremental changes.
*   **Acceptance Criteria:**
    *   Significant reduction in `any` type usage.
    *   New types/interfaces are accurate and used appropriately.
    *   No regressions in frontend functionality.
    *   TypeScript compiler reports fewer (or no new) `any`-related warnings/errors.