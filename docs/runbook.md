# Incident Runbook

Operational guide for the DevOps Lab application. Use this when something goes wrong.

---

## 🔴 App Returns 5xx Errors

**Symptoms**: Users see server errors, alert `alert-backend-errors-dev` fires.

**Diagnosis**:
```bash
# Check container status
az containerapp show \
  --name ca-backend-dev \
  --resource-group rg-devopslab-dev \
  --query "properties.runningStatus"

# Check recent logs
az containerapp logs show \
  --name ca-backend-dev \
  --resource-group rg-devopslab-dev \
  --tail 50

# Check if container is restarting
az containerapp revision list \
  --name ca-backend-dev \
  --resource-group rg-devopslab-dev \
  --query "[0].{name:name, status:properties.runningState, replicas:properties.replicas}"
```

**Common causes & fixes**:

| Cause | Fix |
|-------|-----|
| Bad deployment | Roll back: `az containerapp update --name ca-backend-dev --resource-group rg-devopslab-dev --image acrdevopslabdev.azurecr.io/backend:<previous-sha>` |
| Out of memory | Increase memory in `container_apps.tf`, run `terraform apply` |
| Panic in Go code | Check logs for stack trace, fix code, redeploy |

---

## 🔴 Container Keeps Restarting

**Symptoms**: Alert `alert-backend-restarts-dev` fires, app intermittently available.

**Diagnosis**:
```bash
# Check restart count
az containerapp revision list \
  --name ca-backend-dev \
  --resource-group rg-devopslab-dev \
  --query "[].{name:name, replicas:properties.replicas, state:properties.runningState}" -o table

# Check container logs for crash reason
az containerapp logs show \
  --name ca-backend-dev \
  --resource-group rg-devopslab-dev \
  --tail 100 --type system
```

**Common causes & fixes**:

| Cause | Fix |
|-------|-----|
| Image won't start | Check Dockerfile CMD/ENTRYPOINT, verify image builds locally |
| Wrong platform | Ensure `--platform linux/amd64` in Docker build (not arm64) |
| Missing env vars | Check `env` blocks in `container_apps.tf` |
| Health check failing | Verify `/api/health` works, check ingress config |

---

## 🟡 High Latency (Slow Responses)

**Symptoms**: Users report slow page loads, P95 latency > 2 seconds.

**Diagnosis**:
```bash
# Quick latency test
curl -w "DNS: %{time_namelookup}s\nConnect: %{time_connect}s\nTLS: %{time_appconnect}s\nTotal: %{time_total}s\n" \
  -o /dev/null -s https://ca-backend-dev.jollyflower-21c30f67.eastus.azurecontainerapps.io/api/health

# Check if containers are scaled to 0 (cold start)
az containerapp revision list \
  --name ca-backend-dev \
  --resource-group rg-devopslab-dev \
  --query "[0].properties.replicas" -o tsv
```

**Common causes & fixes**:

| Cause | Fix |
|-------|-----|
| Cold start (scale from 0) | Set `min_replicas = 1` in `container_apps.tf` (costs more) |
| CPU throttling | Increase `cpu` in container spec |
| nginx proxy overhead | Check nginx error logs in frontend container |

---

## 🟡 CI/CD Pipeline Failed

**Symptoms**: GitHub Actions shows red ✗, deployment didn't happen.

### CI Failed

```bash
# Check which job failed
gh run list --workflow ci.yml --limit 5

# View failed run logs
gh run view <run-id> --log-failed
```

| Job | Common fix |
|-----|-----------|
| Backend test | Run `cd backend && go test ./...` locally, fix failures |
| Frontend test | Run `cd frontend && npm test` locally, fix failures |
| Docker build | Check Dockerfile syntax, run `docker build` locally |

### CD Failed

```bash
# Check CD runs
gh run list --workflow cd.yml --limit 5

# View failed run
gh run view <run-id> --log-failed
```

| Step | Common fix |
|------|-----------|
| Azure Login | Verify `AZURE_CREDENTIALS` secret hasn't expired. Regenerate: `az ad sp create-for-rbac --name github-devops-lab-cd --role contributor --scopes /subscriptions/<sub>/resourceGroups/rg-devopslab-dev` |
| ACR Login | Verify `ACR_NAME` secret matches actual registry name |
| Image push | Check ACR is accessible: `az acr show --name acrdevopslabdev` |
| Deploy | Check Container App exists: `az containerapp show --name ca-backend-dev -g rg-devopslab-dev` |
| Health check | App may need more startup time; check container logs |

---

## 🟡 Frontend Can't Reach Backend

**Symptoms**: React app loads but API calls fail, CORS errors in browser console.

**Diagnosis**:
```bash
# Check BACKEND_URL env var on frontend
az containerapp show \
  --name ca-frontend-dev \
  --resource-group rg-devopslab-dev \
  --query "properties.template.containers[0].env" -o table

# Verify backend is reachable
curl -s https://ca-backend-dev.jollyflower-21c30f67.eastus.azurecontainerapps.io/api/health

# Check nginx proxy config (exec into container if needed)
az containerapp exec \
  --name ca-frontend-dev \
  --resource-group rg-devopslab-dev \
  --command "cat /etc/nginx/conf.d/default.conf"
```

**Common causes & fixes**:

| Cause | Fix |
|-------|-----|
| BACKEND_URL wrong | Update env in `container_apps.tf`, redeploy |
| Backend down | Fix backend first (see sections above) |
| SSL proxy error | Ensure `proxy_ssl_server_name on` in nginx.conf |
| CORS blocked | Check `corsMiddleware` in `backend/middleware.go` |

---

## 🔵 Terraform State Issues

**Symptoms**: `terraform plan` hangs or shows lock errors.

**Diagnosis**:
```bash
# Check if state is locked
az storage blob show \
  --account-name devopslabstate5b0cca \
  --container-name tfstate \
  --name devops-lab.tfstate \
  --query "properties.lease.status"
```

**Fixes**:

```bash
# Break a stuck lock (use carefully!)
terraform force-unlock <lock-id>

# If DNS issues prevent blob access, use ARM API
az rest --method GET \
  --url "https://management.azure.com/subscriptions/<sub>/resourceGroups/devops-lab-tfstate/providers/Microsoft.Storage/storageAccounts/devopslabstate5b0cca/blobServices/default/containers/tfstate?api-version=2023-01-01"
```

---

## 🔵 Cost Spike

**Symptoms**: Azure budget alert fires, unexpected charges.

**Actions**:
```bash
# Check what's running
az containerapp list --resource-group rg-devopslab-dev -o table

# Scale down to 0 (stop costs, keep config)
az containerapp update --name ca-backend-dev -g rg-devopslab-dev --min-replicas 0 --max-replicas 0
az containerapp update --name ca-frontend-dev -g rg-devopslab-dev --min-replicas 0 --max-replicas 0

# Nuclear option: destroy all resources (can recreate with terraform apply)
cd infra && terraform destroy
```

**Cost-saving tips**:
- Container Apps: `min_replicas = 0` means no cost when idle
- ACR Basic: $0.167/day flat rate
- Log Analytics: 5 GB/month free, then $2.76/GB
- Destroy resources when not actively learning: `terraform destroy`

---

## Escalation

If you can't resolve an issue:

1. Check Azure Service Health: https://status.azure.com/
2. Check GitHub Status: https://githubstatus.com/
3. Search Azure docs: https://learn.microsoft.com/azure/container-apps/
4. Check Terraform AzureRM provider issues: https://github.com/hashicorp/terraform-provider-azurerm/issues
