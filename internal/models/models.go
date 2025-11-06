package models

import "time"

// AccountLookupRequest represents the request to look up an account
type AccountLookupRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// GroupInfo represents a group with its hierarchy
type GroupInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CompanyCount  int    `json:"company_count"`
	LocationCount int    `json:"location_count"`
}

// AccountLookupResponse represents the account details found
type AccountLookupResponse struct {
	UserID    string      `json:"user_id"`
	Email     string      `json:"email"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Groups    []GroupInfo `json:"groups"`
}

// DeleteAccountRequest represents the deletion request
type DeleteAccountRequest struct {
	Email      string   `json:"email" binding:"required,email"`
	UserID     string   `json:"user_id" binding:"required"`
	GroupIDs   []string `json:"group_ids" binding:"required,min=1"`
	Reason     string   `json:"reason"`
	DeletedBy  string   `json:"deleted_by"`  // Will be set by backend from JWT
}

// DeleteAccountResponse represents the deletion result
type DeleteAccountResponse struct {
	Success        bool      `json:"success"`
	Message        string    `json:"message"`
	DeletedGroups  int       `json:"deleted_groups"`
	DeletedCompanies int     `json:"deleted_companies"`
	DeletedLocations int     `json:"deleted_locations"`
	DeletedAt      time.Time `json:"deleted_at"`
}

// UserProfile represents minimal user info from database
type UserProfile struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

// Group represents a group entity
type Group struct {
	ID     string
	Name   string
	Parent string
}

// Company represents a company entity
type Company struct {
	ID     string
	Name   string
	Parent string
}

// Location represents a location entity
type Location struct {
	ID     string
	Name   string
	Parent string
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            string    `json:"id"`
	Action        string    `json:"action"`
	DeletedByEmail string   `json:"deleted_by_email"`
	TargetEmail   string    `json:"target_email"`
	TargetUserID  string    `json:"target_user_id"`
	GroupIDs      []string  `json:"group_ids"`
	CompanyIDs    []string  `json:"company_ids"`
	LocationIDs   []string  `json:"location_ids"`
	Reason        string    `json:"reason"`
	IPAddress     string    `json:"ip_address"`
	CreatedAt     time.Time `json:"created_at"`
}
