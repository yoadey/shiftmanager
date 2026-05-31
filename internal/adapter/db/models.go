package db

import "time"

// MemberModel is the GORM model for the "members" table.
type MemberModel struct {
	ID                  string     `gorm:"column:id;type:text;primaryKey"`
	FirstName           string     `gorm:"column:first_name;type:text"`
	LastName            string     `gorm:"column:last_name;type:text"`
	Email               string     `gorm:"column:email;type:text;index"`
	JoinedAt            time.Time  `gorm:"column:joined_at"`
	LeftAt              *time.Time `gorm:"column:left_at"`
	IsActive            bool       `gorm:"column:is_active;index"`
	IndividualGoalHours *float64   `gorm:"column:individual_goal_hours"`
	OIDCSubject         *string    `gorm:"column:oidc_subject;type:text"`
	Role                string     `gorm:"column:role;type:text"`
	ReminderOptOut      bool       `gorm:"column:reminder_opt_out"`
}

func (MemberModel) TableName() string { return "members" }

// OIDCLinkModel is the GORM model for the "oidc_links" table.
type OIDCLinkModel struct {
	ID       string    `gorm:"column:id;type:text;primaryKey"`
	MemberID string    `gorm:"column:member_id;type:text;index"`
	Provider string    `gorm:"column:provider;type:text"`
	Subject  string    `gorm:"column:subject;type:text"`
	LinkedAt time.Time `gorm:"column:linked_at"`
}

func (OIDCLinkModel) TableName() string { return "oidc_links" }

// EventModel is the GORM model for the "events" table.
type EventModel struct {
	ID          string    `gorm:"column:id;type:text;primaryKey"`
	Name        string    `gorm:"column:name;type:text"`
	Description string    `gorm:"column:description;type:text"`
	Location    string    `gorm:"column:location;type:text"`
	Category    string    `gorm:"column:category;type:text"`
	StartDate   time.Time `gorm:"column:start_date"`
	EndDate     time.Time `gorm:"column:end_date"`
	Status      string    `gorm:"column:status;type:text;index"`
	Visibility  string    `gorm:"column:visibility;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (EventModel) TableName() string { return "events" }

// ShiftModel is the GORM model for the "shifts" table.
type ShiftModel struct {
	ID                    string    `gorm:"column:id;type:text;primaryKey"`
	EventID               string    `gorm:"column:event_id;type:text;index"`
	Name                  string    `gorm:"column:name;type:text"`
	StartAt               time.Time `gorm:"column:start_at;index"`
	EndAt                 time.Time `gorm:"column:end_at"`
	MinHelpers            int       `gorm:"column:min_helpers"`
	MaxHelpers            int       `gorm:"column:max_helpers"`
	RequiredQualification string    `gorm:"column:required_qualification;type:text"`
	ShiftDate             time.Time `gorm:"column:shift_date"`
}

func (ShiftModel) TableName() string { return "shifts" }

// RegistrationModel is the GORM model for the "registrations" table.
type RegistrationModel struct {
	ID                string     `gorm:"column:id;type:text;primaryKey"`
	ShiftID           string     `gorm:"column:shift_id;type:text;index"`
	MemberID          *string    `gorm:"column:member_id;type:text;index"`
	GuestEmail        *string    `gorm:"column:guest_email;type:text"`
	State             string     `gorm:"column:state;type:text"`
	Comment           string     `gorm:"column:comment;type:text"`
	ReservedUntil     *time.Time `gorm:"column:reserved_until"`
	BookedHours       *float64   `gorm:"column:booked_hours"`
	ConfirmationToken *string    `gorm:"column:confirmation_token;type:text"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
}

func (RegistrationModel) TableName() string { return "registrations" }

