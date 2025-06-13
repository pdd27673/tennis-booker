# GitHub Workflows

This directory contains GitHub Actions workflows for CI/CD pipelines.

## Workflows

### CI (`ci.yml`)

The CI workflow runs on every push to the `main` branch and on pull requests. It includes:

- **Backend Tests**: Runs Go tests for the backend services with MongoDB and Redis services
- **Scraper Tests**: Runs Python tests for the scraper service with Playwright

### CD (`cd.yml`)

The CD workflow runs on every push to the `main` branch and when a tag is pushed. It includes:

- **Build Backend**: Builds the Go backend services and uploads the binaries as artifacts
- **Build Scraper**: Builds the Python scraper service and uploads the package as an artifact
- **Deploy to OCI**: Deploys the application to Oracle Cloud Infrastructure (only on tag pushes)

## Adding New Workflows

To add a new workflow:

1. Create a new YAML file in the `.github/workflows` directory
2. Define the workflow triggers, jobs, and steps
3. Test the workflow by pushing to a feature branch

## Required Secrets

For the CD workflow to deploy to OCI, the following secrets need to be set in the repository:

- `OCI_USER_OCID`: The OCID of the user
- `OCI_FINGERPRINT`: The fingerprint of the API key
- `OCI_PRIVATE_KEY`: The private key for OCI authentication
- `OCI_TENANCY_OCID`: The OCID of the tenancy
- `OCI_REGION`: The OCI region to deploy to 