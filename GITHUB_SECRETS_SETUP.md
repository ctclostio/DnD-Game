# GitHub Secrets Setup Guide

This guide explains how to set up GitHub secrets for the D&D Game repository to secure sensitive configuration in CI/CD workflows.

## Required Secrets

The following secrets need to be configured in your GitHub repository:

### 1. `TEST_DB_PASSWORD`
- **Purpose**: PostgreSQL password for test database in CI/CD
- **Default fallback**: `postgres` (for backward compatibility)
- **Recommended value**: A strong, randomly generated password
- **Example**: `xK9#mP2$vL5^qR8&hN3@`

### 2. `TEST_JWT_SECRET`
- **Purpose**: JWT signing secret for test environment
- **Default fallback**: `test-secret-key-that-is-at-least-32-characters-long`
- **Recommended value**: A 64-character random string
- **Example**: `a7f3b9d5e2c8a1b6d4f7e9c3b5a8d2e6f1a4c7b9d3e5f8a2c4b7d9e1f3a5c8b2`

## How to Add Secrets

1. **Navigate to Repository Settings**
   - Go to your repository on GitHub
   - Click on "Settings" tab
   - In the left sidebar, click on "Secrets and variables" → "Actions"

2. **Add New Repository Secret**
   - Click "New repository secret"
   - Enter the secret name (e.g., `TEST_DB_PASSWORD`)
   - Enter the secret value
   - Click "Add secret"

3. **Repeat for Each Secret**
   - Add `TEST_DB_PASSWORD`
   - Add `TEST_JWT_SECRET`

## Security Benefits

By using GitHub secrets instead of hardcoded values:

1. **No Exposed Credentials**: Sensitive values are never stored in code
2. **SonarCloud Compliance**: Resolves security vulnerability warnings
3. **Easy Rotation**: Secrets can be changed without code modifications
4. **Environment Isolation**: Different values can be used per environment

## CI/CD Workflow Usage

The secrets are used in the following workflows:
- `.github/workflows/ci.yml`
- `.github/workflows/backend-tests.yml`

Example usage in workflows:
```yaml
env:
  DB_PASSWORD: ${{ secrets.TEST_DB_PASSWORD || 'postgres' }}
  JWT_SECRET: ${{ secrets.TEST_JWT_SECRET || 'test-secret-key-that-is-at-least-32-characters-long' }}
```

The `|| 'fallback'` syntax provides backward compatibility if secrets aren't configured yet.

## Generating Secure Values

### For TEST_DB_PASSWORD:
```bash
# Generate a 16-character password
openssl rand -base64 16 | tr -d "=+/" | cut -c1-16
```

### For TEST_JWT_SECRET:
```bash
# Generate a 64-character secret
openssl rand -hex 32
```

## Verification

After setting up the secrets:

1. Push a commit to trigger CI/CD
2. Check the Actions tab to ensure workflows run successfully
3. Verify that SonarCloud no longer reports hardcoded password vulnerabilities

## Important Notes

- Never commit actual secret values to the repository
- Use different secrets for production environments
- Rotate secrets periodically for better security
- The fallback values are only for backward compatibility and should not be relied upon

## Related Security Improvements

This setup addresses the following SonarCloud security issues:
- **BLOCKER**: "Make sure this PostgreSQL database password gets changed and removed from the code"
- **Rule**: secrets:S6698