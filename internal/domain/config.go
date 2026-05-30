package domain

// BillingMode determines whether billing is computed automatically or manually.
type BillingMode string

const (
	BillingModeAuto   BillingMode = "auto"
	BillingModeManual BillingMode = "manuell"
)

// AppSettings holds the configurable business-logic settings for the application.
type AppSettings struct {
	NameMode             NameMode    `json:"nameMode"`
	KioskSearch          bool        `json:"kioskSearch"`
	ReservationHours     int         `json:"reservationHours"`
	DeregisterDeadlineH  int         `json:"deregisterDeadlineH"`
	BillingMode          BillingMode `json:"billingMode"`
}

// DefaultAppSettings returns the factory-default application settings.
func DefaultAppSettings() AppSettings {
	return AppSettings{
		NameMode:            NameModeAbbrev,
		KioskSearch:         false,
		ReservationHours:    48,
		DeregisterDeadlineH: 24,
		BillingMode:         BillingModeManual,
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
	SettingKeyNameMode            = "nameMode"
	SettingKeyKioskSearch         = "kioskSearch"
	SettingKeyReservationHours    = "reservationHours"
	SettingKeyDeregisterDeadlineH = "deregisterDeadlineH"
	SettingKeyBillingMode         = "billingMode"
)
