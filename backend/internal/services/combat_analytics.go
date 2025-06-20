package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
)

type CombatAnalyticsService struct {
	analyticsRepo database.CombatAnalyticsRepository
	combatService *CombatService
}

func NewCombatAnalyticsService(
	analyticsRepo database.CombatAnalyticsRepository,
	combatService *CombatService,
) *CombatAnalyticsService {
	return &CombatAnalyticsService{
		analyticsRepo: analyticsRepo,
		combatService: combatService,
	}
}

// TrackCombatAction logs a combat action for analytics
func (cas *CombatAnalyticsService) TrackCombatAction(_ context.Context, action *models.CombatActionLog) error {
	return cas.analyticsRepo.CreateCombatAction(action)
}

// FinalizeCombatAnalytics generates the final combat report when combat ends
func (cas *CombatAnalyticsService) FinalizeCombatAnalytics(
	_ context.Context,
	combat *models.Combat,
	sessionID uuid.UUID,
) (*models.CombatAnalyticsReport, error) {
	// Get all combat actions
	// Convert Combat.ID string to uuid.UUID
	combatUUID, err := uuid.Parse(combat.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse combat ID: %w", err)
	}

	actions, err := cas.analyticsRepo.GetCombatActions(combatUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get combat actions: %w", err)
	}

	// Calculate combat analytics
	analytics := cas.calculateCombatAnalytics(combat, sessionID, actions, combatUUID)

	// Save main analytics
	if err := cas.analyticsRepo.CreateCombatAnalytics(analytics); err != nil {
		return nil, fmt.Errorf("failed to save combat analytics: %w", err)
	}

	// Calculate individual combatant analytics
	combatantReports := cas.calculateCombatantAnalytics(analytics.ID, combat, actions)

	// Save combatant analytics
	for _, report := range combatantReports {
		if err := cas.analyticsRepo.CreateCombatantAnalytics(report.Analytics); err != nil {
			return nil, fmt.Errorf("failed to save combatant analytics: %w", err)
		}
	}

	// Generate tactical analysis
	tacticalAnalysis := cas.analyzeTactics(combat, actions, combatantReports)

	// Generate recommendations
	recommendations := cas.generateRecommendations(combat, combatantReports, tacticalAnalysis)

	// Update analytics with AI-generated summary
	summary := cas.generateCombatSummary(analytics, combatantReports, tacticalAnalysis)
	updates := map[string]interface{}{
		"combat_summary":  models.JSONB(summary),
		"tactical_rating": calculateOverallScore(tacticalAnalysis),
	}
	_ = cas.analyticsRepo.UpdateCombatAnalytics(analytics.ID, updates)

	return &models.CombatAnalyticsReport{
		Analytics:        analytics,
		CombatantReports: combatantReports,
		TacticalAnalysis: tacticalAnalysis,
		Recommendations:  recommendations,
	}, nil
}

