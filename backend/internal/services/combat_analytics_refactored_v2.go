package services

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

// CombatantStatsTracker handles statistics tracking for a single combatant
type CombatantStatsTracker struct {
	stats *models.CombatantAnalytics
}

// NewCombatantStatsTracker creates a new tracker
func NewCombatantStatsTracker(analyticsID uuid.UUID, combatant *models.Combatant, currentRound int) *CombatantStatsTracker {
	return &CombatantStatsTracker{
		stats: &models.CombatantAnalytics{
			ID:                 uuid.New(),
			CombatAnalyticsID:  analyticsID,
			CombatantID:        combatant.ID,
			CombatantType:      string(combatant.Type),
			CombatantName:      combatant.Name,
			FinalHP:            combatant.HP,
			RoundsSurvived:     currentRound,
			ConditionsSuffered: models.JSONB(`[]`),
			AbilitiesUsed:      models.JSONB(`[]`),
			CreatedAt:          time.Now(),
		},
	}
}

// ActionProcessor handles different types of combat actions
type ActionProcessor interface {
	ProcessAction(action *models.CombatActionLog, tracker *CombatantStatsTracker)
}

// AttackActionProcessor processes attack actions
type AttackActionProcessor struct{}

func (p *AttackActionProcessor) ProcessAction(action *models.CombatActionLog, tracker *CombatantStatsTracker) {
	tracker.stats.AttacksMade++
	
	switch action.Outcome {
	case constants.OutcomeHit:
		tracker.stats.AttacksHit++
		tracker.stats.DamageDealt += action.DamageDealt
	case constants.ActionCritical:
		tracker.stats.AttacksHit++
		tracker.stats.CriticalHits++
		tracker.stats.DamageDealt += action.DamageDealt
	case "miss":
		tracker.stats.AttacksMissed++
	case "critical_miss":
		tracker.stats.AttacksMissed++
		tracker.stats.CriticalMisses++
	}
}

// SpellActionProcessor processes spell and ability actions
type SpellActionProcessor struct{}

func (p *SpellActionProcessor) ProcessAction(action *models.CombatActionLog, tracker *CombatantStatsTracker) {
	tracker.stats.DamageDealt += action.DamageDealt
	tracker.addAbilityUsed(action.ActionType)
}

// HealActionProcessor processes healing actions
type HealActionProcessor struct{}

func (p *HealActionProcessor) ProcessAction(action *models.CombatActionLog, tracker *CombatantStatsTracker) {
	tracker.stats.HealingDone += action.DamageDealt
}

// SaveActionProcessor processes saving throw actions
type SaveActionProcessor struct{}

func (p *SaveActionProcessor) ProcessAction(action *models.CombatActionLog, tracker *CombatantStatsTracker) {
	if action.Outcome == "success" {
		tracker.stats.SavesMade++
	} else {
		tracker.stats.SavesFailed++
	}
}

// CombatantAnalyticsCalculator is the refactored version of calculateCombatantAnalytics
type CombatantAnalyticsCalculator struct {
	actionProcessors map[string]ActionProcessor
}

// NewCombatantAnalyticsCalculator creates a new calculator
func NewCombatantAnalyticsCalculator() *CombatantAnalyticsCalculator {
	return &CombatantAnalyticsCalculator{
		actionProcessors: map[string]ActionProcessor{
			constants.ActionAttack:      &AttackActionProcessor{},
			constants.ActionTypeSpell:   &SpellActionProcessor{},
			constants.ActionTypeAbility: &SpellActionProcessor{},
			"heal":                     &HealActionProcessor{},
			"save":                     &SaveActionProcessor{},
		},
	}
}

// CalculateAnalytics is the refactored main function
func (calc *CombatantAnalyticsCalculator) CalculateAnalytics(
	analyticsID uuid.UUID,
	combat *models.Combat,
	actions []*models.CombatActionLog,
) []*models.CombatantReport {
	// Step 1: Initialize trackers
	trackers := calc.initializeTrackers(analyticsID, combat)
	
	// Step 2: Update defeat times
	calc.updateDefeatTimes(trackers, actions)
	
	// Step 3: Process all actions
	calc.processAllActions(trackers, actions)
	
	// Step 4: Generate reports
	reports := calc.generateReports(trackers)
	
	// Step 5: Sort by performance
	calc.sortReportsByPerformance(reports)
	
	return reports
}

