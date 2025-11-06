# Deployment Guide - Railway.app

This guide will help you deploy the Admin Account Deletion Dashboard to Railway.app for temporary hosting.

## Prerequisites

- GitHub account
- Railway.app account (free tier available)
- Google OAuth credentials
- Access to Appointy production database

## Step-by-Step Deployment

### 1. Prepare the Code

First, initialize a git repository and push to GitHub:

```bash
cd /Users/tharunreddy/Office/Backend/admin-deletion-dashboard

# Initialize git (if not already done)
git init

# Add all files
git add .

# Commit
git commit -m "Initial commit: Admin deletion dashboard"

# Create GitHub repository and push
# (Create a new repo on GitHub first, then:)
git remote add origin https://github.com/YOUR_USERNAME/admin-deletion-dashboard.git
git branch -M main
git push -u origin main
```

### 2. Set Up Railway

1. **Go to Railway.app**
   - Visit https://railway.app
   - Click "Start a New Project"
   - Sign in with GitHub

2. **Create New Project**
   - Click "Deploy from GitHub repo"
   - Select your `admin-deletion-dashboard` repository
   - Railway will auto-detect the Dockerfile

3. **Add PostgreSQL Database** (if using Railway's database)
   - Click "+ New"
   - Select "Database"
   - Choose "PostgreSQL"
   - Railway will provision a database and provide connection URL

### 3. Configure Environment Variables

In Railway dashboard, go to your service → Variables tab and add:

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=production

# Database (if using existing Appointy DB)
DATABASE_URL=postgres://user:password@your-db-host:5432/appointy_prod?sslmode=require

# OR (if using Railway's PostgreSQL)
# DATABASE_URL will be automatically set by Railway

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=https://your-app.railway.app/api/auth/callback

# JWT Secret
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters
```

**Important:** The `GOOGLE_REDIRECT_URL` will be your Railway app URL. You'll get this after deployment.

### 4. Update Google OAuth Settings

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Navigate to "APIs & Services" → "Credentials"
3. Select your OAuth 2.0 Client
4. Under "Authorized redirect URIs", add:
   ```
   https://your-app-name.railway.app/api/auth/callback
   ```
5. Save changes

### 5. Run Database Migration

You have two options:

**Option A: Connect directly to Railway's PostgreSQL**
```bash
# Get the DATABASE_URL from Railway dashboard
railway login
railway link
railway run psql $DATABASE_URL -f migrations/001_create_audit_table.sql
```

**Option B: Use psql locally**
```bash
# Copy DATABASE_URL from Railway dashboard
psql "YOUR_DATABASE_URL" -f migrations/001_create_audit_table.sql
```

### 6. Deploy

Railway will automatically deploy when you push to GitHub:

```bash
# Make any changes
git add .
git commit -m "Update configuration"
git push origin main

# Railway will auto-deploy
```

Or manually trigger deployment from Railway dashboard.

### 7. Access Your Dashboard

1. Once deployed, Railway will provide a URL like: `https://admin-deletion-dashboard-production.up.railway.app`
2. Update the `GOOGLE_REDIRECT_URL` environment variable with this URL
3. Restart the service in Railway
4. Access your dashboard at the Railway URL

## Alternative: Quick Deploy Button

Add this to your GitHub README for one-click deployment:

```markdown
[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/new/template?template=https://github.com/YOUR_USERNAME/admin-deletion-dashboard)
```

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `PORT` | Server port (Railway sets this automatically) | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@host:5432/db` |
| `GOOGLE_CLIENT_ID` | Google OAuth Client ID | `xxx.apps.googleusercontent.com` |
| `GOOGLE_CLIENT_SECRET` | Google OAuth Client Secret | `GOCSPX-xxx` |
| `GOOGLE_REDIRECT_URL` | OAuth callback URL | `https://yourapp.railway.app/api/auth/callback` |
| `JWT_SECRET` | Secret for signing JWT tokens | Min 32 random characters |
| `ENVIRONMENT` | Runtime environment | `production` |

## Monitoring & Logs

### View Logs
```bash
# Using Railway CLI
railway logs

# Or view in Railway dashboard → Deployments → Logs
```

### Health Check
```bash
curl https://your-app.railway.app/health
```

## Troubleshooting

### Issue: "Build failed"
- Check Railway build logs
- Ensure Dockerfile is correct
- Verify Go version compatibility

### Issue: "Application crashed"
- Check environment variables are set correctly
- Verify DATABASE_URL is accessible from Railway
- Check application logs in Railway dashboard

### Issue: OAuth redirect fails
- Ensure GOOGLE_REDIRECT_URL matches Railway URL
- Verify redirect URI is added in Google Cloud Console
- Check that URL uses HTTPS (Railway provides this automatically)

### Issue: Database connection fails
- If using external DB, ensure Railway's IP is whitelisted
- Check DATABASE_URL format is correct
- Verify SSL mode (`sslmode=require` for production)

## Cost Estimates

Railway pricing (as of 2024):
- **Free Tier**: $5 credit/month (enough for testing)
- **Pro Plan**: $20/month for team use
- **Database**: ~$5-10/month for PostgreSQL
- **Bandwidth**: Usually free tier is sufficient

For temporary/testing: **Free tier should work fine**

## Scaling (Future)

If you need more resources:

1. **Vertical Scaling**
   - Railway dashboard → Settings → Resources
   - Increase memory/CPU

2. **Horizontal Scaling**
   - Update `railway.json` → `numReplicas`
   - Railway will load balance automatically

## Security Checklist

Before going live:

- [ ] Strong JWT_SECRET set (32+ characters)
- [ ] ENVIRONMENT=production
- [ ] DATABASE_URL uses SSL (sslmode=require)
- [ ] Google OAuth restricted to @appointy.com
- [ ] Database backups enabled
- [ ] Railway environment variables are encrypted (done automatically)
- [ ] Test OAuth login flow
- [ ] Test account deletion with test data first

## Rollback

If something goes wrong:

```bash
# Via Railway CLI
railway rollback

# Or in Railway dashboard:
# Deployments → Select previous deployment → Redeploy
```

## Maintenance

### Update Application
```bash
git add .
git commit -m "Your changes"
git push origin main
# Railway auto-deploys
```

### Update Environment Variables
- Railway dashboard → Variables → Update → Restart service

### Database Migrations
```bash
railway run psql $DATABASE_URL -f migrations/002_your_new_migration.sql
```

## Support

- Railway Docs: https://docs.railway.app
- Railway Discord: https://discord.gg/railway
- Railway Status: https://status.railway.app

---

**Next Steps:**
1. Push code to GitHub
2. Connect Railway to GitHub repo
3. Add environment variables
4. Run database migration
5. Test the dashboard
6. Share URL with authorized team members