func (cas *CombatAnalyticsService) calculateCombatAnalytics(
	combat *models.Combat,
	sessionID uuid.UUID,
	actions []*models.CombatActionLog,
	combatUUID uuid.UUID,
) *models.CombatAnalytics {
	stats := cas.calculateActionStats(actions)
	mvpInfo := cas.determineMVP(stats.damageByActor, combat.Combatants)
	killingBlowsJSON, _ := json.Marshal(stats.killingBlows)

	return &models.CombatAnalytics{
		ID:               uuid.New(),
		CombatID:         combatUUID,
		GameSessionID:    sessionID,
		CombatDuration:   combat.Round,
		TotalDamageDealt: stats.totalDamage,
		TotalHealingDone: stats.totalHealing,
		KillingBlows:     models.JSONB(killingBlowsJSON),
		MVPID:            mvpInfo.id,
		MVPType:          mvpInfo.combatantType,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

type actionStats struct {
	totalDamage   int
	totalHealing  int
	damageByActor map[string]int
	killingBlows  []map[string]interface{}
}

type mvpInfo struct {
	id            string
	combatantType string
}

func (cas *CombatAnalyticsService) calculateActionStats(actions []*models.CombatActionLog) actionStats {
	stats := actionStats{
		damageByActor: make(map[string]int),
		killingBlows:  []map[string]interface{}{},
	}

	for _, action := range actions {
		cas.processActionForStats(action, &stats)
	}

	return stats
}

func (cas *CombatAnalyticsService) processActionForStats(action *models.CombatActionLog, stats *actionStats) {
	if action.ActionType == constants.ActionAttack || action.ActionType == constants.ActionTypeSpell {
		stats.totalDamage += action.DamageDealt
		stats.damageByActor[action.ActorID] += action.DamageDealt
	}

	if action.ActionType == constants.ActionHeal {
		stats.totalHealing += action.DamageDealt // Healing stored as positive damage
	}

	if action.Outcome == constants.OutcomeKillingBlow {
		stats.killingBlows = append(stats.killingBlows, map[string]interface{}{
			"dealer_id": action.ActorID,
			"target_id": action.TargetID,
			"damage":    action.DamageDealt,
		})
	}
}

func (cas *CombatAnalyticsService) determineMVP(damageByActor map[string]int, combatants []models.Combatant) mvpInfo {
	var mvp mvpInfo
	maxDamage := 0

	for actorID, damage := range damageByActor {
		if damage > maxDamage {
			maxDamage = damage
			mvp.id = actorID
			mvp.combatantType = cas.getCombatantType(actorID, combatants)
		}
	}

	return mvp
}

func (cas *CombatAnalyticsService) getCombatantType(actorID string, combatants []models.Combatant) string {
	for i := range combatants {
		if combatants[i].ID == actorID {
			return string(combatants[i].Type)
		}
	}
	return ""
}

func (cas *CombatAnalyticsService) calculateCombatantAnalytics(
	analyticsID uuid.UUID,
	combat *models.Combat,
	actions []*models.CombatActionLog,
) []*models.CombatantReport {
	// Create and initialize stats for all combatants
	combatantStats := cas.initializeCombatantStats(analyticsID, combat, actions)

	// Process all actions to update stats
	for _, action := range actions {
		cas.processActionForActor(combatantStats, action)
		cas.processActionForTarget(combatantStats, action)
		cas.processSaveAction(combatantStats, action)
	}

	return cas.generateCombatantReports(combatantStats)
}

func (cas *CombatAnalyticsService) initializeCombatantStats(
	analyticsID uuid.UUID,
	combat *models.Combat,
	actions []*models.CombatActionLog,
) map[string]*models.CombatantAnalytics {
	combatantStats := make(map[string]*models.CombatantAnalytics)

	for i := range combat.Combatants {
		combatant := &combat.Combatants[i]
		stats := &models.CombatantAnalytics{
			ID:                 uuid.New(),
			CombatAnalyticsID:  analyticsID,
			CombatantID:        combatant.ID,
			CombatantType:      string(combatant.Type),
			CombatantName:      combatant.Name,
			FinalHP:            combatant.HP,
			RoundsSurvived:     combat.Round,
			ConditionsSuffered: models.JSONB(`[]`),
			AbilitiesUsed:      models.JSONB(`[]`),
			CreatedAt:          time.Now(),
		}

		if combatant.HP <= 0 {
			stats.RoundsSurvived = cas.findDefeatedRound(combatant.ID, actions)
		}

		combatantStats[combatant.ID] = stats
	}

	return combatantStats
}

func (cas *CombatAnalyticsService) findDefeatedRound(combatantID string, actions []*models.CombatActionLog) int {
	for _, action := range actions {
		if action.TargetID != nil && *action.TargetID == combatantID && action.Outcome == constants.OutcomeKillingBlow {
			return action.RoundNumber
		}
	}
	return 0
}

func (cas *CombatAnalyticsService) processActionForActor(
	combatantStats map[string]*models.CombatantAnalytics,
	action *models.CombatActionLog,
) {
	stats, ok := combatantStats[action.ActorID]
	if !ok {
		return
	}

	switch action.ActionType {
	case constants.ActionAttack:
		cas.processAttackAction(stats, action)
	case constants.ActionTypeSpell, constants.ActionTypeAbility:
		cas.processSpellOrAbilityAction(stats, action)
	case "heal":
		stats.HealingDone += action.DamageDealt
	}
}

func (cas *CombatAnalyticsService) processAttackAction(
	stats *models.CombatantAnalytics,
	action *models.CombatActionLog,
) {
	stats.AttacksMade++
	cas.updateAttackOutcome(stats, action.Outcome)
	stats.DamageDealt += action.DamageDealt
}

func (cas *CombatAnalyticsService) updateAttackOutcome(stats *models.CombatantAnalytics, outcome string) {
	switch outcome {
	case constants.OutcomeHit:
		stats.AttacksHit++
	case constants.ActionCritical:
		stats.AttacksHit++
		stats.CriticalHits++
	case "miss":
		stats.AttacksMissed++
	case "critical_miss":
		stats.AttacksMissed++
		stats.CriticalMisses++
	}
}

func (cas *CombatAnalyticsService) processSpellOrAbilityAction(
	stats *models.CombatantAnalytics,
	action *models.CombatActionLog,
) {
	stats.DamageDealt += action.DamageDealt
	
	// Track ability usage
	abilities := []string{}
	_ = json.Unmarshal(stats.AbilitiesUsed, &abilities)
	abilities = append(abilities, action.ActionType)
	abilitiesJSON, _ := json.Marshal(abilities)
	stats.AbilitiesUsed = models.JSONB(abilitiesJSON)
}

func (cas *CombatAnalyticsService) processActionForTarget(
	combatantStats map[string]*models.CombatantAnalytics,
	action *models.CombatActionLog,
) {
	if action.TargetID == nil {
		return
	}

	stats, ok := combatantStats[*action.TargetID]
	if !ok {
		return
	}

	switch action.ActionType {
	case constants.ActionAttack, constants.ActionTypeSpell:
		stats.DamageTaken += action.DamageDealt
	case constants.ActionHeal:
		stats.HealingReceived += action.DamageDealt
	}

	cas.trackConditions(stats, action)
}

func (cas *CombatAnalyticsService) trackConditions(
	stats *models.CombatantAnalytics,
	action *models.CombatActionLog,
) {
	if len(action.ConditionsApplied) == 0 {
		return
	}

	conditions := []string{}
	_ = json.Unmarshal(stats.ConditionsSuffered, &conditions)

	var newConditions []string
	_ = json.Unmarshal(action.ConditionsApplied, &newConditions)
	conditions = append(conditions, newConditions...)

	conditionsJSON, _ := json.Marshal(conditions)
	stats.ConditionsSuffered = models.JSONB(conditionsJSON)
}

func (cas *CombatAnalyticsService) processSaveAction(
	combatantStats map[string]*models.CombatantAnalytics,
	action *models.CombatActionLog,
) {
	if action.ActionType != "save" {
		return
	}

	stats, ok := combatantStats[action.ActorID]
	if !ok {
		return
	}

	if action.Outcome == "success" {
		stats.SavesMade++
	} else {
		stats.SavesFailed++
	}
}

func (cas *CombatAnalyticsService) generateCombatantReports(
	combatantStats map[string]*models.CombatantAnalytics,
) []*models.CombatantReport {
	reports := []*models.CombatantReport{}

	for _, stats := range combatantStats {
		report := &models.CombatantReport{
			Analytics:         stats,
			PerformanceRating: cas.ratePerformance(stats),
			Highlights:        cas.generateHighlights(stats),
		}
		reports = append(reports, report)
	}

	// Sort by damage dealt
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Analytics.DamageDealt > reports[j].Analytics.DamageDealt
	})

	return reports
}

