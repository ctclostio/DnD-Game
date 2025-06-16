# Kubernetes Deployment Guide

## Overview

This directory contains Kubernetes manifests for deploying the D&D Game application to a Kubernetes cluster. The deployment uses Kustomize for managing different environments and includes production-ready configurations.

## Architecture

```
┌─────────────┐     ┌─────────────┐
│   Ingress   │────▶│  Frontend   │
│  Controller │     │   (Nginx)   │
└─────────────┘     └─────────────┘
       │                    
       │            ┌─────────────┐     ┌─────────────┐
       └───────────▶│   Backend   │────▶│  PostgreSQL │
                    │   (API)     │     └─────────────┘
                    └─────────────┘            │
                           │            ┌─────────────┐
                           └───────────▶│    Redis    │
                                       └─────────────┘
```

## Prerequisites

1. Kubernetes cluster (1.24+)
2. kubectl installed and configured
3. Kustomize (built into kubectl 1.14+)
4. Ingress controller (nginx-ingress recommended)
5. cert-manager (for automatic TLS)
6. Container registry access

## Quick Start

### 1. Update Configuration

Edit the following files with your specific values:

```bash
# Update image registry in kustomization.yaml
vim k8s/kustomization.yaml

# Update domain names in ingress.yaml
vim k8s/ingress.yaml

# Create secrets file (DO NOT COMMIT)
cat > k8s/secrets.env << EOF
DB_PASSWORD=your-secure-password
REDIS_PASSWORD=your-redis-password
JWT_SECRET=your-jwt-secret-min-32-chars
AI_API_KEY=your-openai-api-key
EOF
```

### 2. Deploy to Development

```bash
# Create development namespace
kubectl create namespace dnd-game-dev

# Deploy using kustomize
kubectl apply -k k8s/overlays/dev

# Check deployment status
kubectl -n dnd-game-dev get pods
kubectl -n dnd-game-dev get svc
kubectl -n dnd-game-dev get ingress
```

### 3. Deploy to Production

```bash
# Create production namespace
kubectl create namespace dnd-game-prod

# Create production secrets (use real values)
kubectl create secret generic dnd-game-secrets \
  --from-literal=DB_PASSWORD='secure-password' \
  --from-literal=REDIS_PASSWORD='redis-password' \
  --from-literal=JWT_SECRET='jwt-secret-min-32-chars' \
  --from-literal=AI_API_KEY='openai-api-key' \
  -n dnd-game-prod

# Deploy using kustomize
kubectl apply -k k8s/overlays/prod

# Check deployment
kubectl -n dnd-game-prod get pods
kubectl -n dnd-game-prod get pvc
kubectl -n dnd-game-prod get ingress
```

## Environment Structure

```
k8s/
├── base/                    # Base configurations
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── postgres.yaml
│   ├── redis.yaml
│   ├── backend.yaml
│   ├── frontend.yaml
│   └── ingress.yaml
├── overlays/
│   ├── dev/                # Development environment
│   │   ├── kustomization.yaml
│   │   ├── namespace.yaml
│   │   └── deployment-patch.yaml
│   ├── staging/           # Staging environment
│   │   └── kustomization.yaml
│   └── prod/              # Production environment
│       ├── kustomization.yaml
│       ├── namespace.yaml
│       ├── network-policy.yaml
│       └── monitoring.yaml
└── README.md
```

## Configuration

### Environment Variables

Key environment variables configured via ConfigMap:

- `ENV`: Environment (development/staging/production)
- `DB_*`: Database connection settings
- `REDIS_*`: Redis connection settings
- `AI_*`: AI service configuration
- `REACT_APP_*`: Frontend configuration

### Secrets Management

Sensitive data stored in Kubernetes secrets:

- Database passwords
- Redis password
- JWT secret
- API keys

**Best Practices:**
- Use external secret management (Vault, Sealed Secrets)
- Rotate secrets regularly
- Never commit secrets to git

### Resource Limits

Default resource allocations:

| Component | Requests (CPU/Memory) | Limits (CPU/Memory) |
|-----------|---------------------|-------------------|
| Backend   | 250m / 256Mi       | 1000m / 512Mi     |
| Frontend  | 100m / 128Mi       | 500m / 256Mi      |
| PostgreSQL| 250m / 256Mi       | 1000m / 1Gi       |
| Redis     | 100m / 256Mi       | 500m / 512Mi      |

## Scaling

### Horizontal Pod Autoscaling

HPA configured for backend and frontend:

```yaml
# Backend: 3-10 replicas based on CPU/Memory
# Frontend: 2-5 replicas based on CPU/Memory
```

Monitor scaling:
```bash
kubectl -n dnd-game-prod get hpa
kubectl -n dnd-game-prod describe hpa backend-hpa
```

### Manual Scaling

```bash
# Scale backend
kubectl -n dnd-game-prod scale deployment backend --replicas=5

# Scale frontend
kubectl -n dnd-game-prod scale deployment frontend --replicas=3
```

## Monitoring

### Health Checks

