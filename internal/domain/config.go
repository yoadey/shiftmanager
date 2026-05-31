package domain

// BillingMode determines whether billing is computed automatically or manually.
type BillingMode string

const (
	BillingModeAuto   BillingMode = "auto"
	BillingModeManual BillingMode = "manuell"
)

// AppSettings holds the configurable business-logic settings for the application.
type AppSettings struct {
	NameMode                NameMode    `json:"nameMode"`
	KioskSearch             bool        `json:"kioskSearch"`
	ReservationHours        int         `json:"reservationHours"`
	DeregisterDeadlineH     int         `json:"deregisterDeadlineH"`
	BillingMode             BillingMode `json:"billingMode"`
	ReminderHourOfDay       int         `json:"reminderHourOfDay"`       // N-002: hour-of-day (UTC) reminders are sent
	ReminderLeadWeeks       int         `json:"reminderLeadWeeks"`       // N-002: weeks before a shift for the "early" reminder
	BillingWarningLeadWeeks int         `json:"billingWarningLeadWeeks"` // weeks before year end the missing-hours warning is sent
	KioskLocked             bool        `json:"kioskLocked"`             // K-012: when true the public kiosk routes are disabled
}

// DefaultAppSettings returns the factory-default application settings.
func DefaultAppSettings() AppSettings {
	return AppSettings{
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