func (cas *CombatAnalyticsService) ratePerformance(stats *models.CombatantAnalytics) string {
	score := 0
	
	score += cas.scoreDamageEfficiency(stats)
	score += cas.scoreSurvival(stats)
	score += cas.scoreImpact(stats)
	score += cas.scoreSpecialContributions(stats)
	
	return cas.getPerformanceRating(score)
}

func (cas *CombatAnalyticsService) scoreDamageEfficiency(stats *models.CombatantAnalytics) int {
	if stats.AttacksMade == 0 {
		return 0
	}
	
	hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
	switch {
	case hitRate > 0.75:
		return 3
	case hitRate > 0.5:
		return 2
	case hitRate > 0.25:
		return 1
	default:
		return 0
	}
}

func (cas *CombatAnalyticsService) scoreSurvival(stats *models.CombatantAnalytics) int {
	if stats.FinalHP <= 0 {
		return 0
	}
	
	score := 2
	if stats.DamageTaken == 0 {
		score += 2 // No damage taken
	}
	return score
}

func (cas *CombatAnalyticsService) scoreImpact(stats *models.CombatantAnalytics) int {
	if stats.DamageDealt > stats.DamageTaken*2 {
		return 2
	}
	return 0
}

func (cas *CombatAnalyticsService) scoreSpecialContributions(stats *models.CombatantAnalytics) int {
	score := 0
	if stats.CriticalHits > 0 {
		score++
	}
	if stats.HealingDone > 0 {
		score += 2
	}
	return score
}