All services include health checks:

- **Liveness Probe**: `/health/live` - Restarts unhealthy pods
- **Readiness Probe**: `/health/ready` - Removes from load balancer

Check health:
```bash
# Backend health
curl https://api.dndgame.example.com/health

# Frontend health
curl https://dndgame.example.com/health
```

### Logs

View logs:
```bash
# Backend logs
kubectl -n dnd-game-prod logs -l app.kubernetes.io/component=backend -f

# Frontend logs
kubectl -n dnd-game-prod logs -l app.kubernetes.io/component=frontend -f

# Database logs
kubectl -n dnd-game-prod logs -l app.kubernetes.io/component=database -f
```

### Metrics

If Prometheus is installed:
```bash
# View metrics
kubectl -n dnd-game-prod port-forward svc/backend-service 8080:8080
curl http://localhost:8080/metrics
```

## Backup and Recovery

### Database Backup

Manual backup:
```bash
# Create backup
kubectl -n dnd-game-prod exec -it postgres-0 -- pg_dump -U dndgame dndgame > backup.sql

# Automated backup with CronJob
kubectl apply -f k8s/jobs/backup-cronjob.yaml
```

### Restore Database

```bash
# Copy backup to pod
kubectl -n dnd-game-prod cp backup.sql postgres-0:/tmp/

# Restore
kubectl -n dnd-game-prod exec -it postgres-0 -- psql -U dndgame dndgame < /tmp/backup.sql
```

## Troubleshooting

### Common Issues

1. **Pods not starting**
   ```bash
   kubectl -n dnd-game-prod describe pod <pod-name>
   kubectl -n dnd-game-prod logs <pod-name> --previous
   ```

2. **Database connection issues**
   ```bash
   # Check database service
   kubectl -n dnd-game-prod get svc postgres-service
   kubectl -n dnd-game-prod get endpoints postgres-service
   ```

3. **Ingress not working**
   ```bash
   # Check ingress controller
   kubectl -n ingress-nginx get pods
   kubectl -n dnd-game-prod describe ingress dnd-game-ingress
   ```

4. **Storage issues**
   ```bash
   kubectl -n dnd-game-prod get pvc
   kubectl -n dnd-game-prod describe pvc postgres-pvc
   ```

### Debug Commands

```bash
# Execute commands in pods
kubectl -n dnd-game-prod exec -it backend-xxxxx -- /bin/sh

# Port forward for local testing
kubectl -n dnd-game-prod port-forward svc/backend-service 8080:8080

# Check resource usage
kubectl -n dnd-game-prod top pods
kubectl -n dnd-game-prod top nodes
```

## Security

### Network Policies

Production environment includes strict network policies:
- Default deny all ingress
- Explicit allow rules for required communication
- Egress restrictions

### Pod Security

- Non-root containers
- Read-only root filesystem where possible
- Security contexts enforced

### Secrets Encryption

Enable encryption at rest:
```bash
# Check if encryption is enabled
kubectl get secrets --all-namespaces -o json | jq '.items[].data' | grep -v null | wc -l
```

## CI/CD Integration

### GitHub Actions Example

```yaml
- name: Deploy to Kubernetes
  run: |
    kubectl apply -k k8s/overlays/${{ env.ENVIRONMENT }}
    kubectl -n dnd-game-${{ env.ENVIRONMENT }} rollout status deployment/backend
    kubectl -n dnd-game-${{ env.ENVIRONMENT }} rollout status deployment/frontend
```

### ArgoCD Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: dnd-game
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/dnd-game
    targetRevision: HEAD
    path: k8s/overlays/prod
  destination:
    server: https://kubernetes.default.svc
    namespace: dnd-game-prod
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

## Maintenance

### Updates

1. **Update images**:
   ```bash
   # Update kustomization.yaml with new tags
   vim k8s/overlays/prod/kustomization.yaml
   
   # Apply changes
   kubectl apply -k k8s/overlays/prod
   ```

2. **Rolling updates**:
   ```bash
   # Update backend image
   kubectl -n dnd-game-prod set image deployment/backend backend=dnd-backend:v1.1.0
   
   # Check rollout status
   kubectl -n dnd-game-prod rollout status deployment/backend
   ```

3. **Rollback**:
   ```bash
   # Rollback to previous version
   kubectl -n dnd-game-prod rollout undo deployment/backend
   ```

### Cleanup

Remove all resources:
```bash
# Development
kubectl delete -k k8s/overlays/dev
kubectl delete namespace dnd-game-dev

# Production (BE CAREFUL!)
kubectl delete -k k8s/overlays/prod
kubectl delete namespace dnd-game-prod
```

## Cost Optimization

1. **Use spot instances** for non-critical workloads
2. **Enable cluster autoscaling**
3. **Set resource requests/limits** appropriately
4. **Use PVC resize** instead of creating new volumes
5. **Clean up unused resources** regularly

## Support

For issues or questions:
1. Check pod logs and events
2. Review this documentation
3. Check Kubernetes documentation
4. Open an issue in the repository