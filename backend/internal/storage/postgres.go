package storage

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// DB wraps sqlx.DB with TerraVault-specific query methods.
type DB struct {
	*sqlx.DB
}

// NewPostgres creates a new PostgreSQL connection pool.
func NewPostgres(dsn string) (*DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &DB{db}, nil
}

// RunMigrations applies SQL migration files from the given directory in order.
func (db *DB) RunMigrations(dir string) error {
	// Ensure migrations table exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	// Read applied migrations
	applied := map[string]bool{}
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("reading applied migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	// List migration files
	var files []string
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".sql") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking migrations dir: %w", err)
	}
	sort.Strings(files)

	for _, f := range files {
		version := filepath.Base(f)
		if applied[version] {
			continue
		}
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", f, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("applying migration %s: %w", version, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", version, err)
		}
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// Project queries
// ────────────────────────────────────────────────────────────────────────────

type Project struct {
	ID                 int64     `db:"id"`
	OnChainPubkey      string    `db:"on_chain_pubkey"`
	DeveloperWallet    string    `db:"developer_wallet"`
	State              string    `db:"state"`
	ProjectType        string    `db:"project_type"`
	MetadataURI        string    `db:"metadata_uri"`
	FundraiseTarget    int64     `db:"fundraise_target"`
	FundraiseHardCap   int64     `db:"fundraise_hard_cap"`
	FundraiseDeadline  time.Time `db:"fundraise_deadline"`
	EscrowBalance      int64     `db:"escrow_balance"`
	TotalRaised        int64     `db:"total_raised"`
	TokenPrice         int64     `db:"token_price"`
	TotalTokens        int64     `db:"total_tokens"`
	TokensSold         int64     `db:"tokens_sold"`
	MilestoneCount     int       `db:"milestone_count"`
	MilestonesCompleted int      `db:"milestones_completed"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func (db *DB) ListProjects(limit, offset int) ([]Project, error) {
	var projects []Project
	err := db.Select(&projects,
		`SELECT * FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	return projects, err
}

