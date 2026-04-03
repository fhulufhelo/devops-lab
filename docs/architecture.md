# Architecture Overview

## System Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           GITHUB                                        │
│                                                                         │
│  ┌──────────────┐    PR merge    ┌──────────────┐   workflow_run        │
│  │   CI Pipeline │──────────────▶│  CD Pipeline  │                      │
│  │              │    (on main)   │              │                      │
│  │ • Go test    │               │ • Build amd64 │                      │
│  │ • Go vet     │               │   images      │                      │
│  │ • npm lint   │               │ • Push to ACR │                      │
│  │ • npm test   │               │ • Deploy apps │                      │
│  │ • Docker bld │               │ • Health check│                      │
│  └──────────────┘               └───────┬───────┘                      │
│                                         │                               │
│  ┌──────────────┐                       │                               │
│  │   Security   │                       │                               │
│  │ • Trivy scan │                       │                               │
│  │ • CodeQL     │                       │                               │
│  │ • Dependabot │                       │                               │
│  └──────────────┘                       │                               │
└─────────────────────────────────────────┼───────────────────────────────┘
                                          │
                    docker push           │  az containerapp update
              ┌───────────────────────────┼────────────────────┐
              ▼                           ▼                    │
┌─────────────────────────────────────────────────────────────────────────┐
│                     AZURE  (East US)                                    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Resource Group: rg-devopslab-dev                               │    │
│  │                                                                 │    │
│  │  ┌────────────────────┐                                         │    │
│  │  │  Azure Container   │      ┌────────────────────────────┐     │    │
│  │  │  Registry (ACR)    │      │  Container Apps Environment│     │    │
│  │  │                    │      │  cae-devopslab-dev          │     │    │
│  │  │  backend:sha-abc   │      │                            │     │    │
│  │  │  backend:latest    │─────▶│  ┌────────────────────┐    │     │    │
│  │  │  frontend:sha-abc  │      │  │ ca-backend-dev     │    │     │    │
│  │  │  frontend:latest   │─────▶│  │ Go API (:8080)     │    │     │    │
│  │  └────────────────────┘      │  │ 0.25 vCPU / 0.5 Gi│    │     │    │
│  │                              │  │ min: 0  max: 2     │    │     │    │
│  │                              │  └────────┬───────────┘    │     │    │
│  │                              │           │ HTTPS          │     │    │
│  │                              │  ┌────────┴───────────┐    │     │    │
│  │                              │  │ ca-frontend-dev    │    │     │    │
│  │                              │  │ nginx → React      │    │     │    │
│  │                              │  │ 0.25 vCPU / 0.5 Gi│    │     │    │
│  │                              │  │ /api/* → backend   │    │     │    │
│  │                              │  └────────────────────┘    │     │    │
│  │                              └────────────────────────────┘     │    │
│  │                                                                 │    │
│  │  ┌────────────────────┐      ┌────────────────────────────┐     │    │
│  │  │  Log Analytics     │◀─────│  Application Insights      │     │    │
│  │  │  log-devopslab-dev │      │  appi-devopslab-dev        │     │    │
│  │  │  30-day retention  │      │  • Request metrics         │     │    │
│  │  └────────────────────┘      │  • Error tracking          │     │    │
│  │                              └────────────────────────────┘     │    │
│  │                                                                 │    │
│  │  ┌────────────────────────────────────────────────────────┐     │    │
│  │  │  Alert Rules                                           │     │    │
│  │  │  • 5xx error count > 10 in 5 min → Severity 2         │     │    │
│  │  │  • Restart count  > 3  in 5 min → Severity 1          │     │    │
│  │  └────────────────────────────────────────────────────────┘     │    │
│  └─────────────────────────────────────────────────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐    │
│  │  Resource Group: devops-lab-tfstate                              │    │
│  │  Storage Account: devopslabstate5b0cca                          │    │
│  │  Container: tfstate → devops-lab.tfstate (remote state)         │    │
│  └─────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────┘
```

## Request Flow

```
User Browser
     │
     │  HTTPS
     ▼
┌──────────────────┐
│  ca-frontend-dev │
│  nginx:alpine    │
│                  │
│  GET /           │──▶ Serve React SPA (static files)
│  GET /api/*      │──▶ Reverse proxy ──┐
└──────────────────┘                    │
                                        │ HTTPS (SNI)
                                        ▼
                              ┌──────────────────┐
                              │  ca-backend-dev  │
                              │  Go API          │
                              │                  │
                              │  Middleware chain:│
                              │  1. Rate limiter │
                              │  2. Request ID   │
                              │  3. Logger       │
                              │  4. CORS         │
                              │        │         │
                              │        ▼         │
                              │  Route handlers  │
                              │  /api/health     │
                              │  /api/ready      │
                              │  /api/tasks CRUD │
                              └──────────────────┘
```

## CI/CD Pipeline Flow

```
Developer pushes code
        │
        ▼
┌─────────────────────────────────┐
│  Pull Request Created           │
│  Triggers: CI + Security        │
│                                 │
│  CI (4 jobs):                   │
│  ├── Backend Go 1.23 ─────┐    │
│  ├── Backend Go 1.24 ─────┤    │
│  ├── Frontend Node 20 ────┤    │
│  └── Docker Build ────────┘    │
│         all must pass ✓        │
│                                 │
│  Security (4 jobs):             │
│  ├── Trivy backend ───────┐    │
│  ├── Trivy frontend ──────┤    │
│  ├── CodeQL Go ───────────┤    │
│  └── CodeQL JS/TS ────────┘    │
│         report to Security tab │
└────────────┬────────────────────┘
             │  PR merged to main
             ▼
┌─────────────────────────────────┐
│  CI runs on main                │
│         │ passes                │
│         ▼                       │
│  CD triggers (workflow_run)     │
│  ├── Build amd64 images        │
│  ├── Tag with git SHA           │
│  ├── Push to ACR               │
│  ├── Deploy to Container Apps  │
│  └── Health check verification │
└─────────────────────────────────┘
```

## Infrastructure as Code

All Azure resources are defined in Terraform (`infra/` directory):

| File | Purpose |
|------|---------|
| `main.tf` | Provider config, remote backend, resource group |
| `variables.tf` | Input variables (project, env, location, subscription) |
| `acr.tf` | Azure Container Registry |
| `container_apps.tf` | Container Apps Environment + backend + frontend apps |
| `monitoring.tf` | Application Insights + alert rules |
| `budget.tf` | Cost management budget + alerts |
| `outputs.tf` | Useful outputs (URLs, resource names) |

### State Management

```
Terraform CLI (local) ──▶ Azure Blob Storage (remote state)
                          devopslabstate5b0cca/tfstate/devops-lab.tfstate
```

Remote state ensures:
- Team members share the same state
- State is locked during operations (prevents conflicts)
- State is backed up automatically
- Sensitive values are encrypted at rest
