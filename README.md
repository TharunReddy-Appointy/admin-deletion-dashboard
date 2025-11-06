# Admin Account Deletion Dashboard

A secure web application for Appointy administrators to permanently delete user accounts and their associated data hierarchy (groups, companies, locations) from the production database.

## 🔒 Security Features

- **Google OAuth Authentication**: Only `@appointy.com` email addresses can access
- **JWT-based Sessions**: Secure token-based authentication
- **Soft Delete**: All deletions are soft deletes with `deleted_by` tracking
- **Audit Logging**: Complete audit trail of all deletion operations
- **Two-Step Confirmation**: Requires explicit confirmation before deletion
- **Transaction Safety**: All deletions happen in database transactions

## 📋 Features

1. **Account Lookup**: Search for users by email
2. **Hierarchy Visualization**: View all groups, companies, and locations owned by a user
3. **Selective Deletion**: Choose which groups to delete (with "Select All" option)
4. **Deletion Summary**: See counts before confirming deletion
5. **Audit Trail**: Track who deleted what and when
6. **Responsive UI**: Modern React-based interface

## 🏗️ Architecture

```
User Email → User Profile (user_profile table)
              └── Groups (groups table)
                   └── Companies (company table)
                        └── Locations (location table)
```

### Tech Stack

**Backend:**
- Go 1.21+
- Gin (HTTP framework)
- PostgreSQL (database)
- Google OAuth2
- JWT authentication

**Frontend:**
- React 18 (embedded via CDN)
- Vanilla JavaScript (no build step required)
- Responsive CSS

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database (with access to Appointy production DB)
- Google OAuth Client ID and Secret
- Git

### Installation

1. **Clone the repository:**
   ```bash
   cd /Users/tharunreddy/Office/Backend/admin-deletion-dashboard
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Set up Google OAuth:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select existing
   - Enable Google+ API
   - Create OAuth 2.0 credentials
   - Add authorized redirect URI: `http://localhost:8080/api/auth/callback` (dev) or your production URL
   - Copy Client ID and Client Secret

4. **Configure environment variables:**
   ```bash
   cp .env.example .env
   ```

   Edit `.env` with your configuration:
   ```env
   PORT=8080
   DATABASE_URL=postgres://user:password@localhost:5432/appointy_db?sslmode=disable
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/callback
   JWT_SECRET=your-super-secret-jwt-key
   ENVIRONMENT=development
   ```

5. **Run database migrations:**
   ```bash
   make migrate
   # Or manually:
   psql "$DATABASE_URL" -f migrations/001_create_audit_table.sql
   ```

6. **Build and run:**
   ```bash
   # Development mode
   make run

   # Or build binary
   make build
   ./admin-deletion-dashboard
   ```

7. **Access the dashboard:**
   Open browser to `http://localhost:8080`

## 🐳 Docker Deployment

### Build Docker Image

```bash
make docker-build
# Or:
docker build -t appointy/admin-deletion-dashboard:latest .
```

### Run with Docker

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/db" \
  -e GOOGLE_CLIENT_ID="your-client-id" \
  -e GOOGLE_CLIENT_SECRET="your-secret" \
  -e GOOGLE_REDIRECT_URL="http://localhost:8080/api/auth/callback" \
  -e JWT_SECRET="your-jwt-secret" \
  appointy/admin-deletion-dashboard:latest
```

### Docker Compose (example)

```yaml
version: '3.8'
services:
  admin-dashboard:
    image: appointy/admin-deletion-dashboard:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
      - GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
      - GOOGLE_REDIRECT_URL=${GOOGLE_REDIRECT_URL}
      - JWT_SECRET=${JWT_SECRET}
      - ENVIRONMENT=production
    restart: unless-stopped
