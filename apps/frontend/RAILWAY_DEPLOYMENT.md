# Deploying the Frontend to Railway

This guide provides instructions for deploying the Tennis Booker frontend application to Railway.

## Prerequisites

- A Railway account (https://railway.app/)
- Railway CLI installed (optional, for local development)
- Git repository connected to Railway

## Deployment Steps

### 1. Connect Your Repository

1. Log in to your Railway account
2. Click "New Project"
3. Select "Deploy from GitHub repo"
4. Choose the repository containing the Tennis Booker application
5. Select the branch you want to deploy

### 2. Configure the Project

1. In the Railway dashboard, navigate to your new project
2. Add the following environment variables:

```
PORT=80
API_URL=https://your-backend-api-url.railway.app
NODE_ENV=production
CLERK_PUBLISHABLE_KEY=your_clerk_publishable_key
APP_VERSION=1.0.0
DEPLOYMENT_ENV=production
```

3. Update the `API_URL` to point to your backend service URL

### 3. Deploy the Application

Railway will automatically detect the `railway.json` configuration file and use the `Dockerfile.railway` for building the application.

The deployment process:
1. Builds the application using the multi-stage Dockerfile
2. Sets up Nginx to serve the static files
3. Configures runtime environment variable injection
4. Sets up health checks for monitoring

### 4. Verify the Deployment

1. Once deployed, Railway will provide a URL for your application
2. Visit the URL to verify that the frontend is working correctly
3. Check the Railway logs for any errors or issues

### 5. Custom Domain (Optional)

1. In the Railway dashboard, go to Settings > Domains
2. Add your custom domain
3. Configure the DNS settings as instructed by Railway
4. Wait for the SSL certificate to be provisioned

## Environment Variables

| Variable | Description |
|----------|-------------|
| PORT | The port the server will listen on (default: 80) |
| API_URL | URL of the backend API service |
| NODE_ENV | Environment mode (production, development) |
| CLERK_PUBLISHABLE_KEY | Public key for Clerk authentication |
| APP_VERSION | Application version number |
| DEPLOYMENT_ENV | Deployment environment name |

## Troubleshooting

### Health Check Failures

If the health check is failing:
1. Check the logs in the Railway dashboard
2. Verify that the `/health` endpoint is responding correctly
3. Make sure the Nginx configuration is properly set up

### Environment Variables Not Working

If environment variables are not being injected correctly:
1. Check that the `env-config.js` file exists in the public directory
2. Verify that the environment variables are correctly set in the Railway dashboard
3. Check the logs for any errors during the startup script execution

## Monitoring

Railway provides built-in monitoring for your application:
1. Health checks status
2. CPU and memory usage
3. Request logs
4. Application logs

Visit the Railway dashboard to access these monitoring features.
