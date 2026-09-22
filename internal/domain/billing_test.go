package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestComputeBilling_NoMissingHours(t *testing.T) {
	tiers := []FeeTier{{Position: 1, AmountCents: 500}}
	acc := MemberHourAccount{MissingHours: 0, TargetHours: 10, ConfirmedHours: 10}
	res := ComputeBilling(Member{}, acc, tiers)
	assert.Equal(t, 0, res.TotalCents)
	assert.Empty(t, res.TierBreakdown)
}

func TestComputeBilling_NegativeMissingClamped(t *testing.T) {
	tiers := []FeeTier{{Position: 1, AmountCents: 500}}
	acc := MemberHourAccount{MissingHours: -3}
	res := ComputeBilling(Member{}, acc, tiers)
	assert.Equal(t, 0.0, res.MissingHours)
	assert.Equal(t, 0, res.TotalCents)
}

func TestComputeBilling_NoTiers(t *testing.T) {
	acc := MemberHourAccount{MissingHours: 5}
	res := ComputeBilling(Member{}, acc, nil)
	assert.Equal(t, 0, res.TotalCents)
	assert.Equal(t, 5.0, res.MissingHours)
}

func TestComputeBilling_SingleTierAppliesToAll(t *testing.T) {
	// G-001: one tier, per-missing-hour fee applies to every missing hour.
	tiers := []FeeTier{{Position: 1, AmountCents: 500}}
	acc := MemberHourAccount{MissingHours: 4}
	res := ComputeBilling(Member{}, acc, tiers)
	// Single (last) tier covers all 4 hours at 500 = 2000.
	assert.Equal(t, 2000, res.TotalCents)
	assert.Len(t, res.TierBreakdown, 1)
	assert.Equal(t, 4.0, res.TierBreakdown[0].Hours)
	assert.Equal(t, 2000, res.TierBreakdown[0].SubtotalCents)
}

func TestComputeBilling_MultipleTiersWithOverflow(t *testing.T) {
	// G-002: tiers cover one hour each; the last tier rate applies to the overflow.
	tiers := []FeeTier{
		{Position: 1, AmountCents: 300}, // hour 1
		{Position: 2, AmountCents: 500}, // hours 2..n (last tier, overflow)
	}
	acc := MemberHourAccount{MissingHours: 5}
	res := ComputeBilling(Member{}, acc, tiers)

	// Tier 1: 1h * 300 = 300
	// Tier 2 (last): remaining 4h * 500 = 2000
	assert.Equal(t, 2300, res.TotalCents)
	if assert.Len(t, res.TierBreakdown, 2) {
		assert.Equal(t, 1.0, res.TierBreakdown[0].Hours)
		assert.Equal(t, 300, res.TierBreakdown[0].SubtotalCents)
		assert.Equal(t, 4.0, res.TierBreakdown[1].Hours)
		assert.Equal(t, 2000, res.TierBreakdown[1].SubtotalCents)
	}
}

func TestComputeBilling_FewerMissingThanTiers(t *testing.T) {
	tiers := []FeeTier{
		{Position: 1, AmountCents: 300},
		{Position: 2, AmountCents: 500},
		{Position: 3, AmountCents: 700},
	}
	acc := MemberHourAccount{MissingHours: 1.5}
	res := ComputeBilling(Member{}, acc, tiers)

	// Tier 1: 1.0h * 300 = 300; Tier 2: 0.5h * 500 = 250
	assert.Equal(t, 550, res.TotalCents)
	assert.Len(t, res.TierBreakdown, 2)
}

func TestComputeBilling_FractionalSingleTier(t *testing.T) {
	tiers := []FeeTier{{Position: 1, AmountCents: 250}}
	acc := MemberHourAccount{MissingHours: 2.5}
	res := ComputeBilling(Member{}, acc, tiers)
	assert.Equal(t, int(2.5*250), res.TotalCents)
}

func TestBillingResult_TotalEuro(t *testing.T) {
	b := BillingResult{TotalCents: 1234}
	assert.InDelta(t, 12.34, b.TotalEuro(), 1e-9)
}

func TestComputeBilling_PopulatesMemberAndYear(t *testing.T) {
	memberID := uuid.New()
	yearID := uuid.New()
	m := Member{ID: memberID, FirstName: "A", LastName: "B"}
	acc := MemberHourAccount{ClubYearID: yearID, MissingHours: 2, TargetHours: 10, ConfirmedHours: 8}
	tiers := []FeeTier{{Position: 1, AmountCents: 100}}
	res := ComputeBilling(m, acc, tiers)
	assert.Equal(t, memberID, res.MemberID)
	assert.Equal(t, yearID, res.YearID)
	assert.Equal(t, 10.0, res.TargetHours)
	assert.Equal(t, 8.0, res.ConfirmedHours)
}
