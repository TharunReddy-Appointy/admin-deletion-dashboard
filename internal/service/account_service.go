package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.appointy.com/admin-deletion-dashboard/internal/models"
	"github.com/lib/pq"
)

// AccountService handles account operations
type AccountService struct {
	db *sql.DB
}

// NewAccountService creates a new account service
func NewAccountService(db *sql.DB) *AccountService {
	return &AccountService{
		db: db,
	}
}

// LookupAccount finds a user and their owned groups/companies/locations
func (s *AccountService) LookupAccount(ctx context.Context, email string) (*models.AccountLookupResponse, error) {
	// Step 1: Find user profile
	user, err := s.getUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Step 2: Find groups owned by this user
	groups, err := s.getGroupsByOwner(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find groups: %w", err)
	}

	// Step 3: For each group, count companies and locations
	groupInfos := make([]models.GroupInfo, 0, len(groups))
	for _, group := range groups {
		companyCount, locationCount, err := s.getHierarchyCounts(ctx, group.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get hierarchy counts for group %s: %w", group.ID, err)
		}

		groupInfos = append(groupInfos, models.GroupInfo{
			ID:            group.ID,
			Name:          group.Name,
			CompanyCount:  companyCount,
			LocationCount: locationCount,
		})
	}

	return &models.AccountLookupResponse{
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Groups:    groupInfos,
	}, nil
}

// DeleteAccount performs soft delete on user and selected groups hierarchy
func (s *AccountService) DeleteAccount(ctx context.Context, req *models.DeleteAccountRequest) (*models.DeleteAccountResponse, error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	deletedGroups := 0
	deletedCompanies := 0
	deletedLocations := 0

	// For each selected group
	for _, groupID := range req.GroupIDs {
		// Get all companies under this group
		companies, err := s.getCompaniesByParent(ctx, groupID)
		if err != nil {
			return nil, fmt.Errorf("failed to get companies for group %s: %w", groupID, err)
		}

		// For each company, get and delete locations
		for _, company := range companies {
			locations, err := s.getLocationsByParent(ctx, company.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get locations for company %s: %w", company.ID, err)
			}

			// Soft delete locations
			for _, location := range locations {
				if err := s.softDeleteLocation(ctx, tx, location.ID, req.DeletedBy, now); err != nil {
					return nil, fmt.Errorf("failed to delete location %s: %w", location.ID, err)
				}
				deletedLocations++
			}

			// Soft delete company
			if err := s.softDeleteCompany(ctx, tx, company.ID, req.DeletedBy, now); err != nil {
				return nil, fmt.Errorf("failed to delete company %s: %w", company.ID, err)
			}
			deletedCompanies++
		}

		// Soft delete group
		if err := s.softDeleteGroup(ctx, tx, groupID, req.DeletedBy, now); err != nil {
			return nil, fmt.Errorf("failed to delete group %s: %w", groupID, err)
		}
		deletedGroups++
	}

	// Soft delete user profile
	if err := s.softDeleteUser(ctx, tx, req.UserID, req.DeletedBy, now); err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	// Create audit log
	if err := s.createAuditLog(ctx, tx, req, deletedGroups, deletedCompanies, deletedLocations, now); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &models.DeleteAccountResponse{
		Success:          true,
		Message:          "Account and selected hierarchy deleted successfully",
		DeletedGroups:    deletedGroups,
		DeletedCompanies: deletedCompanies,
		DeletedLocations: deletedLocations,
		DeletedAt:        now,
	}, nil
}