func (cas *CombatAnalyticsService) getPerformanceRating(score int) string {
	switch {
	case score >= 8:
		return "excellent"
	case score >= 5:
		return "good"
	case score >= 3:
		return "fair"
	default:
		return constants.EconomicPoor
	}
}

func (cas *CombatAnalyticsService) generateHighlights(stats *models.CombatantAnalytics) []string {
	highlights := []string{}
	
	cas.addAccuracyHighlight(&highlights, stats)
	cas.addCriticalHitHighlight(&highlights, stats)
	cas.addDamageHighlight(&highlights, stats)
	cas.addHealingHighlight(&highlights, stats)
	cas.addDefenseHighlight(&highlights, stats)
	cas.addSavingThrowHighlight(&highlights, stats)
	
	return highlights
}

func (cas *CombatAnalyticsService) addAccuracyHighlight(highlights *[]string, stats *models.CombatantAnalytics) {
	if stats.AttacksMade == 0 {
		return
	}
	
	hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
	if hitRate > 0.75 {
		*highlights = append(*highlights, fmt.Sprintf("Exceptional accuracy: %.0f%% hit rate", hitRate*100))
	}
}

func (cas *CombatAnalyticsService) addCriticalHitHighlight(highlights *[]string, stats *models.CombatantAnalytics) {
	if stats.CriticalHits > 1 {
		*highlights = append(*highlights, fmt.Sprintf("Scored %d critical hits", stats.CriticalHits))
	}
}

func (cas *CombatAnalyticsService) addDamageHighlight(highlights *[]string, stats *models.CombatantAnalytics) {
	if stats.DamageDealt > 50 {
		*highlights = append(*highlights, fmt.Sprintf("Dealt %d total damage", stats.DamageDealt))
	}
}

func (cas *CombatAnalyticsService) addHealingHighlight(highlights *[]string, stats *models.CombatantAnalytics) {
	if stats.HealingDone > 30 {
		*highlights = append(*highlights, fmt.Sprintf("Healed %d HP to allies", stats.HealingDone))
	}
}

func (cas *CombatAnalyticsService) addDefenseHighlight(highlights *[]string, stats *models.CombatantAnalytics) {
	if stats.DamageTaken == 0 && stats.RoundsSurvived > 3 {
		*highlights = append(*highlights, "Avoided all damage")
	}
}

func (cas *CombatAnalyticsService) addSavingThrowHighlight(highlights *[]string, stats *models.CombatantAnalytics) {
	if stats.SavesMade > stats.SavesFailed && stats.SavesMade > 2 {
		*highlights = append(*highlights, "Strong saving throws")
	}
}

func (cas *CombatAnalyticsService) analyzeTactics(
	combat *models.Combat,
	actions []*models.CombatActionLog,
	reports []*models.CombatantReport,
) *models.TacticalAnalysis {
	analysis := &models.TacticalAnalysis{
		PositioningScore:     cas.analyzePositioning(actions),
		ResourceManagement:   cas.analyzeResourceUse(actions),
		TargetPrioritization: cas.analyzeTargeting(actions, combat),
		TeamworkScore:        cas.analyzeTeamwork(actions, reports),
		MissedOpportunities:  cas.findMissedOpportunities(actions, combat),
	}

	return analysis
}

