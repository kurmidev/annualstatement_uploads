package data

import (
	"database/sql"
	"time"
)

type InvInvestorStatementLog struct {
	ID            int            `db:"id,omitempty"`
	StatementType string         `db:"statement_type"`
	InvestorId    int            `db:"investor_id"`
	FileName      string         `db:"file_name"`
	FilePath      string         `db:"file_path"`
	StartDate     time.Time      `db:"start_date"`
	EndDate       time.Time      `db:"end_date"`
	EmailStatus   string         `db:"email_status,omitempty"`
	SentAt        sql.NullTime   `db:"sent_at,omitempty"`
	RawData       sql.NullString `db:"raw_data"`
	Status        string         `db:"status"`
	IsDeleted     bool           `db:"is_deleted"`
	CreatedBy     int            `db:"created_by"`
	UpdatedBy     int            `db:"updated_by"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}
