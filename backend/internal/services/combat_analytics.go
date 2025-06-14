package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ctclostio/DnD-Game/backend/internal/constants"
	"github.com/ctclostio/DnD-Game/backend/internal/database"
	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/google/uuid"
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
func (cas *CombatAnalyticsService) TrackCombatAction(ctx context.Context, action *models.CombatActionLog) error {
	return cas.analyticsRepo.CreateCombatAction(action)
}

// FinalizeCombatAnalytics generates the final combat report when combat ends
func (cas *CombatAnalyticsService) FinalizeCombatAnalytics(
	ctx context.Context,
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
	// Calculate total damage and healing
	totalDamage := 0
	totalHealing := 0
	killingBlows := []map[string]interface{}{}

	// Track damage by combatant for MVP calculation
	damageByActor := make(map[string]int)

	for _, action := range actions {
		if action.ActionType == constants.ActionAttack || action.ActionType == "spell" {
			totalDamage += action.DamageDealt
			damageByActor[action.ActorID] += action.DamageDealt
		}

		if action.ActionType == constants.ActionHeal {
			totalHealing += action.DamageDealt // Healing stored as positive damage
		}

		// Check for killing blows
		if action.Outcome == "killing_blow" {
			killingBlows = append(killingBlows, map[string]interface{}{
				"dealer_id": action.ActorID,
				"target_id": action.TargetID,
				"damage":    action.DamageDealt,
			})
		}
	}

	// Determine MVP
	mvpID := ""
	mvpType := ""
	maxDamage := 0

	for actorID, damage := range damageByActor {
		if damage > maxDamage {
			maxDamage = damage
			mvpID = actorID
			// Determine if MVP is character or NPC
			for _, c := range combat.Combatants {
				if c.ID == actorID {
					mvpType = string(c.Type)
					break
				}
			}
		}
	}

	killingBlowsJSON, _ := json.Marshal(killingBlows)

	return &models.CombatAnalytics{
		ID:               uuid.New(),
		CombatID:         combatUUID,
		GameSessionID:    sessionID,
		CombatDuration:   combat.Round,
		TotalDamageDealt: totalDamage,
		TotalHealingDone: totalHealing,
		KillingBlows:     models.JSONB(killingBlowsJSON),
		MVPID:            mvpID,
		MVPType:          mvpType,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func (cas *CombatAnalyticsService) calculateCombatantAnalytics(
	analyticsID uuid.UUID,
	combat *models.Combat,
	actions []*models.CombatActionLog,
) []*models.CombatantReport {
	reports := []*models.CombatantReport{}

	// Create a map to track analytics for each combatant
	combatantStats := make(map[string]*models.CombatantAnalytics)

	// Initialize stats for all combatants
	for _, combatant := range combat.Combatants {
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
			// Find when they were defeated
			for _, action := range actions {
				if action.TargetID != nil && *action.TargetID == combatant.ID && action.Outcome == "killing_blow" {
					stats.RoundsSurvived = action.RoundNumber
					break
				}
			}
		}

		combatantStats[combatant.ID] = stats
	}

	// Process all actions to update stats
	for _, action := range actions {
		if stats, ok := combatantStats[action.ActorID]; ok {
			// Update attacker stats
			switch action.ActionType {
			case constants.ActionAttack:
				stats.AttacksMade++
				switch action.Outcome {
				case "hit", constants.ActionCritical:
					stats.AttacksHit++
					if action.Outcome == constants.ActionCritical {
						stats.CriticalHits++
					}
				case "miss":
					stats.AttacksMissed++
				case "critical_miss":
					stats.AttacksMissed++
					stats.CriticalMisses++
				}
				stats.DamageDealt += action.DamageDealt
			case "spell", "ability":
				stats.DamageDealt += action.DamageDealt
				// Track ability usage
				abilities := []string{}
				_ = json.Unmarshal(stats.AbilitiesUsed, &abilities)
				abilities = append(abilities, action.ActionType)
				abilitiesJSON, _ := json.Marshal(abilities)
				stats.AbilitiesUsed = models.JSONB(abilitiesJSON)
			case "heal":
				stats.HealingDone += action.DamageDealt
			}
		}

		// Update target stats
		if action.TargetID != nil {
			if stats, ok := combatantStats[*action.TargetID]; ok {
				switch action.ActionType {
				case constants.ActionAttack, "spell":
					stats.DamageTaken += action.DamageDealt
				case constants.ActionHeal:
					stats.HealingReceived += action.DamageDealt
				}

				// Track conditions
				if len(action.ConditionsApplied) > 0 {
					conditions := []string{}
					_ = json.Unmarshal(stats.ConditionsSuffered, &conditions)

					var newConditions []string
					_ = json.Unmarshal(action.ConditionsApplied, &newConditions)
					conditions = append(conditions, newConditions...)

					conditionsJSON, _ := json.Marshal(conditions)
					stats.ConditionsSuffered = models.JSONB(conditionsJSON)
				}
			}
		}

		// Track saves
		if action.ActionType == "save" {
			if stats, ok := combatantStats[action.ActorID]; ok {
				if action.Outcome == "success" {
					stats.SavesMade++
				} else {
					stats.SavesFailed++
				}
			}
		}
	}

	// Generate reports for each combatant
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

	// Damage efficiency
	if stats.AttacksMade > 0 {
		hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
		if hitRate > 0.75 {
			score += 3
		} else if hitRate > 0.5 {
			score += 2
		} else if hitRate > 0.25 {
			score += 1
		}
	}

	// Survival
	if stats.FinalHP > 0 {
		score += 2
		if stats.DamageTaken == 0 {
			score += 2 // No damage taken
		}
	}

	// Impact
	if stats.DamageDealt > stats.DamageTaken*2 {
		score += 2
	}

	// Critical hits
	if stats.CriticalHits > 0 {
		score += 1
	}

	// Healing contribution
	if stats.HealingDone > 0 {
		score += 2
	}

	if score >= 8 {
		return "excellent"
	} else if score >= 5 {
		return "good"
	} else if score >= 3 {
		return "fair"
	}
	return constants.EconomicPoor
}

func (cas *CombatAnalyticsService) generateHighlights(stats *models.CombatantAnalytics) []string {
	highlights := []string{}

	if stats.AttacksMade > 0 {
		hitRate := float64(stats.AttacksHit) / float64(stats.AttacksMade)
		if hitRate > 0.75 {
			highlights = append(highlights, fmt.Sprintf("Exceptional accuracy: %.0f%% hit rate", hitRate*100))
		}
	}

	if stats.CriticalHits > 1 {
		highlights = append(highlights, fmt.Sprintf("Scored %d critical hits", stats.CriticalHits))
	}

	if stats.DamageDealt > 50 {
		highlights = append(highlights, fmt.Sprintf("Dealt %d total damage", stats.DamageDealt))
	}

	if stats.HealingDone > 30 {
		highlights = append(highlights, fmt.Sprintf("Healed %d HP to allies", stats.HealingDone))
	}

	if stats.DamageTaken == 0 && stats.RoundsSurvived > 3 {
		highlights = append(highlights, "Avoided all damage")
	}

	if stats.SavesMade > stats.SavesFailed && stats.SavesMade > 2 {
		highlights = append(highlights, "Strong saving throws")
	}

	return highlights
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
		score += 1
	}

	if advantageousPositions > 3 {
		score += 2
	} else if advantageousPositions > 1 {
		score += 1
	}

	return min(10, score)
}

