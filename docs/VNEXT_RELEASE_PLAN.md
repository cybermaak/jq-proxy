# vNext Release Plan

## Objective
Deliver the next release of JQ Proxy with a focus on production hardening, API contract stability, and better operational ergonomics.

## Inputs considered
- Existing product goals and feature set from README.
- Current API/documentation behavior.
- Open in-flight PRs:
  - https://github.com/cybermaak/jq-proxy/pull/2
  - https://github.com/cybermaak/jq-proxy/pull/3
  - https://github.com/cybermaak/jq-proxy/pull/4
  - https://github.com/cybermaak/jq-proxy/pull/5

## Release Theme
**Production hardening + API ergonomics**

---

## Milestones

### M1 — Reliability & Security Baseline
Target: first merge wave

#### 1) Configurable upstream timeout
- **Status:** In flight (PR #5)
- **Goal:** Remove hardcoded upstream timeout behavior and make timeout policy explicit and configurable.
- **Acceptance criteria:**
  - Upstream timeout comes from configuration.
  - Works with file and env config.
  - Tests cover default + override behavior.

#### 2) `/config` endpoint hardening
- **Status:** Planned
- **Goal:** Reduce risk of leaking internal targets/infra topology.
- **Scope:**
  - Add config flag to enable/disable `/config` endpoint.
  - Default disabled for production profile.
  - Add optional target redaction modes when enabled.
- **Acceptance criteria:**
  - Endpoint can be disabled.
  - Optional redaction works as documented.
  - Tests and docs updated.

---

### M2 — Operability & Contract Stability
Target: second merge wave

#### 3) Liveness/readiness split
- **Status:** In flight (PR #4)
- **Goal:** Improve orchestration health semantics.
- **Acceptance criteria:**
  - Distinct liveness/readiness endpoints.
  - Readiness reflects service/config readiness.
  - Behavior and status codes are documented.

#### 4) Metrics schema stabilization
- **Status:** In flight (PR #2)
- **Goal:** Provide stable machine-consumable observability contract.
- **Scope:**
  - Explicit response DTO for metrics.
  - Stable naming convention and units.
  - Include `schema_version` for compatibility management.
- **Acceptance criteria:**
  - Contract tests pin response shape.
  - Docs define units and compatibility expectations.

---

### M3 — API UX & Compatibility
Target: third merge wave

#### 5) Passthrough transformation mode
- **Status:** In flight (PR #3)
- **Goal:** Allow proxy-only operation without requiring jq query.
- **Acceptance criteria:**
  - `transformation_mode: none` supported.
  - `jq_query` required only in jq mode.
  - API docs/examples updated.

#### 6) Endpoint key normalization policy
- **Status:** Planned
- **Goal:** Reduce case-sensitivity and key-shape surprises in env-driven endpoint config.
- **Scope:**
  - Canonical normalization mode (e.g., lowercase-hyphen).
  - Compatibility mode to preserve current behavior.
  - Collision detection and clear startup errors.
- **Acceptance criteria:**
  - Normalization behavior deterministic and documented.
  - Migration path is explicit.
  - Tests cover collisions and compatibility mode.

---

## Non-goals for this release
- Full authentication/authorization framework for all routes.
- Advanced traffic controls (rate limiting/circuit breaker) beyond basic groundwork.

These can be proposed for the next roadmap increment after vNext stabilizes.

---

## Dependency map
- PR #2, #3, #4, #5 are already in flight and should be reconciled before starting any overlapping work.
- `/config` hardening and endpoint key normalization are currently the main uncovered core vNext items.

---

## Proposed merge order
1. PR #5 (timeout)
2. PR #4 (health/readiness)
3. PR #2 (metrics contract)
4. PR #3 (passthrough mode)
5. New PR: `/config` hardening
6. New PR: endpoint key normalization
7. Docs pass: release notes + migration guide

---

## Definition of Done (release-level)
- All vNext scoped items merged or explicitly deferred.
- Tests pass in CI for all touched modules.
- Documentation updated:
  - API reference
  - Configuration reference
  - Deployment notes
  - Migration/release notes
- Backward compatibility impacts are clearly documented.

---

## Risks and mitigations
- **Risk:** Behavior changes break existing clients.
  - **Mitigation:** Compatibility flags + migration notes + phased rollout.
- **Risk:** Observability contract changes break dashboards.
  - **Mitigation:** `schema_version` + documented transition.
- **Risk:** Timeout tuning regressions.
  - **Mitigation:** sensible defaults + test coverage + staged rollout.