// getUserByEmail retrieves user profile by email
func (s *AccountService) getUserByEmail(ctx context.Context, email string) (*models.UserProfile, error) {
	query := `
		SELECT id, email, first_name, last_name
		FROM saastack_user_v1.user_profile
		WHERE LOWER(email) = LOWER($1) AND (is_deleted = false OR is_deleted IS NULL)
		LIMIT 1
	`

	var user models.UserProfile
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found with email: %s", email)
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// getGroupsByOwner retrieves all groups owned by a user
func (s *AccountService) getGroupsByOwner(ctx context.Context, userID string) ([]models.Group, error) {
	query := `
		SELECT id, name, parent
		FROM saastack_group_v1.groups
		WHERE created_by = $1 AND (is_deleted = false OR is_deleted IS NULL)
		ORDER BY created_on DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]models.Group, 0)
	for rows.Next() {
		var group models.Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Parent); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}

	return groups, rows.Err()
}

// getCompaniesByParent retrieves all companies under a group
func (s *AccountService) getCompaniesByParent(ctx context.Context, parentID string) ([]models.Company, error) {
	query := `
		SELECT id, name, parent
		FROM saastack_company_v1.company
		WHERE parent = $1 AND (is_deleted = false OR is_deleted IS NULL)
	`

	rows, err := s.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	companies := make([]models.Company, 0)
	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.Name, &company.Parent); err != nil {
			return nil, err
		}
		companies = append(companies, company)
	}

	return companies, rows.Err()
}

// getLocationsByParent retrieves all locations under a company
func (s *AccountService) getLocationsByParent(ctx context.Context, parentID string) ([]models.Location, error) {
	query := `
		SELECT id, name, parent
		FROM saastack_location_v1.location
		WHERE parent = $1 AND (is_deleted = false OR is_deleted IS NULL)
	`

	rows, err := s.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := make([]models.Location, 0)
	for rows.Next() {
		var location models.Location
		if err := rows.Scan(&location.ID, &location.Name, &location.Parent); err != nil {
			return nil, err
		}
		locations = append(locations, location)
	}

	return locations, rows.Err()
}

// getHierarchyCounts counts companies and locations for a group
func (s *AccountService) getHierarchyCounts(ctx context.Context, groupID string) (int, int, error) {
	// Count companies
	companyQuery := `
		SELECT COUNT(*)
		FROM saastack_company_v1.company
		WHERE parent = $1 AND (is_deleted = false OR is_deleted IS NULL)
	`
	var companyCount int
	if err := s.db.QueryRowContext(ctx, companyQuery, groupID).Scan(&companyCount); err != nil {
		return 0, 0, err
	}

	// Count locations (for all companies under this group)
	locationQuery := `
		SELECT COUNT(l.*)
		FROM saastack_location_v1.location l
		INNER JOIN saastack_company_v1.company c ON l.parent = c.id
		WHERE c.parent = $1
			AND (l.is_deleted = false OR l.is_deleted IS NULL)
			AND (c.is_deleted = false OR c.is_deleted IS NULL)
	`
	var locationCount int
	if err := s.db.QueryRowContext(ctx, locationQuery, groupID).Scan(&locationCount); err != nil {
		return 0, 0, err
	}

	return companyCount, locationCount, nil
}

// softDeleteLocation marks a location as deleted
func (s *AccountService) softDeleteLocation(ctx context.Context, tx *sql.Tx, locationID, deletedBy string, deletedOn time.Time) error {
	query := `
		UPDATE saastack_location_v1.location
		SET is_deleted = true, deleted_by = $1, deleted_on = $2
		WHERE id = $3
	`
	_, err := tx.ExecContext(ctx, query, deletedBy, deletedOn, locationID)
	return err
}

// softDeleteCompany marks a company as deleted
func (s *AccountService) softDeleteCompany(ctx context.Context, tx *sql.Tx, companyID, deletedBy string, deletedOn time.Time) error {
	query := `
		UPDATE saastack_company_v1.company
		SET is_deleted = true, deleted_by = $1, deleted_on = $2
		WHERE id = $3
	`
	_, err := tx.ExecContext(ctx, query, deletedBy, deletedOn, companyID)
	return err
}

// softDeleteGroup marks a group as deleted
func (s *AccountService) softDeleteGroup(ctx context.Context, tx *sql.Tx, groupID, deletedBy string, deletedOn time.Time) error {
	query := `
		UPDATE saastack_group_v1.groups
		SET is_deleted = true, deleted_by = $1, deleted_on = $2
		WHERE id = $3
	`
	_, err := tx.ExecContext(ctx, query, deletedBy, deletedOn, groupID)
	return err
}

// softDeleteUser marks a user profile as deleted
func (s *AccountService) softDeleteUser(ctx context.Context, tx *sql.Tx, userID, deletedBy string, deletedOn time.Time) error {
	query := `
		UPDATE saastack_user_v1.user_profile
		SET is_deleted = true, deleted_by = $1, deleted_on = $2
		WHERE id = $3
	`
	_, err := tx.ExecContext(ctx, query, deletedBy, deletedOn, userID)
	return err
}

// createAuditLog creates an audit log entry
func (s *AccountService) createAuditLog(ctx context.Context, tx *sql.Tx, req *models.DeleteAccountRequest, deletedGroups, deletedCompanies, deletedLocations int, timestamp time.Time) error {
	query := `
		INSERT INTO admin_deletion_audit_log
		(action, deleted_by_email, target_email, target_user_id, group_ids, reason, deleted_groups, deleted_companies, deleted_locations, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := tx.ExecContext(ctx, query,
		"ACCOUNT_DELETION",
		req.DeletedBy,
		req.Email,
		req.UserID,
		pq.Array(req.GroupIDs),
		req.Reason,
		deletedGroups,
		deletedCompanies,
		deletedLocations,
		timestamp,
	)

	return err
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AccountService) GetAuditLogs(ctx context.Context, limit int, offset int) ([]models.AuditLog, error) {
	query := `
		SELECT action, deleted_by_email, target_email, target_user_id, group_ids, reason, created_at
		FROM admin_deletion_audit_log
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]models.AuditLog, 0)
	for rows.Next() {
		var log models.AuditLog
		var groupIDs pq.StringArray

		if err := rows.Scan(
			&log.Action,
			&log.DeletedByEmail,
			&log.TargetEmail,
			&log.TargetUserID,
			&groupIDs,
			&log.Reason,
			&log.CreatedAt,
		); err != nil {
			return nil, err
		}

		log.GroupIDs = []string(groupIDs)
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