func (cas *CombatAnalyticsService) analyzeResourceUse(actions []*models.CombatActionLog) int {
	// Analyze spell slot and ability usage efficiency
	score := 5

	highLevelSpellsOnMinions := 0
	wastedHealing := 0
	efficientResourceUse := 0

	for _, action := range actions {
		if action.ResourcesUsed != nil {
			var resources map[string]interface{}
			_ = json.Unmarshal(action.ResourcesUsed, &resources)

			if spellLevel, ok := resources["spell_level"].(float64); ok {
				if spellLevel >= 3 && action.DamageDealt < 20 {
					highLevelSpellsOnMinions++
				} else if action.DamageDealt > int(spellLevel)*10 {
					efficientResourceUse++
				}
			}
		}

		if action.ActionType == constants.ActionHeal {
			// Check if healing was wasted (overhealing)
			if action.DamageDealt > 20 && action.Outcome == "overheal" {
				wastedHealing++
			}
		}
	}

	if highLevelSpellsOnMinions > 2 {
		score -= 2
	}

	if wastedHealing > 3 {
		score -= 1
	}

	if efficientResourceUse > 5 {
		score += 2
	}

	return max(1, min(10, score))
}

func (cas *CombatAnalyticsService) analyzeTargeting(actions []*models.CombatActionLog, combat *models.Combat) int {
	// Analyze target selection priorities
	score := 5

	// Track who was targeted and when
	targetPriority := make(map[string]int)
	dangerousEnemiesEliminated := 0

	for _, action := range actions {
		if action.TargetID != nil && action.ActionType == constants.ActionAttack {
			targetPriority[*action.TargetID]++

			if action.Outcome == "killing_blow" {
				// Check if this was a high-priority target
				for _, combatant := range combat.Combatants {
					if combatant.ID == *action.TargetID {
						// Simple heuristic: casters and high damage dealers are priority
						if combatant.Type == models.CombatantTypeNPC && action.RoundNumber < 5 {
							dangerousEnemiesEliminated++
						}
						break
					}
				}
			}
		}
	}

	if dangerousEnemiesEliminated > 0 {
		score += min(3, dangerousEnemiesEliminated)
	}

	return min(10, score)
}

