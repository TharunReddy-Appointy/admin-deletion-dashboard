-- Migration: Create audit log table for account deletions
-- Created: 2025-11-06

-- Create the audit log table
CREATE TABLE IF NOT EXISTS admin_deletion_audit_log (
    id SERIAL PRIMARY KEY,
    action VARCHAR(50) NOT NULL DEFAULT 'ACCOUNT_DELETION',
    deleted_by_email VARCHAR(255) NOT NULL,
    target_email VARCHAR(255) NOT NULL,
    target_user_id TEXT NOT NULL,
    group_ids TEXT[] DEFAULT '{}',
    reason TEXT DEFAULT '',
    deleted_groups INTEGER DEFAULT 0,
    deleted_companies INTEGER DEFAULT 0,
    deleted_locations INTEGER DEFAULT 0,
    ip_address VARCHAR(45) DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_audit_deleted_by ON admin_deletion_audit_log(deleted_by_email);
CREATE INDEX IF NOT EXISTS idx_audit_target_email ON admin_deletion_audit_log(target_email);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON admin_deletion_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_target_user_id ON admin_deletion_audit_log(target_user_id);

-- Add comment to table
COMMENT ON TABLE admin_deletion_audit_log IS 'Audit log for account deletions performed through admin dashboard';
COMMENT ON COLUMN admin_deletion_audit_log.deleted_by_email IS 'Email of the admin who performed the deletion';
COMMENT ON COLUMN admin_deletion_audit_log.target_email IS 'Email of the user account that was deleted';
COMMENT ON COLUMN admin_deletion_audit_log.target_user_id IS 'User ID of the deleted account';
COMMENT ON COLUMN admin_deletion_audit_log.group_ids IS 'Array of group IDs that were deleted';
COMMENT ON COLUMN admin_deletion_audit_log.reason IS 'Optional reason provided for the deletion';
COMMENT ON COLUMN admin_deletion_audit_log.deleted_groups IS 'Number of groups deleted';
COMMENT ON COLUMN admin_deletion_audit_log.deleted_companies IS 'Number of companies deleted';
COMMENT ON COLUMN admin_deletion_audit_log.deleted_locations IS 'Number of locations deleted';

-- Grant permissions (adjust based on your database user setup)
-- GRANT SELECT, INSERT ON admin_deletion_audit_log TO your_app_user;
-- GRANT USAGE, SELECT ON SEQUENCE admin_deletion_audit_log_id_seq TO your_app_user;