// Step 1: Initialize trackers for all combatants
func (calc *CombatantAnalyticsCalculator) initializeTrackers(
	analyticsID uuid.UUID,
	combat *models.Combat,
) map[string]*CombatantStatsTracker {
	trackers := make(map[string]*CombatantStatsTracker)
	
	for i := range combat.Combatants {
		combatant := &combat.Combatants[i]
		tracker := NewCombatantStatsTracker(analyticsID, combatant, combat.Round)
		trackers[combatant.ID] = tracker
	}
	
	return trackers
}

// Step 2: Update defeat times for defeated combatants
func (calc *CombatantAnalyticsCalculator) updateDefeatTimes(
	trackers map[string]*CombatantStatsTracker,
	actions []*models.CombatActionLog,
) {
	for _, action := range actions {
		if action.TargetID != nil && action.Outcome == constants.OutcomeKillingBlow {
			if tracker, exists := trackers[*action.TargetID]; exists {
				if tracker.stats.FinalHP <= 0 {
					tracker.stats.RoundsSurvived = action.RoundNumber
				}
			}
		}
	}
}

// Step 3: Process all combat actions
func (calc *CombatantAnalyticsCalculator) processAllActions(
	trackers map[string]*CombatantStatsTracker,
	actions []*models.CombatActionLog,
) {
	for _, action := range actions {
		// Process actor actions
		if tracker, exists := trackers[action.ActorID]; exists {
			calc.processActorAction(action, tracker)
		}
		
		// Process target effects
		if action.TargetID != nil {
			if tracker, exists := trackers[*action.TargetID]; exists {
				calc.processTargetEffects(action, tracker)
			}
		}
	}
}

// Process action from actor's perspective
func (calc *CombatantAnalyticsCalculator) processActorAction(
	action *models.CombatActionLog,
	tracker *CombatantStatsTracker,
) {
	if processor, exists := calc.actionProcessors[action.ActionType]; exists {
		processor.ProcessAction(action, tracker)
	}
}

// Process effects on the target
func (calc *CombatantAnalyticsCalculator) processTargetEffects(
	action *models.CombatActionLog,
	tracker *CombatantStatsTracker,
) {
	switch action.ActionType {
	case constants.ActionAttack, constants.ActionTypeSpell:
		tracker.stats.DamageTaken += action.DamageDealt
	case constants.ActionHeal:
		tracker.stats.HealingReceived += action.DamageDealt
	}
	
	// Track conditions applied
	if len(action.ConditionsApplied) > 0 {
		tracker.addConditions(action.ConditionsApplied)
	}
}

// Step 4: Generate performance reports
func (calc *CombatantAnalyticsCalculator) generateReports(
	trackers map[string]*CombatantStatsTracker,
) []*models.CombatantReport {
	reports := make([]*models.CombatantReport, 0, len(trackers))
	
	rater := NewPerformanceRater()
	highlighter := NewHighlightGenerator()
	
	for _, tracker := range trackers {
		report := &models.CombatantReport{
			Analytics:         tracker.stats,
			PerformanceRating: rater.RatePerformance(tracker.stats),
			Highlights:        highlighter.GenerateHighlights(tracker.stats),
		}
		reports = append(reports, report)
	}
	
	return reports
}

// Step 5: Sort reports by damage dealt
func (calc *CombatantAnalyticsCalculator) sortReportsByPerformance(reports []*models.CombatantReport) {
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Analytics.DamageDealt > reports[j].Analytics.DamageDealt
	})
}

// Helper methods for CombatantStatsTracker
func (t *CombatantStatsTracker) addAbilityUsed(ability string) {
	abilities := t.unmarshalAbilities()
	abilities = append(abilities, ability)
	t.stats.AbilitiesUsed = t.marshalAbilities(abilities)
}