func (cas *CombatAnalyticsService) analyzeTeamwork(actions []*models.CombatActionLog, reports []*models.CombatantReport) int {
	// Analyze coordination and teamwork
	score := 5

	comboAttacks := 0
	coordinatedHealing := 0
	setupActions := 0

	// Look for patterns indicating teamwork
	for i, action := range actions {
		// Check for combo attacks (multiple attacks on same target in same round)
		if action.ActionType == constants.ActionAttack && i > 0 {
			prevAction := actions[i-1]
			if prevAction.RoundNumber == action.RoundNumber &&
				prevAction.TargetID != nil && action.TargetID != nil &&
				*prevAction.TargetID == *action.TargetID {
				comboAttacks++
			}
		}

		// Check for timely healing
		if action.ActionType == constants.ActionHeal && action.TargetID != nil {
			// Was the target low on health?
			for _, report := range reports {
				if report.Analytics.CombatantID == *action.TargetID {
					if report.Analytics.DamageTaken > report.Analytics.FinalHP {
						coordinatedHealing++
					}
					break
				}
			}
		}

		// Check for setup actions (buffs, debuffs)
		if action.ActionType == "spell" || action.ActionType == "ability" {
			var conditions []string
			_ = json.Unmarshal(action.ConditionsApplied, &conditions)
			if len(conditions) > 0 {
				setupActions++
			}
		}
	}

	if comboAttacks > 5 {
		score += 2
	}

	if coordinatedHealing > 3 {
		score += 1
	}

	if setupActions > 4 {
		score += 2
	}

	return min(10, score)
}

