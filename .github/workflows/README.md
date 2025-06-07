# GitHub Actions Workflows

This directory contains our CI/CD pipeline configurations.

## Workflows

### ci.yml - Main CI/CD Pipeline
Runs on every push and pull request to main/develop branches.

**Jobs:**
- **backend-lint**: Runs golangci-lint to check code quality
- **backend-test**: Runs all backend tests with coverage reporting (currently at 84%!)
- **backend-security**: Runs Gosec and Trivy security scans
- **frontend-lint**: Runs ESLint and TypeScript checks
- **frontend-test**: Runs frontend tests with Jest
- **docker-build**: Builds and scans Docker images
- **integration-test**: Runs integration tests with full stack
- **code-quality**: Runs SonarCloud analysis (if configured)
- **summary**: Provides a summary of all job results

### pr-checks.yml - Pull Request Validations
Runs on pull request creation and updates.

**Checks:**
- PR size analysis (warns if PR is too large)
- Test file detection (reminds to add tests)
- Database migration conflict detection

### release.yml - Release Automation
Triggered when a version tag is pushed (e.g., v1.0.0).

**Features:**
- Builds binaries for Linux, macOS, and Windows
- Creates GitHub release with changelog
- Builds and pushes Docker images (if Docker Hub credentials are configured)

## Configuration

### Required Secrets
- `GITHUB_TOKEN`: Automatically provided by GitHub
- `DOCKER_USERNAME`: Docker Hub username (optional, for releases)
- `DOCKER_PASSWORD`: Docker Hub password (optional, for releases)
- `SONAR_TOKEN`: SonarCloud token (optional, for code quality)

### Environment Variables
- `GO_VERSION`: Go version to use (currently 1.21)
- `NODE_VERSION`: Node.js version to use (currently 18)
- `POSTGRES_VERSION`: PostgreSQL version for tests (currently 14)

## Local Testing

To test workflows locally, you can use [act](https://github.com/nektos/act):

```bash
# Test the CI workflow
act -j backend-test

# Test with specific event
act pull_request -j pr-size
```

## Maintenance

- Update Go/Node versions in the workflow files as needed
- Review and update linting rules in `.golangci.yml`
- Check Dependabot PRs weekly for dependency updates