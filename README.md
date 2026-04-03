# DevOps Lab 🚀

A full-stack task tracker application built to learn DevOps practices hands-on.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | React + TypeScript (Vite) |
| Backend | Go (standard library) |
| CI/CD | GitHub Actions |
| Infrastructure | Terraform |
| Cloud | Microsoft Azure |
| Containers | Docker |

## Project Structure

```
├── frontend/          → React + Vite app
├── backend/           → Go REST API
├── infra/             → Terraform (Azure)
├── .github/workflows/ → CI/CD pipelines
└── docker-compose.yml → Local development
```

## Getting Started

### Prerequisites

- [Node.js](https://nodejs.org/) (v20+)
- [Go](https://go.dev/) (1.22+)
- [Docker](https://www.docker.com/) (optional, for containerized dev)

### Run Locally

**Backend:**
```bash
cd backend
go run .
# API running at http://localhost:8080
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
# App running at http://localhost:5173
```

## DevOps Learning Path

This project progressively layers in DevOps practices:

1. ✅ Project scaffolding & Git workflow
2. 🔲 Docker containerization
3. 🔲 CI pipeline (GitHub Actions)
4. 🔲 Infrastructure as Code (Terraform)
5. 🔲 CD pipeline to Azure
6. 🔲 Monitoring & observability
7. 🔲 Security hardening
8. 🔲 Advanced topics & interview prep

## License

MIT