func (cas *CombatAnalyticsService) findMissedOpportunities(actions []*models.CombatActionLog, combat *models.Combat) []string {
	opportunities := []string{}

	// Analyze for common tactical mistakes
	aoeOpportunities := 0
	healingDelays := 0

	for _, action := range actions {
		// Check for missed AoE opportunities
		if action.ActionType == constants.ActionAttack && action.TargetID != nil {
			// Count enemies in same round
			enemiesClose := 0
			for _, otherAction := range actions {
				if otherAction.RoundNumber == action.RoundNumber &&
					otherAction.ActorType == "npc" {
					enemiesClose++
				}
			}
			if enemiesClose >= 3 {
				aoeOpportunities++
			}
		}
	}

	if aoeOpportunities > 3 {
		opportunities = append(opportunities, "Multiple opportunities for area-of-effect spells were missed")
	}

	if healingDelays > 2 {
		opportunities = append(opportunities, "Healing was delayed, resulting in preventable unconsciousness")
	}

	// Check for poor resource management
	var lastRoundActions []*models.CombatActionLog
	for _, action := range actions {
		if action.RoundNumber == combat.Round {
			lastRoundActions = append(lastRoundActions, action)
		}
	}

	highLevelResourcesUnused := false
	for _, action := range lastRoundActions {
		if action.ResourcesUsed != nil {
			var resources map[string]interface{}
			_ = json.Unmarshal(action.ResourcesUsed, &resources)
			if slots, ok := resources["spell_slots_remaining"].(map[string]interface{}); ok {
				for level, remaining := range slots {
					if level >= "3" && remaining.(float64) > 0 {
						highLevelResourcesUnused = true
						break
					}
				}
			}
		}
	}

	if highLevelResourcesUnused {
		opportunities = append(opportunities, "High-level spell slots remained unused")
	}

	return opportunities
}

func (cas *CombatAnalyticsService) generateRecommendations(
	combat *models.Combat,
	reports []*models.CombatantReport,
	analysis *models.TacticalAnalysis,
) []string {
	recommendations := []string{}

	// Based on tactical analysis scores
	if analysis.PositioningScore < 5 {
		recommendations = append(recommendations, "Focus on using cover and terrain advantages more effectively")
	}

	if analysis.ResourceManagement < 5 {
		recommendations = append(recommendations, "Improve resource management - save high-level spells for tougher enemies")
	}

	if analysis.TargetPrioritization < 5 {
		recommendations = append(recommendations, "Prioritize dangerous enemies like spellcasters and high-damage dealers")
	}

	if analysis.TeamworkScore < 5 {
		recommendations = append(recommendations, "Coordinate attacks and support actions for better synergy")
	}

	// Based on individual performance
	poorPerformers := 0
	for _, report := range reports {
		if report.PerformanceRating == constants.EconomicPoor {
			poorPerformers++
		}
	}

	if poorPerformers > len(reports)/3 {
		recommendations = append(recommendations, "Consider adjusting difficulty or providing tactical guidance to struggling players")
	}

	// Based on combat duration
	if combat.Round > 10 {
		recommendations = append(recommendations, "Combat lasted very long - consider more aggressive tactics to speed up encounters")
	} else if combat.Round < 3 {
		recommendations = append(recommendations, "Combat ended very quickly - consider adding environmental challenges or reinforcements")
	}

	return recommendations
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

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// calculateOverallScore calculates the overall tactical score
func calculateOverallScore(ta *models.TacticalAnalysis) int {
	return (ta.PositioningScore + ta.ResourceManagement + ta.TargetPrioritization + ta.TeamworkScore) / 4
}

// GetCombatAnalytics retrieves analytics for a combat
func (cas *CombatAnalyticsService) GetCombatAnalytics(ctx context.Context, combatID uuid.UUID) (*models.CombatAnalytics, error) {
	return cas.analyticsRepo.GetCombatAnalytics(combatID)
}

// GetCombatantAnalytics retrieves combatant analytics for a combat
func (cas *CombatAnalyticsService) GetCombatantAnalytics(ctx context.Context, analyticsID uuid.UUID) ([]*models.CombatantAnalytics, error) {
	return cas.analyticsRepo.GetCombatantAnalytics(analyticsID)
}

// GetCombatAnalyticsBySession retrieves all combat analytics for a session
func (cas *CombatAnalyticsService) GetCombatAnalyticsBySession(ctx context.Context, sessionID uuid.UUID) ([]*models.CombatAnalytics, error) {
	return cas.analyticsRepo.GetCombatAnalyticsBySession(sessionID)
}