// HourEntryModel is the GORM model for the "hour_entries" table.
type HourEntryModel struct {
	ID          string    `gorm:"column:id;type:text;primaryKey"`
	MemberID    string    `gorm:"column:member_id;type:text;index"`
	ShiftID     *string   `gorm:"column:shift_id;type:text"`
	ClubYearID  string    `gorm:"column:club_year_id;type:text;index"`
	Hours       float64   `gorm:"column:hours"`
	Type        string    `gorm:"column:type;type:text"`
	Status      string    `gorm:"column:status;type:text"`
	BookedBy    *string   `gorm:"column:booked_by;type:text"`
	Description string    `gorm:"column:description;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (HourEntryModel) TableName() string { return "hour_entries" }

// ClubYearModel is the GORM model for the "club_years" table.
type ClubYearModel struct {
	ID                 string    `gorm:"column:id;type:text;primaryKey"`
	Label              string    `gorm:"column:label;type:text"`
	StartDate          time.Time `gorm:"column:start_date"`
	EndDate            time.Time `gorm:"column:end_date"`
	DefaultTargetHours float64   `gorm:"column:default_target_hours"`
	IsActive           bool      `gorm:"column:is_active"`
}

func (ClubYearModel) TableName() string { return "club_years" }

// HourTargetModel is the GORM model for the "hour_targets" table.
type HourTargetModel struct {
	ID          string  `gorm:"column:id;type:text;primaryKey"`
	MemberID    string  `gorm:"column:member_id;type:text;index"`
	ClubYearID  string  `gorm:"column:club_year_id;type:text;index"`
	TargetHours float64 `gorm:"column:target_hours"`
}

func (HourTargetModel) TableName() string { return "hour_targets" }

// AuditEntryModel is the GORM model for the "audit_log" table.
type AuditEntryModel struct {
	ID        string    `gorm:"column:id;type:text;primaryKey"`
	ActorID   *string   `gorm:"column:actor_id;type:text;index"`
	Action    string    `gorm:"column:action;type:text"`
	Entity    string    `gorm:"column:entity;type:text;index"`
	EntityID  string    `gorm:"column:entity_id;type:text;index"`
	Before    string    `gorm:"column:before_state;type:text"`
	After     string    `gorm:"column:after_state;type:text"`
	ChangedAt time.Time `gorm:"column:changed_at;index"`
}

func (AuditEntryModel) TableName() string { return "audit_log" }

// SettingModel is the GORM model for the "settings" (app_settings) table.
type SettingModel struct {
	Key   string `gorm:"column:key;type:text;primaryKey"`
	Value string `gorm:"column:value;type:text"`
}

func (SettingModel) TableName() string { return "app_settings" }

// BrandingModel is the GORM model for the "branding_config" table (single row, ID=1).
type BrandingModel struct {
	ID           int    `gorm:"column:id;primaryKey"`
	ClubName     string `gorm:"column:club_name;type:text"`
	PrimaryColor string `gorm:"column:primary_color;type:text"`
	AccentColor  string `gorm:"column:accent_color;type:text"`
	LogoURL      string `gorm:"column:logo_url;type:text"`
}

func (BrandingModel) TableName() string { return "branding_config" }

// FeeTierModel is the GORM model for the "fee_tiers" table.
type FeeTierModel struct {
	ID          string `gorm:"column:id;type:text;primaryKey"`
	ClubYearID  string `gorm:"column:club_year_id;type:text;index"`
	Position    int    `gorm:"column:position"`
	AmountCents int    `gorm:"column:amount_cents"`
}

func (FeeTierModel) TableName() string { return "fee_tiers" }

// MemberFeeTierModel is the GORM model for the "member_fee_tiers" table.
type MemberFeeTierModel struct {
	ID          string `gorm:"column:id;type:text;primaryKey"`
	MemberID    string `gorm:"column:member_id;type:text;index"`
	ClubYearID  string `gorm:"column:club_year_id;type:text;index"`
	Position    int    `gorm:"column:position"`
	AmountCents int    `gorm:"column:amount_cents"`
}

func (MemberFeeTierModel) TableName() string { return "member_fee_tiers" }

// EmailTemplateModel is the GORM model for the "email_templates" table.
type EmailTemplateModel struct {
	ID      string `gorm:"column:id;type:text;primaryKey"`
	Name    string `gorm:"column:name;type:text;uniqueIndex"`
	Subject string `gorm:"column:subject;type:text"`
	Body    string `gorm:"column:body;type:text"`
}

func (EmailTemplateModel) TableName() string { return "email_templates" }

// EmailLogModel is the GORM model for the "email_log" table.
type EmailLogModel struct {
	ID        string    `gorm:"column:id;type:text;primaryKey"`
	To        string    `gorm:"column:to_address;type:text"`
	Template  string    `gorm:"column:template;type:text"`
	Subject   string    `gorm:"column:subject;type:text"`
	Body      string    `gorm:"column:body;type:text"`
	Status    string    `gorm:"column:status;type:text"`
	Error     string    `gorm:"column:error;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;index"`
}

func (EmailLogModel) TableName() string { return "email_log" }