func (db *DB) GetProjectByPubkey(pubkey string) (*Project, error) {
	var p Project
	err := db.Get(&p, `SELECT * FROM projects WHERE on_chain_pubkey = $1`, pubkey)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpsertProject(p *Project) error {
	_, err := db.NamedExec(`
		INSERT INTO projects (
			on_chain_pubkey, developer_wallet, state, project_type, metadata_uri,
			fundraise_target, fundraise_hard_cap, fundraise_deadline, escrow_balance,
			total_raised, token_price, total_tokens, tokens_sold,
			milestone_count, milestones_completed, updated_at
		) VALUES (
			:on_chain_pubkey, :developer_wallet, :state, :project_type, :metadata_uri,
			:fundraise_target, :fundraise_hard_cap, :fundraise_deadline, :escrow_balance,
			:total_raised, :token_price, :total_tokens, :tokens_sold,
			:milestone_count, :milestones_completed, NOW()
		)
		ON CONFLICT (on_chain_pubkey) DO UPDATE SET
			state = EXCLUDED.state,
			escrow_balance = EXCLUDED.escrow_balance,
			total_raised = EXCLUDED.total_raised,
			tokens_sold = EXCLUDED.tokens_sold,
			milestones_completed = EXCLUDED.milestones_completed,
			updated_at = NOW()
	`, p)
	return err
}

// ────────────────────────────────────────────────────────────────────────────
// Milestone queries
// ────────────────────────────────────────────────────────────────────────────

type Milestone struct {
	ID              int64      `db:"id"`
	ProjectID       int64      `db:"project_id"`
	MilestoneIndex  int        `db:"milestone_index"`
	Description     string     `db:"description"`
	ReleaseBPS      int        `db:"release_bps"`
	Status          string     `db:"status"`
	ProofURI        string     `db:"proof_uri"`
	ProofHash       string     `db:"proof_hash"`
	SubmittedAt     *time.Time `db:"submitted_at"`
	ApprovedAt      *time.Time `db:"approved_at"`
	ReleasedAmount  int64      `db:"released_amount"`
	DisputeDeadline *time.Time `db:"dispute_deadline"`
}

func (db *DB) GetMilestonesByProject(projectID int64) ([]Milestone, error) {
	var ms []Milestone
	err := db.Select(&ms,
		`SELECT * FROM milestones WHERE project_id = $1 ORDER BY milestone_index`,
		projectID,
	)
	return ms, err
}

func (db *DB) UpsertMilestone(m *Milestone) error {
	_, err := db.NamedExec(`
		INSERT INTO milestones (
			project_id, milestone_index, description, release_bps, status,
			proof_uri, proof_hash, submitted_at, approved_at, released_amount, dispute_deadline
		) VALUES (
			:project_id, :milestone_index, :description, :release_bps, :status,
			:proof_uri, :proof_hash, :submitted_at, :approved_at, :released_amount, :dispute_deadline
		)
		ON CONFLICT (project_id, milestone_index) DO UPDATE SET
			status = EXCLUDED.status,
			proof_uri = EXCLUDED.proof_uri,
			proof_hash = EXCLUDED.proof_hash,
			submitted_at = EXCLUDED.submitted_at,
			approved_at = EXCLUDED.approved_at,
			released_amount = EXCLUDED.released_amount,
			dispute_deadline = EXCLUDED.dispute_deadline
	`, m)
	return err
}

// ────────────────────────────────────────────────────────────────────────────
// KYC queries
// ────────────────────────────────────────────────────────────────────────────

type KYCRecord struct {
	ID                int64      `db:"id"`
	WalletAddress     string     `db:"wallet_address"`
	KYCStatus         string     `db:"kyc_status"`
	Country           string     `db:"country"`
	VerificationLevel int        `db:"verification_level"`
	ProviderSessionID string     `db:"provider_session_id"`
	VerifiedAt        *time.Time `db:"verified_at"`
	CreatedAt         time.Time  `db:"created_at"`
}

func (db *DB) GetKYCByWallet(wallet string) (*KYCRecord, error) {
	var r KYCRecord
	err := db.Get(&r, `SELECT * FROM kyc_records WHERE wallet_address = $1`, wallet)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) UpsertKYC(r *KYCRecord) error {
	_, err := db.NamedExec(`
		INSERT INTO kyc_records (wallet_address, kyc_status, country, verification_level, provider_session_id, verified_at)
		VALUES (:wallet_address, :kyc_status, :country, :verification_level, :provider_session_id, :verified_at)
		ON CONFLICT (wallet_address) DO UPDATE SET
			kyc_status = EXCLUDED.kyc_status,
			country = EXCLUDED.country,
			verification_level = EXCLUDED.verification_level,
			provider_session_id = EXCLUDED.provider_session_id,
			verified_at = EXCLUDED.verified_at
	`, r)
	return err
}

// ────────────────────────────────────────────────────────────────────────────
// Distribution queries
// ────────────────────────────────────────────────────────────────────────────

type Distribution struct {
	ID                  int64     `db:"id"`
	ProjectID           int64     `db:"project_id"`
	Round               int32     `db:"round"`
	TotalUSDC           int64     `db:"total_usdc"`
	TotalTokens         int64     `db:"total_tokens"`
	USDCPerTokenScaled  string    `db:"usdc_per_token_scaled"`
	Source              string    `db:"source"`
	ClaimDeadline       time.Time `db:"claim_deadline"`
	CreatedAt           time.Time `db:"created_at"`
}

func (db *DB) GetDistributionsByProject(projectID int64) ([]Distribution, error) {
	var ds []Distribution
	err := db.Select(&ds,
		`SELECT * FROM distributions WHERE project_id = $1 ORDER BY round`,
		projectID,
	)
	return ds, err
}

func (db *DB) InsertDistribution(d *Distribution) error {
	_, err := db.NamedExec(`
		INSERT INTO distributions (project_id, round, total_usdc, total_tokens, usdc_per_token_scaled, source, claim_deadline)
		VALUES (:project_id, :round, :total_usdc, :total_tokens, :usdc_per_token_scaled, :source, :claim_deadline)
		ON CONFLICT (project_id, round) DO NOTHING
	`, d)
	return err
}

// ────────────────────────────────────────────────────────────────────────────
// Dispute queries
// ────────────────────────────────────────────────────────────────────────────

type Dispute struct {
	ID            int64      `db:"id"`
	ProjectPubkey string     `db:"project_pubkey"`
	RaisedBy      string     `db:"raised_by"`
	ReasonHash    string     `db:"reason_hash"`
	EvidenceHash  string     `db:"evidence_hash"`
	Status        string     `db:"status"`   // open | resolved
	Decision      string     `db:"decision"` // PayInvestors | RefundAndExtend | ForceClose | ""
	Deadline      time.Time  `db:"deadline"`
	ResolvedAt    *time.Time `db:"resolved_at"`
	CreatedAt     time.Time  `db:"created_at"`
}

func (db *DB) GetDisputeByProject(projectPubkey string) (*Dispute, error) {
	var d Dispute
	err := db.Get(&d,
		`SELECT * FROM disputes WHERE project_pubkey = $1 AND status = 'open' LIMIT 1`,
		projectPubkey,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (db *DB) UpsertDispute(d *Dispute) error {
	_, err := db.NamedExec(`
		INSERT INTO disputes (project_pubkey, raised_by, reason_hash, evidence_hash, status, decision, deadline)
		VALUES (:project_pubkey, :raised_by, :reason_hash, :evidence_hash, :status, :decision, :deadline)
		ON CONFLICT (project_pubkey) DO UPDATE SET
			evidence_hash = EXCLUDED.evidence_hash,
			status        = EXCLUDED.status,
			decision      = EXCLUDED.decision,
			resolved_at   = EXCLUDED.resolved_at
	`, d)
	return err
}

// ────────────────────────────────────────────────────────────────────────────
// Blacklist queries
// ────────────────────────────────────────────────────────────────────────────

type BlacklistEntry struct {
	ID          int64     `db:"id"`
	DeveloperPubkey string `db:"developer_pubkey"`
	ReasonHash  string    `db:"reason_hash"`
	FlaggedAt   time.Time `db:"flagged_at"`
}

func (db *DB) GetBlacklist() ([]BlacklistEntry, error) {
	var entries []BlacklistEntry
	err := db.Select(&entries, `SELECT * FROM blacklist ORDER BY flagged_at DESC`)
	return entries, err
}

func (db *DB) GetBlacklistEntry(developerPubkey string) (*BlacklistEntry, error) {
	var e BlacklistEntry
	err := db.Get(&e, `SELECT * FROM blacklist WHERE developer_pubkey = $1`, developerPubkey)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (db *DB) InsertBlacklistEntry(e *BlacklistEntry) error {
	_, err := db.NamedExec(`
		INSERT INTO blacklist (developer_pubkey, reason_hash, flagged_at)
		VALUES (:developer_pubkey, :reason_hash, :flagged_at)
		ON CONFLICT (developer_pubkey) DO NOTHING
	`, e)
	return err
}
