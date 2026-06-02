package domain

// BillingMode determines whether billing is computed automatically or manually.
type BillingMode string

const (
	BillingModeAuto   BillingMode = "auto"
	BillingModeManual BillingMode = "manuell"
)

// AppSettings holds the configurable business-logic settings for the application.
type AppSettings struct {
	ClubName                string      `json:"clubName"`
	ClubYear                string      `json:"clubYear"`
	YearGoal                int         `json:"yearGoal"`
	FeeSchedule             []float32   `json:"feeSchedule"`
	NameMode                NameMode    `json:"nameMode"`
	KioskSearch             bool        `json:"kioskSearch"`
	ReservationHours        int         `json:"reservationHours"`
	DeregisterDeadlineH     int         `json:"deregisterDeadlineH"`
	BillingMode             BillingMode `json:"billingMode"`
	ReminderHourOfDay       int         `json:"reminderHourOfDay"`
	ReminderLeadWeeks       int         `json:"reminderLeadWeeks"`
	BillingWarningLeadWeeks int         `json:"billingWarningLeadWeeks"`
	KioskLocked             bool        `json:"kioskLocked"`
}

// DefaultAppSettings returns the factory-default application settings.
func DefaultAppSettings() AppSettings {
	return AppSettings{
		ClubName:                "TSC Schwarz-Gelb Aachen",
		ClubYear:                "2026",
		YearGoal:                20,
		FeeSchedule:             []float32{5, 7, 10, 15},
		NameMode:                NameModeAbbrev,
		KioskSearch:             false,
		ReservationHours:        48,
		DeregisterDeadlineH:     24,
		BillingMode:             BillingModeManual,
		ReminderHourOfDay:       8,
		ReminderLeadWeeks:       1,
		BillingWarningLeadWeeks: 4,
		KioskLocked:             false,
	}
}

// BrandingConfig holds the visual identity settings for the club.
type BrandingConfig struct {
	ClubName     string `json:"clubName"`
	PrimaryColor string `json:"primaryColor"`
	AccentColor  string `json:"accentColor"`
	LogoURL      string `json:"logoUrl"`
}

// DefaultBranding returns the factory default branding.
func DefaultBranding() BrandingConfig {
	return BrandingConfig{
		ClubName:     "TSC Schwarz-Gelb Aachen",
		PrimaryColor: "#000000",
		AccentColor:  "#F4B63F",
		LogoURL:      "",
	}
}

// AppSettingsKeys maps each setting field to its key-value store key.
const (
	SettingKeyClubName                = "clubName"
	SettingKeyClubYear                = "clubYear"
	SettingKeyYearGoal                = "yearGoal"
	SettingKeyFeeSchedule             = "feeSchedule"
	SettingKeyNameMode                = "nameMode"
	SettingKeyKioskSearch             = "kioskSearch"
	SettingKeyReservationHours        = "reservationHours"
	SettingKeyDeregisterDeadlineH     = "deregisterDeadlineH"
	SettingKeyBillingMode             = "billingMode"
	SettingKeyReminderHourOfDay       = "reminderHourOfDay"
	SettingKeyReminderLeadWeeks       = "reminderLeadWeeks"
	SettingKeyBillingWarningLeadWeeks = "billingWarningLeadWeeks"
	SettingKeyKioskLocked             = "kioskLocked"
)