func (cas *CombatAnalyticsService) analyzePositioning(actions []*models.CombatActionLog) int {
	// Analyze movement and positioning choices
	score := 5 // Base score

	coverUses := 0
	advantageousPositions := 0

	for _, action := range actions {
		if action.PositionData != nil {
			var posData map[string]interface{}
			_ = json.Unmarshal(action.PositionData, &posData)

			if cover, ok := posData["used_cover"].(bool); ok && cover {
				coverUses++
			}

			if advantage, ok := posData["high_ground"].(bool); ok && advantage {
				advantageousPositions++
			}
		}
	}

	if coverUses > 5 {
		score += 2
	} else if coverUses > 2 {
		score++
	}

	if advantageousPositions > 3 {
		score += 2
	} else if advantageousPositions > 1 {
		score++
	}

	return min(10, score)
}

func (cas *CombatAnalyticsService) analyzeResourceUse(actions []*models.CombatActionLog) int {
	// Analyze spell slot and ability usage efficiency
	metrics := cas.calculateResourceMetrics(actions)
	return cas.scoreResourceUsage(metrics)
}

type resourceMetrics struct {
	highLevelSpellsOnMinions int
	wastedHealing           int
	efficientResourceUse    int
}

func (cas *CombatAnalyticsService) calculateResourceMetrics(actions []*models.CombatActionLog) resourceMetrics {
	metrics := resourceMetrics{}

	for _, action := range actions {
		cas.analyzeSpellUsage(action, &metrics)
		cas.analyzeHealingUsage(action, &metrics)
	}

	return metrics
}

func (cas *CombatAnalyticsService) analyzeSpellUsage(action *models.CombatActionLog, metrics *resourceMetrics) {
	if action.ResourcesUsed == nil {
		return
	}

	spellLevel := cas.extractSpellLevel(action.ResourcesUsed)
	if spellLevel == 0 {
		return
	}

	if spellLevel >= 3 && action.DamageDealt < 20 {
		metrics.highLevelSpellsOnMinions++
	} else if action.DamageDealt > int(spellLevel)*10 {
		metrics.efficientResourceUse++
	}
}

func (cas *CombatAnalyticsService) extractSpellLevel(resourceData []byte) int {
	var resources map[string]interface{}
	if err := json.Unmarshal(resourceData, &resources); err != nil {
		return 0
	}

	if spellLevel, ok := resources["spell_level"].(float64); ok {
		return int(spellLevel)
	}
	return 0
}

func (cas *CombatAnalyticsService) analyzeHealingUsage(action *models.CombatActionLog, metrics *resourceMetrics) {
	if action.ActionType != constants.ActionHeal {
		return
	}

	// Check if healing was wasted (overhealing)
	if action.DamageDealt > 20 && action.Outcome == "overheal" {
		metrics.wastedHealing++
	}
}

func (cas *CombatAnalyticsService) scoreResourceUsage(metrics resourceMetrics) int {
	score := 5

	if metrics.highLevelSpellsOnMinions > 2 {
		score -= 2
	}

	if metrics.wastedHealing > 3 {
		score--
	}

	if metrics.efficientResourceUse > 5 {
		score += 2
	}

	return max(1, min(10, score))
}

func (cas *CombatAnalyticsService) analyzeTargeting(actions []*models.CombatActionLog, combat *models.Combat) int {
	score := 5
	targetingData := cas.analyzeTargetPriorities(actions, combat)
	
	if targetingData.dangerousEnemiesEliminated > 0 {
		score += min(3, targetingData.dangerousEnemiesEliminated)
	}

	return min(10, score)
}

type targetingAnalysis struct {
	targetPriority              map[string]int
	dangerousEnemiesEliminated int
}

func (cas *CombatAnalyticsService) analyzeTargetPriorities(actions []*models.CombatActionLog, combat *models.Combat) targetingAnalysis {
	analysis := targetingAnalysis{
		targetPriority: make(map[string]int),
	}

	for _, action := range actions {
		if cas.isTargetedAttack(action) {
			analysis.targetPriority[*action.TargetID]++
			
			if cas.isHighPriorityElimination(action, combat) {
				analysis.dangerousEnemiesEliminated++
			}
		}
	}

	return analysis
}

func (cas *CombatAnalyticsService) isTargetedAttack(action *models.CombatActionLog) bool {
	return action.TargetID != nil && action.ActionType == constants.ActionAttack
}

