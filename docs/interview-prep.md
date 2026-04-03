# DevOps Interview Prep Guide

Everything you built in this project, organized by interview topic. Each section includes what you did, why it matters, and how to talk about it.

---

## How to Use This Guide

For each topic, you have:
- **What you built** — the concrete thing you can demo
- **Key decisions** — shows you understand trade-offs, not just tutorials
- **STAR story** — ready-to-use behavioral answer format
- **Likely follow-ups** — questions interviewers will ask next

---

## 1. CI/CD Pipelines

### What you built
- **CI pipeline** (`.github/workflows/ci.yml`): 4 parallel jobs — Go test (matrix: 1.23 + 1.24), React lint/test/build, Docker build verification
- **CD pipeline** (`.github/workflows/cd.yml`): Triggered by `workflow_run` after CI passes on main, builds amd64 images, pushes to ACR, deploys to Container Apps, runs health checks
- **Security pipeline** (`.github/workflows/security.yml`): Trivy container scanning + CodeQL SAST, uploads SARIF to GitHub Security tab

### Key decisions you made
| Decision | Why |
|----------|-----|
| Separate CI and CD workflows | CD uses `workflow_run` trigger — cleaner separation, CD only runs after CI succeeds |
| Matrix builds for Go | Tests against multiple Go versions, catches compatibility issues early |
| Git SHA image tags | Every deployment is traceable to exact commit, `latest` tag is convenience only |
| Concurrency control | `cancel-in-progress: true` for CI (fast feedback), `false` for CD (don't abort deploys) |

### STAR Story
> **Situation**: I was building a full-stack application and needed a reliable way to ship changes without breaking production.
>
> **Task**: Design a CI/CD pipeline that catches bugs early and deploys automatically with confidence.
>
> **Action**: I built a multi-stage pipeline in GitHub Actions. CI runs 4 parallel jobs (Go matrix tests, React tests, Docker builds) on every PR. Security scans run Trivy and CodeQL in parallel. When a PR merges to main and CI passes, CD automatically triggers — it builds linux/amd64 images, tags them with the git SHA, pushes to Azure Container Registry, deploys to Container Apps, and verifies with automated health checks.
>
> **Result**: Every merge to main is automatically deployed within 5 minutes. I caught a platform mismatch issue early (arm64 vs amd64) and a Vitest config breaking change through the pipeline. Zero manual deployments needed after setup.

### Likely follow-ups
- **"How do you handle rollbacks?"** → Redeploy the previous image tag: `az containerapp update --image backend:<previous-sha>`. Every image is tagged with its git SHA, so I can roll back to any commit.
- **"What if CI is slow?"** → I use path filters so Go jobs only run when `backend/` changes. Dependencies are cached. Docker layer caching speeds up builds. Matrix jobs run in parallel.
- **"How do you handle secrets in CI/CD?"** → GitHub Actions secrets for Azure credentials, ACR name, and resource group. Never in code or logs. Azure service principal has least-privilege scope (single resource group, Contributor role).

---

## 2. Docker & Containerization

### What you built
- **Go backend**: Multi-stage build — build in `golang:1.24-alpine`, run in `alpine:3.21` → ~14MB final image
- **React frontend**: Multi-stage build — build in `node:20-alpine`, serve with `nginx:alpine` → ~62MB final image
- **Docker Compose**: Local dev stack with frontend + backend + network bridge
- **nginx reverse proxy**: Template-based config with `envsubst` for dynamic `BACKEND_URL`

### Key decisions you made
| Decision | Why |
|----------|-----|
| Alpine base images | Smallest footprint, fewer CVEs than Debian-based images |
| Multi-stage builds | Build tools don't ship to production — smaller, more secure |
| `envsubst` for nginx | Same image works in Docker Compose (`http://backend:8080`) and Azure (HTTPS external URL) |
| `--platform linux/amd64` | Mac M-series builds arm64 by default, Azure needs amd64 |

### STAR Story
> **Situation**: I needed to package a Go API and React SPA into containers that work identically in local development and cloud production.
>
> **Task**: Build optimized Docker images that are small, secure, and environment-agnostic.
>
> **Action**: I used multi-stage Dockerfiles — the Go image is 14MB (build in golang:alpine, copy binary to alpine). The React app builds static files, then serves them via nginx:alpine. For the nginx proxy, I used envsubst templates so the same image dynamically configures the backend URL at startup — `http://backend:8080` locally, the HTTPS FQDN in Azure.
>
> **Result**: Same Docker images run in Docker Compose and Azure Container Apps with zero code changes. I also discovered and fixed a cross-platform issue — Docker on Mac builds arm64 images, but Azure Container Apps requires linux/amd64, so I added `--platform linux/amd64` to all CI/CD builds.

### Likely follow-ups
- **"Why not scratch instead of alpine?"** → Alpine gives me a shell for debugging (`exec` into container), a package manager if I need to install tools, and DNS resolution. For production I might use `distroless`, but alpine is the right trade-off for a learning project.
- **"How do you handle secrets in containers?"** → Never baked into images. Passed as environment variables at runtime. In Azure, Container Apps references secrets from the app config, not plaintext env vars.
- **"CMD vs ENTRYPOINT?"** → ENTRYPOINT is the executable that always runs. CMD provides default arguments. I use CMD for flexibility — you can override it without rebuilding the image.

---

## 3. Infrastructure as Code (Terraform)

### What you built
- **10 Azure resources** defined in Terraform: Resource Group, ACR, Container Apps Environment, 2 Container Apps, Log Analytics, Application Insights, 2 Alert Rules, Budget
- **Remote state** in Azure Blob Storage (separate resource group)
- **Variables and outputs** for reusability

### Key decisions you made
| Decision | Why |
|----------|-----|
| Remote state in Azure Blob | Team-safe, locked during operations, encrypted at rest |
| Separate state resource group | State storage survives `terraform destroy` of app resources |
| Container Apps over AKS | Cheaper, simpler, sufficient for this workload — no cluster management |
| ACR Basic with admin auth | Simplest integration with Container Apps, appropriate for dev |
| `terraform.tfvars` gitignored | Subscription IDs and secrets stay out of version control |

### STAR Story
> **Situation**: I needed to provision cloud infrastructure reproducibly — spinning up and tearing down environments for learning without manual Azure Portal clicks.
>
> **Task**: Define all infrastructure as code so any environment can be created or destroyed with a single command.
>
> **Action**: I wrote Terraform configs for the full Azure stack — Container Registry, Container Apps Environment, two Container Apps, Log Analytics, Application Insights, and monitoring alerts. I set up remote state in Azure Blob Storage in a separate resource group so it survives `terraform destroy`. I chose Container Apps over AKS because it's serverless (no cluster to manage) and has a generous free tier — the right tool for this workload size.
>
> **Result**: I can run `terraform apply` to create the entire environment in about 3 minutes, and `terraform destroy` to tear it down. Infrastructure changes go through the same PR review process as application code. I saved significant money by destroying resources when not actively learning.

### Likely follow-ups
- **"How do you handle Terraform state conflicts?"** → Remote state with Azure Blob Storage provides locking. If a lock gets stuck (e.g., killed terraform plan), I use `terraform force-unlock`.
- **"Why not ARM templates or Bicep?"** → Terraform is cloud-agnostic and has a larger ecosystem. Skills transfer to AWS/GCP. HCL is more readable than JSON (ARM) for infrastructure.
- **"How do you manage different environments?"** → Variables for environment name (`dev`, `staging`, `prod`). All resource names include the environment. Different `terraform.tfvars` per environment.

---

## 4. Monitoring & Observability

### What you built
- **Structured logging** with Go's `log/slog` — JSON format with method, path, status, duration_ms, bytes, request_id
- **Correlation IDs** via `X-Request-ID` header — generated if missing, propagated through the system
- **Health checks**: `/api/health` (liveness) and `/api/ready` (readiness)
- **Application Insights** connected to Log Analytics (30-day retention)
- **Alert rules**: 5xx error rate and container restart count

### Key decisions you made
| Decision | Why |
|----------|-----|
| `slog` over third-party logger | Standard library, zero dependencies, structured JSON out of the box |
| Separate health and readiness | Health = "process alive", Ready = "can serve traffic". Different use cases for orchestrators |
| Correlation IDs as middleware | Every log line is traceable across services without code changes in handlers |
| Alert on 5xx not 4xx | 4xx is client error (expected). 5xx is server bug (needs attention) |

### STAR Story
> **Situation**: After deploying to Azure, I had no visibility into how the application was behaving in production.
>
> **Task**: Implement observability so I can diagnose issues without SSH-ing into containers.
>
> **Action**: I replaced Go's default logger with `slog` for structured JSON logging — every request logs method, path, status code, duration, and a correlation ID. I added health and readiness endpoints. On the infrastructure side, I provisioned Application Insights and created alert rules for 5xx error spikes and container restarts using Terraform.
>
> **Result**: I can trace any request through the system using its correlation ID. When the backend returns errors, I get alerted automatically. The structured logs are queryable in Azure Log Analytics — I can filter by status code, latency percentile, or specific request IDs.

### Likely follow-ups
- **"What are RED metrics?"** → Rate (requests/sec), Errors (error rate), Duration (latency). The three signals that tell you if a service is healthy.
- **"Health check vs readiness check?"** → Health check: "is the process running?" If it fails, restart the container. Readiness: "can it serve traffic?" If it fails, stop sending requests but don't restart — it might be warming up.
- **"How would you add distributed tracing?"** → Add OpenTelemetry SDK, propagate trace context via W3C Trace Context headers, export spans to Application Insights. The correlation ID middleware I built is the foundation for this.

---

## 5. Security

### What you built
- **Dependabot** — auto-creates PRs for vulnerable dependencies (Go, npm, Docker, GitHub Actions)
- **Trivy** — scans container images for CVEs (CRITICAL + HIGH), reports to GitHub Security tab
- **CodeQL** — SAST analysis for Go and TypeScript, catches security anti-patterns
- **Rate limiting** — 100 req/min per IP with sliding window, returns 429 + Retry-After header
- **Least-privilege service principal** — Contributor role scoped to single resource group

### Key decisions you made
| Decision | Why |
|----------|-----|
| Trivy in report-only mode | Alpine base images always have some CVEs — blocking deploys on known/accepted CVEs wastes time. Report and triage instead |
| CodeQL for SAST | Free, integrated with GitHub, catches real bugs (SQL injection, path traversal, etc.) |
| IP-based rate limiting | Simple and effective for API protection. Would use API keys or JWT for production |
| Service principal scoped to resource group | If credentials leak, blast radius is limited to one resource group |

### STAR Story
> **Situation**: The application was functional but had no security scanning or runtime protection.
>
> **Task**: Add defense-in-depth — catch vulnerabilities before deployment and protect the API at runtime.
>
> **Action**: I added three layers of security scanning: Dependabot for dependency updates, Trivy for container CVE scanning, and CodeQL for static analysis. All run automatically on every PR. At runtime, I implemented rate limiting middleware in Go (100 req/min per IP with sliding window). For Azure access, the CD pipeline uses a service principal with Contributor role scoped to a single resource group — not subscription-wide.
>
> **Result**: The security pipeline runs 4 jobs in parallel alongside CI's 4 jobs — 8 total checks on every PR. Dependabot auto-creates PRs when vulnerabilities are found in dependencies. The rate limiter returns 429 with Retry-After header, protecting against basic abuse. All secrets are in GitHub Secrets or Azure config, never in code.

### Likely follow-ups
- **"What's the difference between SAST and DAST?"** → SAST (Static) analyzes source code without running it — finds code-level bugs. DAST (Dynamic) tests the running application — finds runtime vulnerabilities like open ports or misconfigurations. I use SAST (CodeQL). For DAST, I'd add OWASP ZAP.
- **"How do you handle secrets?"** → Layered approach: `.tfvars` is gitignored, GitHub Actions secrets for CI/CD, Azure Container Apps secret references for runtime. Never in code, Docker images, or logs.
- **"What about authentication?"** → This project uses CORS + rate limiting. For production: add JWT authentication, OAuth 2.0 with Azure AD, or API key validation. The middleware pattern makes it easy to add — just add another middleware to the chain.

---

## 6. Git Workflow & Collaboration

### What you built
- **Branch protection** on `main` — PRs required, status checks must pass
- **PR template** with checklist (tests, docs, breaking changes)
- **Trunk-based development** — short-lived feature branches, squash merge
- **Automated checks** — 8 CI/Security jobs gate every PR

### STAR Story
> **Situation**: I needed a workflow that enforces quality without slowing down development.
>
> **Task**: Set up branch protection and automated quality gates.
>
> **Action**: I configured `main` as a protected branch requiring PR reviews and passing status checks. Every PR runs 8 automated checks (4 CI + 4 Security). I use trunk-based development with short-lived feature branches and squash merges to keep history clean. The PR template ensures reviewers check for tests, documentation, and breaking changes.
>
> **Result**: No code reaches main without passing all 8 checks. Every commit on main is deployable. The squash merge strategy keeps the git history readable — one commit per feature, not dozens of WIP commits.

---

## 7. Cost Management (Azure)

### What you built
- **Cost tagging** — every resource tagged with project, environment, team, managed_by
- **Budget alert** — $50/month budget with alerts at 50%, 80%, and 100%
- **Scale-to-zero** — Container Apps configured with `min_replicas = 0`

### Key talking points
- "I can filter Azure Cost Analysis by tag to see exactly what each project/environment costs."
- "Container Apps with min_replicas=0 means I pay nothing when there's no traffic."
- "I destroy resources when not learning — `terraform destroy` takes 2 minutes, `terraform apply` recreates everything."
- "ACR Basic is $5/month. Container Apps are consumption-based. Log Analytics has 5 GB/month free. Total dev cost: ~$5-10/month when actively using."

---

## General Interview Tips

### Do's
- **Show the repo** — "Here's the actual pipeline running in production." Live demos beat slides.
- **Talk about failures** — "I hit an arm64/amd64 mismatch issue..." shows real experience.
- **Explain trade-offs** — "I chose Container Apps over AKS because..." shows engineering judgment.
- **Use numbers** — "14MB image", "8 automated checks", "5-minute deploy", "100 req/min rate limit."

### Don'ts
- Don't say "I just followed a tutorial." Say "I designed and built..."
- Don't memorize definitions. Explain what you built and why.
- Don't claim expertise you can't demo. Everything here is real and running.

### Common Questions You Should Prepare For

| Question | Your answer source |
|----------|-------------------|
| "Walk me through your CI/CD pipeline" | Section 1 STAR story + live demo |
| "How do you handle infrastructure?" | Section 3 + `terraform plan` demo |
| "What monitoring do you have?" | Section 4 + show Application Insights |
| "How do you handle secrets?" | Section 5 + show GitHub Secrets config |
| "What happens when a deployment fails?" | Runbook + rollback with image tags |
| "How do you optimize Docker images?" | Section 2 + show multi-stage Dockerfile |
| "Tell me about a technical challenge you solved" | arm64/amd64 story or nginx proxy/SNI story |

### Your Strongest Interview Story

The **nginx reverse proxy + SSL/SNI** issue is your best "debugging in production" story:

> "In local Docker Compose, the frontend proxied to the backend using the Docker network hostname. In Azure Container Apps, that hostname didn't exist. I switched to the external FQDN, but that uses HTTPS — and nginx failed with an SSL handshake error. I discovered that Server Name Indication (SNI) is required for Azure Container Apps' shared load balancer, and nginx doesn't send SNI by default. I fixed it with `proxy_ssl_server_name on` and `proxy_set_header Host $proxy_host`. This taught me about TLS SNI, reverse proxy configuration, and the difference between Docker networking and cloud networking."

This story hits everything interviewers want: real problem → systematic debugging → technical depth → lesson learned.
