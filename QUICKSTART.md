# 🚀 Quick Start Guide

Get the Admin Deletion Dashboard running in **5 minutes** on Railway.app

## Prerequisites

✅ GitHub account
✅ Railway.app account (sign up free at https://railway.app)
✅ Google OAuth credentials
✅ Database access (Appointy production DB)

## Deploy in 5 Steps

### Step 1: Push to GitHub (2 min)

```bash
cd admin-deletion-dashboard

# Initialize git
git init
git add .
git commit -m "Initial commit"

# Push to GitHub (create repo first on github.com)
git remote add origin https://github.com/YOUR_USERNAME/admin-deletion-dashboard.git
git push -u origin main
```

### Step 2: Deploy to Railway (1 min)

1. Go to https://railway.app/new
2. Click **"Deploy from GitHub repo"**
3. Select `admin-deletion-dashboard`
4. Railway will auto-detect Dockerfile and start building

### Step 3: Setup Google OAuth (1 min)

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create OAuth 2.0 Client (or use existing)
3. Add redirect URI (replace with your Railway URL):
   ```
   https://admin-deletion-dashboard-production.up.railway.app/api/auth/callback
   ```
4. Copy **Client ID** and **Client Secret**

### Step 4: Add Environment Variables (1 min)

In Railway dashboard → Your Service → Variables tab:

```bash
DATABASE_URL=postgres://user:pass@host:5432/appointy_prod?sslmode=require
GOOGLE_CLIENT_ID=your-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-your-secret
GOOGLE_REDIRECT_URL=https://your-app.railway.app/api/auth/callback
JWT_SECRET=your-random-32-char-secret-key-here
ENVIRONMENT=production
```

**Pro tip:** Generate JWT secret:
```bash
openssl rand -base64 32
```

### Step 5: Run Database Migration (30 sec)

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login and link to your project
railway login
railway link

# Run migration
railway run psql $DATABASE_URL -f migrations/001_create_audit_table.sql
```

## ✅ Done!

Your dashboard is live at: `https://your-app-name.railway.app`

## First Login

1. Open the Railway URL in browser
2. Click **"Sign in with Google"**
3. Login with your **@appointy.com** account
4. Start managing account deletions!

## Test the Dashboard

### Test Account Lookup

1. Enter a test email in the dashboard
2. Click "Lookup Account"
3. Review groups, companies, locations

### Test Deletion (Use test data!)

1. Select groups to delete
2. Add optional reason
3. Click "Delete Selected"
4. Confirm in modal
5. Verify success message

## Monitoring

### Check Logs
```bash
railway logs
```

### Check Health
```bash
curl https://your-app.railway.app/health
```

## Common Issues & Fixes

| Issue | Fix |
|-------|-----|
| Build fails | Check Railway build logs, verify Dockerfile |
| OAuth fails | Verify redirect URL matches in Google Console |
| DB connection fails | Check DATABASE_URL, ensure SSL enabled |
| Only @appointy.com allowed | Working as intended! Use Appointy email |

## Environment Variables Checklist

- [ ] `DATABASE_URL` - PostgreSQL connection string
- [ ] `GOOGLE_CLIENT_ID` - From Google Console
- [ ] `GOOGLE_CLIENT_SECRET` - From Google Console
- [ ] `GOOGLE_REDIRECT_URL` - Your Railway URL + `/api/auth/callback`
- [ ] `JWT_SECRET` - Random 32+ character string
- [ ] `ENVIRONMENT` - Set to `production`

## Next Steps

✅ Share Railway URL with team
✅ Bookmark the dashboard
✅ Test with non-production data first
✅ Review audit logs regularly

## Need Help?

- 📖 Full docs: See [DEPLOYMENT.md](./DEPLOYMENT.md)
- 📖 README: See [README.md](./README.md)
- 🐛 Issues: Check Railway logs
- 💬 Support: Contact #engineering-support on Slack

---

**⚠️ Security Reminder:** This dashboard performs permanent deletions. Always verify before confirming!
