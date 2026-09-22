package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// FeeTier defines the fee charged for a specific missing-hour bracket.
// Position 1 is the cheapest tier; tiers are applied in ascending position order.
// AmountCents is the per-missing-hour fee in euro cents.
type FeeTier struct {
	ID          uuid.UUID `json:"id"`
	ClubYearID  uuid.UUID `json:"clubYearId"`
	Position    int       `json:"position"`   // 1-based ordering
	AmountCents int       `json:"amountCents"` // e.g. 500 = €5.00 per missing hour
}

// BillingResult holds the computed billing for a single member.
type BillingResult struct {
	MemberID       uuid.UUID `json:"memberId"`
	Member         Member    `json:"member"`
	YearID         uuid.UUID `json:"yearId"`
	TargetHours    float64   `json:"targetHours"`
	ConfirmedHours float64   `json:"confirmedHours"`
	MissingHours   float64   `json:"missingHours"`
	TotalCents     int       `json:"totalCents"`
	TierBreakdown  []TierLine `json:"tierBreakdown"`
}

// TierLine shows how many hours fell into each fee tier.
type TierLine struct {
	TierPosition int     `json:"tierPosition"`
	AmountCents  int     `json:"amountCents"`
	Hours        float64 `json:"hours"`
	SubtotalCents int    `json:"subtotalCents"`
}

// TotalEuro returns the total amount as a floating-point euro value.
func (b *BillingResult) TotalEuro() float64 {
	return float64(b.TotalCents) / 100.0
}

// YearBillingReport aggregates billing results for a whole club year.
type YearBillingReport struct {
	ClubYear    ClubYear        `json:"clubYear"`
	Results     []BillingResult `json:"results"`
	TotalCents  int             `json:"totalCents"`
	ComputedAt  string          `json:"computedAt"`
}

// ComputeBilling applies fee tiers to the missing hours and returns a BillingResult.
// Tiers are applied proportionally: the first tier covers the first bracket of hours,
// subsequent tiers cover the remaining hours. In the current simple model, all missing
// hours are charged at the tier rate corresponding to their bracket position.
func ComputeBilling(member Member, account MemberHourAccount, tiers []FeeTier) BillingResult {
	missing := account.MissingHours
	if missing < 0 {
		missing = 0
	}

	result := BillingResult{
		MemberID:       member.ID,
		Member:         member,
		YearID:         account.ClubYearID,
		TargetHours:    account.TargetHours,
		ConfirmedHours: account.ConfirmedHours,
		MissingHours:   missing,
	}

	if missing == 0 || len(tiers) == 0 {
		return result
	}

	// Sort tiers by position (assume already sorted from repo).
	// Each tier covers one "slot" of missing hours.
	// If there are more missing hours than tiers, the last tier rate applies to the remainder.
	remaining := missing
	for i, tier := range tiers {
		if remaining <= 0 {
			break
		}

		var hoursInTier float64
		if i < len(tiers)-1 {
			// Each tier covers 1 hour slot; beyond last tier the last rate applies.
			hoursInTier = 1.0
			if remaining < hoursInTier {
				hoursInTier = remaining
			}
		} else {
			// Last tier covers all remaining hours.
			hoursInTier = remaining
		}

		subtotal := int(hoursInTier * float64(tier.AmountCents))
		result.TierBreakdown = append(result.TierBreakdown, TierLine{
			TierPosition:  tier.Position,
			AmountCents:   tier.AmountCents,
			Hours:         hoursInTier,
			SubtotalCents: subtotal,
		})
		result.TotalCents += subtotal
		remaining -= hoursInTier
	}

	return result
}

// Errors for billing operations.
var (
	ErrBillingYearNotFound = fmt.Errorf("billing year not found")
	ErrNoFeeTiers          = fmt.Errorf("no fee tiers defined for this year")
)