func (cas *CombatAnalyticsService) isHighPriorityElimination(action *models.CombatActionLog, combat *models.Combat) bool {
	if action.Outcome != constants.OutcomeKillingBlow {
		return false
	}

	for i := range combat.Combatants {
		if combat.Combatants[i].ID == *action.TargetID {
			// Simple heuristic: early elimination of NPCs indicates good target prioritization
			return combat.Combatants[i].Type == models.CombatantTypeNPC && action.RoundNumber < 5
		}
	}
	
	return false
}

func (cas *CombatAnalyticsService) analyzeTeamwork(actions []*models.CombatActionLog, reports []*models.CombatantReport) int {
	// Analyze coordination and teamwork
	metrics := cas.calculateTeamworkMetrics(actions, reports)
	return cas.scoreTeamwork(metrics)
}

type teamworkMetrics struct {
	comboAttacks       int
	coordinatedHealing int
	setupActions       int
}

func (cas *CombatAnalyticsService) calculateTeamworkMetrics(actions []*models.CombatActionLog, reports []*models.CombatantReport) teamworkMetrics {
	metrics := teamworkMetrics{}

	for i, action := range actions {
		if cas.isComboAttack(i, action, actions) {
			metrics.comboAttacks++
		}

		if cas.isCoordinatedHealing(action, reports) {
			metrics.coordinatedHealing++
		}

		if cas.isSetupAction(action) {
			metrics.setupActions++
		}
	}

	return metrics
}

func (cas *CombatAnalyticsService) isComboAttack(index int, action *models.CombatActionLog, actions []*models.CombatActionLog) bool {
	if action.ActionType != constants.ActionAttack || index == 0 {
		return false
	}

	prevAction := actions[index-1]
	return prevAction.RoundNumber == action.RoundNumber &&
		prevAction.TargetID != nil && action.TargetID != nil &&
		*prevAction.TargetID == *action.TargetID
}

func (cas *CombatAnalyticsService) isCoordinatedHealing(action *models.CombatActionLog, reports []*models.CombatantReport) bool {
	if action.ActionType != constants.ActionHeal || action.TargetID == nil {
		return false
	}

	for _, report := range reports {
		if report.Analytics.CombatantID == *action.TargetID {
			return report.Analytics.DamageTaken > report.Analytics.FinalHP
		}
	}
	return false
}

func (cas *CombatAnalyticsService) isSetupAction(action *models.CombatActionLog) bool {
	if action.ActionType != constants.ActionTypeSpell && action.ActionType != constants.ActionTypeAbility {
		return false
	}

	var conditions []string
	_ = json.Unmarshal(action.ConditionsApplied, &conditions)
	return len(conditions) > 0
}

func (cas *CombatAnalyticsService) scoreTeamwork(metrics teamworkMetrics) int {
	score := 5

	if metrics.comboAttacks > 5 {
		score += 2
	}

	if metrics.coordinatedHealing > 3 {
		score++
	}

	if metrics.setupActions > 4 {
		score += 2
	}

	return min(10, score)
}

func (cas *CombatAnalyticsService) findMissedOpportunities(actions []*models.CombatActionLog, combat *models.Combat) []string {
	opportunities := []string{}

	aoeOpportunities := cas.countMissedAoEOpportunities(actions)
	if aoeOpportunities > 3 {
		opportunities = append(opportunities, "Multiple opportunities for area-of-effect spells were missed")
	}

	// Note: healingDelays calculation was not implemented in original code
	// Keeping the check for consistency
	healingDelays := 0
	if healingDelays > 2 {
		opportunities = append(opportunities, "Healing was delayed, resulting in preventable unconsciousness")
	}

	if cas.hasUnusedHighLevelResources(actions, combat.Round) {
		opportunities = append(opportunities, "High-level spell slots remained unused")
	}

	return opportunities
}

func (cas *CombatAnalyticsService) countMissedAoEOpportunities(actions []*models.CombatActionLog) int {
	aoeOpportunities := 0

	for _, action := range actions {
		if cas.shouldHaveUsedAoE(action, actions) {
			aoeOpportunities++
		}
	}

	return aoeOpportunities
}