func (t *CombatantStatsTracker) addConditions(newConditions models.JSONB) {
	conditions := t.unmarshalConditions()
	
	var toAdd []string
	_ = json.Unmarshal(newConditions, &toAdd)
	conditions = append(conditions, toAdd...)
	
	t.stats.ConditionsSuffered = t.marshalConditions(conditions)
}

func (t *CombatantStatsTracker) unmarshalAbilities() []string {
	var abilities []string
	_ = json.Unmarshal(t.stats.AbilitiesUsed, &abilities)
	return abilities
}

func (t *CombatantStatsTracker) marshalAbilities(abilities []string) models.JSONB {
	data, _ := json.Marshal(abilities)
	return models.JSONB(data)
}

func (t *CombatantStatsTracker) unmarshalConditions() []string {
	var conditions []string
	_ = json.Unmarshal(t.stats.ConditionsSuffered, &conditions)
	return conditions
}

func (t *CombatantStatsTracker) marshalConditions(conditions []string) models.JSONB {
	data, _ := json.Marshal(conditions)
	return models.JSONB(data)
}

// PerformanceRater calculates performance ratings
type PerformanceRater struct {
	weights PerformanceWeights
}

type PerformanceWeights struct {
	HitRateThreshold    float64
	HitRateScore        int
	CriticalHitScore    int
	DamagePerRoundScore int
	SurvivalScore       int
}

func NewPerformanceRater() *PerformanceRater {
	return &PerformanceRater{
		weights: PerformanceWeights{
			HitRateThreshold:    0.75,
			HitRateScore:        3,
			CriticalHitScore:    2,
			DamagePerRoundScore: 1,
			SurvivalScore:       2,
		},
	}
}

func (r *PerformanceRater) RatePerformance(stats *models.CombatantAnalytics) string {
	score := 0
	
	// Calculate hit rate score
	if stats.AttacksMade > 0 {
		hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
		if hitRate > r.weights.HitRateThreshold {
			score += r.weights.HitRateScore
		}
	}
	
	// Add critical hit bonus
	score += stats.CriticalHits * r.weights.CriticalHitScore
	
	// Add damage per round score
	if stats.RoundsSurvived > 0 {
		damagePerRound := stats.DamageDealt / stats.RoundsSurvived
		score += damagePerRound / 10 * r.weights.DamagePerRoundScore
	}
	
	// Add survival bonus
	if stats.FinalHP > 0 {
		score += r.weights.SurvivalScore
	}
	
	return r.scoreToRating(score)
}

func (r *PerformanceRater) scoreToRating(score int) string {
	switch {
	case score >= 10:
		return "Legendary"
	case score >= 7:
		return "Excellent"
	case score >= 4:
		return "Good"
	case score >= 2:
		return "Average"
	default:
		return "Poor"
	}
}

// HighlightGenerator creates performance highlights
type HighlightGenerator struct{}

func NewHighlightGenerator() *HighlightGenerator {
	return &HighlightGenerator{}
}

func (g *HighlightGenerator) GenerateHighlights(stats *models.CombatantAnalytics) []string {
	highlights := []string{}
	
	// Check for perfect accuracy
	if stats.AttacksMade > 5 && stats.AttacksMissed == 0 {
		highlights = append(highlights, "Perfect Accuracy!")
	}
	
	// Check for high critical rate
	if stats.AttacksMade > 0 {
		critRate := float64(stats.CriticalHits) / float64(stats.AttacksMade)
		if critRate > 0.2 {
			highlights = append(highlights, "Critical Hit Master")
		}
	}
	
	// Check for tank performance
	if stats.DamageTaken > 50 && stats.FinalHP > 0 {
		highlights = append(highlights, "Damage Sponge")
	}
	
	// Check for healer performance
	if stats.HealingDone > 30 {
		highlights = append(highlights, "Combat Medic")
	}
	
	// Check for damage dealer
	if stats.DamageDealt > 100 {
		highlights = append(highlights, "Damage Dealer Extraordinaire")
	}
	
	return highlights
}