```

## 🔧 API Endpoints

### Authentication

- `GET /api/auth/login` - Initiate Google OAuth login
- `GET /api/auth/callback` - OAuth callback handler
- `GET /api/auth/me` - Get current user info (requires auth)
- `POST /api/auth/logout` - Logout user

### Account Management

- `POST /api/account/lookup` - Lookup account by email (requires auth)
  ```json
  {
    "email": "user@example.com"
  }
  ```

- `POST /api/account/delete` - Delete account (requires auth)
  ```json
  {
    "email": "user@example.com",
    "user_id": "usr_123",
    "group_ids": ["grp_1", "grp_2"],
    "reason": "User requested deletion"
  }
  ```

- `GET /api/account/audit-logs` - Get audit logs (requires auth)
  Query params: `limit` (default: 50), `offset` (default: 0)

### Health Check

- `GET /health` - Service health check

## 📊 Database Schema

### Audit Log Table

```sql
CREATE TABLE admin_deletion_audit_log (
    id SERIAL PRIMARY KEY,
    action VARCHAR(50) NOT NULL,
    deleted_by_email VARCHAR(255) NOT NULL,
    target_email VARCHAR(255) NOT NULL,
    target_user_id TEXT NOT NULL,
    group_ids TEXT[],
    reason TEXT,
    deleted_groups INTEGER,
    deleted_companies INTEGER,
    deleted_locations INTEGER,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🔐 Security Considerations

### Production Deployment Checklist

- [ ] Use strong JWT secret (minimum 32 characters)
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS policies
- [ ] Set `ENVIRONMENT=production`
- [ ] Use read-only database credentials where possible
- [ ] Enable database connection pooling
- [ ] Add rate limiting middleware
- [ ] Configure IP whitelisting (Appointy office IPs only)
- [ ] Set up monitoring and alerting
- [ ] Enable database backup before deletions
- [ ] Review and rotate OAuth credentials regularly
- [ ] Enable 2FA for Google Workspace accounts
- [ ] Set session timeout appropriately

### Access Control

Only users with `@appointy.com` email addresses can authenticate. The application validates:
1. Email domain during OAuth callback
2. Email verification status from Google
3. JWT token validity on each protected request

## 📝 Usage Guide

### Deleting an Account

1. **Login**: Click "Sign in with Google" and authenticate with your @appointy.com account

2. **Lookup Account**: Enter the user's email address and click "Lookup Account"

3. **Review Hierarchy**: The dashboard will show:
   - User details (ID, name, email)
   - All groups owned by the user
   - Company and location counts for each group

4. **Select Groups**:
   - Click on individual groups to select them
   - Or click "Select All" to select all groups

5. **Add Reason** (optional): Enter a reason for the deletion in the text area

6. **Review Summary**: Before deletion, you'll see:
   - Number of groups to be deleted
   - Total companies affected
   - Total locations affected

7. **Confirm Deletion**: Click "Delete Selected" and confirm in the modal

8. **Result**: You'll see a success message with deletion statistics

### Viewing Audit Logs

Audit logs track:
- Who performed the deletion (`deleted_by_email`)
- Which account was deleted (`target_email`)
- When the deletion occurred (`created_at`)
- What was deleted (group/company/location counts)
- Why it was deleted (`reason`)

Access logs via: `GET /api/account/audit-logs`

## 🛠️ Development

### Project Structure

```
admin-deletion-dashboard/
├── cmd/
│   └── main/
│       └── main.go              # Application entry point
├── internal/
│   ├── auth/
│   │   └── google_auth.go       # Google OAuth & JWT logic
│   ├── handler/
│   │   ├── auth_handler.go      # Auth HTTP handlers
│   │   └── account_handler.go   # Account HTTP handlers
│   ├── service/
│   │   └── account_service.go   # Business logic & DB operations
│   └── models/
│       └── models.go            # Data models
├── web/
│   └── index.html               # Frontend (React embedded)
├── migrations/
│   └── 001_create_audit_table.sql
├── .env.example
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

### Adding New Features

1. **Backend**: Add logic to `internal/service/`
2. **API**: Add handlers to `internal/handler/`
3. **Frontend**: Modify `web/index.html` React components
4. **Routes**: Update `cmd/main/main.go` router setup

### Running Tests

```bash
make test
```

### Linting & Formatting

```bash
make fmt    # Format code
make lint   # Run linter
```

## 📈 Monitoring

### Health Checks

The application exposes a health endpoint:
```bash
curl http://localhost:8080/health
# Response: {"status":"healthy"}
```

### Logging

The application logs:
- Server startup
- Database connections
- Authentication events
- All deletion operations
- Errors and warnings

### Metrics (Future Enhancement)

Consider adding:
- Prometheus metrics
- Request duration tracking
- Deletion operation counts
- Error rates

## 🚨 Troubleshooting

### Common Issues

**Issue**: "Failed to connect to database"
- **Solution**: Check `DATABASE_URL` is correct and database is accessible

**Issue**: "Only @appointy.com emails are allowed"
- **Solution**: Ensure you're logging in with an Appointy Google account

**Issue**: "Invalid token"
- **Solution**: Logout and login again. JWT token may have expired

**Issue**: "User not found"
- **Solution**: Verify the email address is correct and user exists in database

**Issue**: OAuth callback fails
- **Solution**: Ensure `GOOGLE_REDIRECT_URL` matches the authorized redirect URI in Google Console

## 📄 License

Internal Appointy tool - Confidential and Proprietary

## 👥 Contributors

Appointy Engineering Team

## 📞 Support

For issues or questions, contact:
- Engineering Team: engineering@appointy.com
- Slack: #engineering-support

---

**⚠️ Warning**: This tool performs irreversible database operations. Always verify the account details before confirming deletion.