func (cas *CombatAnalyticsService) shouldHaveUsedAoE(action *models.CombatActionLog, allActions []*models.CombatActionLog) bool {
	if action.ActionType != constants.ActionAttack || action.TargetID == nil {
		return false
	}

	enemiesInRound := cas.countEnemiesInRound(action.RoundNumber, allActions)
	return enemiesInRound >= 3
}

func (cas *CombatAnalyticsService) countEnemiesInRound(roundNumber int, actions []*models.CombatActionLog) int {
	count := 0
	for _, action := range actions {
		if action.RoundNumber == roundNumber && action.ActorType == "npc" {
			count++
		}
	}
	return count
}

func (cas *CombatAnalyticsService) hasUnusedHighLevelResources(actions []*models.CombatActionLog, finalRound int) bool {
	lastRoundActions := cas.getActionsForRound(actions, finalRound)

	for _, action := range lastRoundActions {
		if cas.hasHighLevelSpellSlotsRemaining(action) {
			return true
		}
	}

	return false
}

func (cas *CombatAnalyticsService) getActionsForRound(actions []*models.CombatActionLog, round int) []*models.CombatActionLog {
	var roundActions []*models.CombatActionLog
	for _, action := range actions {
		if action.RoundNumber == round {
			roundActions = append(roundActions, action)
		}
	}
	return roundActions
}

func (cas *CombatAnalyticsService) hasHighLevelSpellSlotsRemaining(action *models.CombatActionLog) bool {
	if action.ResourcesUsed == nil {
		return false
	}

	var resources map[string]interface{}
	_ = json.Unmarshal(action.ResourcesUsed, &resources)
	
	slots, ok := resources["spell_slots_remaining"].(map[string]interface{})
	if !ok {
		return false
	}

	for level, remaining := range slots {
		if level >= "3" && remaining.(float64) > 0 {
			return true
		}
	}

	return false
}

func (cas *CombatAnalyticsService) generateRecommendations(
	combat *models.Combat,
	reports []*models.CombatantReport,
	analysis *models.TacticalAnalysis,
) []string {
	recommendations := []string{}
	
	cas.addTacticalRecommendations(&recommendations, analysis)
	cas.addPerformanceRecommendations(&recommendations, reports)
	cas.addDurationRecommendations(&recommendations, combat.Round)
	
	return recommendations
}

func (cas *CombatAnalyticsService) addTacticalRecommendations(recommendations *[]string, analysis *models.TacticalAnalysis) {
	tacticalChecks := []struct {
		score   int
		message string
	}{
		{analysis.PositioningScore, "Focus on using cover and terrain advantages more effectively"},
		{analysis.ResourceManagement, "Improve resource management - save high-level spells for tougher enemies"},
		{analysis.TargetPrioritization, "Prioritize dangerous enemies like spellcasters and high-damage dealers"},
		{analysis.TeamworkScore, "Coordinate attacks and support actions for better synergy"},
	}
	
	for _, check := range tacticalChecks {
		if check.score < 5 {
			*recommendations = append(*recommendations, check.message)
		}
	}
}

func (cas *CombatAnalyticsService) addPerformanceRecommendations(recommendations *[]string, reports []*models.CombatantReport) {
	poorPerformers := cas.countPoorPerformers(reports)
	
	if poorPerformers > len(reports)/3 {
		*recommendations = append(*recommendations, "Consider adjusting difficulty or providing tactical guidance to struggling players")
	}
}

func (cas *CombatAnalyticsService) countPoorPerformers(reports []*models.CombatantReport) int {
	count := 0
	for _, report := range reports {
		if report.PerformanceRating == constants.EconomicPoor {
			count++
		}
	}
	return count
}

func (cas *CombatAnalyticsService) addDurationRecommendations(recommendations *[]string, rounds int) {
	switch {
	case rounds > 10:
		*recommendations = append(*recommendations, "Combat lasted very long - consider more aggressive tactics to speed up encounters")
	case rounds < 3:
		*recommendations = append(*recommendations, "Combat ended very quickly - consider adding environmental challenges or reinforcements")
	}
}

func (cas *CombatAnalyticsService) generateCombatSummary(
	analytics *models.CombatAnalytics,
	reports []*models.CombatantReport,
	analysis *models.TacticalAnalysis,
) []byte {
	summary := map[string]interface{}{
		"overview": fmt.Sprintf("Combat lasted %d rounds with %d total damage dealt and %d HP healed.",
			analytics.CombatDuration, analytics.TotalDamageDealt, analytics.TotalHealingDone),
		"mvp": map[string]interface{}{
			"id":     analytics.MVPID,
			"type":   analytics.MVPType,
			"reason": "Highest damage dealer",
		},
		"key_moments": cas.extractKeyMoments(reports),
		"tactical_summary": map[string]interface{}{
			"positioning":     analysis.PositioningScore,
			"resource_use":    analysis.ResourceManagement,
			"target_priority": analysis.TargetPrioritization,
			"teamwork":        analysis.TeamworkScore,
		},
		"outcome_factors": cas.determineOutcomeFactors(analytics, reports),
	}

	summaryJSON, _ := json.Marshal(summary)
	return summaryJSON
}

func (cas *CombatAnalyticsService) extractKeyMoments(reports []*models.CombatantReport) []string {
	moments := []string{}

	for _, report := range reports {
		if report.Analytics.CriticalHits > 2 {
			moments = append(moments, fmt.Sprintf("%s landed %d critical hits",
				report.Analytics.CombatantName, report.Analytics.CriticalHits))
		}

		if report.Analytics.HealingDone > 50 {
			moments = append(moments, fmt.Sprintf("%s provided crucial healing (%d HP)",
				report.Analytics.CombatantName, report.Analytics.HealingDone))
		}

		if report.Analytics.DamageTaken == 0 && report.Analytics.AttacksMade > 5 {
			moments = append(moments, fmt.Sprintf("%s fought flawlessly without taking damage",
				report.Analytics.CombatantName))
		}
	}

	return moments
}

func (cas *CombatAnalyticsService) determineOutcomeFactors(
	analytics *models.CombatAnalytics,
	reports []*models.CombatantReport,
) []string {
	factors := []string{}

	// Check for decisive factors
	totalCrits := 0
	totalHealing := 0
	survivorCount := 0

	for _, report := range reports {
		totalCrits += report.Analytics.CriticalHits
		totalHealing += report.Analytics.HealingDone
		if report.Analytics.FinalHP > 0 && report.Analytics.CombatantType == "character" {
			survivorCount++
		}
	}

	if totalCrits > 5 {
		factors = append(factors, "Multiple critical hits turned the tide of battle")
	}

	if totalHealing > analytics.TotalDamageDealt/3 {
		factors = append(factors, "Effective healing kept the party in fighting shape")
	}

	if analytics.CombatDuration <= 3 {
		factors = append(factors, "Swift tactical execution ended combat quickly")
	}

	if survivorCount == len(reports) {
		factors = append(factors, "Excellent teamwork ensured no casualties")
	}

	return factors
}

// Helper functions removed - using built-in min/max from Go 1.21+

// calculateOverallScore calculates the overall tactical score
func calculateOverallScore(ta *models.TacticalAnalysis) int {
	return (ta.PositioningScore + ta.ResourceManagement + ta.TargetPrioritization + ta.TeamworkScore) / 4
}

// GetCombatAnalytics retrieves analytics for a combat
func (cas *CombatAnalyticsService) GetCombatAnalytics(_ context.Context, combatID uuid.UUID) (*models.CombatAnalytics, error) {
	return cas.analyticsRepo.GetCombatAnalytics(combatID)
}

// GetCombatantAnalytics retrieves combatant analytics for a combat
func (cas *CombatAnalyticsService) GetCombatantAnalytics(_ context.Context, analyticsID uuid.UUID) ([]*models.CombatantAnalytics, error) {
	return cas.analyticsRepo.GetCombatantAnalytics(analyticsID)
}

// GetCombatAnalyticsBySession retrieves all combat analytics for a session
func (cas *CombatAnalyticsService) GetCombatAnalyticsBySession(_ context.Context, sessionID uuid.UUID) ([]*models.CombatAnalytics, error) {
	return cas.analyticsRepo.GetCombatAnalyticsBySession(sessionID)
}
