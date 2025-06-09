package models

import (
	"time"
)

type Combat struct {
	ID               string              `json:"id"`
	GameSessionID    string              `json:"gameSessionId"`
	Name             string              `json:"name"`
	Round            int                 `json:"round"`
	CurrentTurn      int                 `json:"currentTurn"`
	Combatants       []Combatant         `json:"combatants"`
	TurnOrder        []string            `json:"turnOrder"` // Combatant IDs in initiative order
	ActiveEffects    []CombatEffect      `json:"activeEffects"`
	IsActive         bool                `json:"isActive"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
	
	// Aliases for backward compatibility
	SessionID        string              `json:"-"` // Alias for GameSessionID
	Status           CombatStatus        `json:"-"` // Alias for IsActive
	Turn             int                 `json:"-"` // Alias for CurrentTurn
	ActionHistory    []CombatAction      `json:"-"` // For test compatibility
}

type CombatStatus string

const (
	CombatStatusActive    CombatStatus = "active"
	CombatStatusPaused    CombatStatus = "paused"
	CombatStatusCompleted CombatStatus = "completed"
)

type CombatantType string

const (
	CombatantTypeCharacter CombatantType = "character"
	CombatantTypeNPC       CombatantType = "npc"
)

type Combatant struct {
	ID                string              `json:"id"`
	CharacterID       string              `json:"characterId,omitempty"`
	Name              string              `json:"name"`
	Type              CombatantType       `json:"type"`
	Initiative        int                 `json:"initiative"`
	InitiativeRoll    int                 `json:"initiativeRoll"`
	HP                int                 `json:"hp"`
	MaxHP             int                 `json:"maxHp"`
	TempHP            int                 `json:"tempHp"`
	AC                int                 `json:"ac"`
	Speed             int                 `json:"speed"`
	Condition         string              `json:"condition,omitempty"` // Simple status field
	
	// Action Economy
	Actions           int                 `json:"actions"`
	BonusActions      int                 `json:"bonusActions"`
	Reactions         int                 `json:"reactions"`
	Movement          int                 `json:"movement"`
	
	// Status
	Conditions        []Condition         `json:"conditions"`
	DeathSaves        DeathSaves          `json:"deathSaves"`
	IsConcentrating   bool                `json:"isConcentrating"`
	ConcentrationSpell string             `json:"concentrationSpell,omitempty"`
	
	// Combat Stats
	AttackBonus       int                 `json:"attackBonus"`
	SpellAttackBonus  int                 `json:"spellAttackBonus"`
	SpellSaveDC       int                 `json:"spellSaveDc"`
	
	// Resistances and Vulnerabilities
	Resistances       []DamageType        `json:"resistances"`
	Immunities        []DamageType        `json:"immunities"`
	Vulnerabilities   []DamageType        `json:"vulnerabilities"`
	DamageResistances []string            `json:"damageResistances,omitempty"`
	DamageImmunities  []string            `json:"damageImmunities,omitempty"`
	DamageVulnerabilities []string        `json:"damageVulnerabilities,omitempty"`
	
	// Ability Scores and Modifiers
	Abilities         map[string]int      `json:"abilities"`
	SavingThrows      map[string]int      `json:"savingThrows"`
	Skills            map[string]int      `json:"skills"`
	
	IsPlayerCharacter bool                `json:"isPlayerCharacter"`
	IsVisible         bool                `json:"isVisible"`
	Notes             string              `json:"notes,omitempty"`
	
	// Additional fields for test compatibility
	DeathSaveSuccesses int                `json:"deathSaveSuccesses,omitempty"`
	DeathSaveFailures  int                `json:"deathSaveFailures,omitempty"`
	Position          Position            `json:"position,omitempty"`
}

type DeathSaves struct {
	Successes int `json:"successes"`
	Failures  int `json:"failures"`
	IsStable  bool `json:"isStable"`
	IsDead    bool `json:"isDead"`
}

type Condition string

const (
	ConditionBlinded       Condition = "blinded"
	ConditionCharmed       Condition = "charmed"
	ConditionDeafened      Condition = "deafened"
	ConditionFrightened    Condition = "frightened"
	ConditionGrappled      Condition = "grappled"
	ConditionIncapacitated Condition = "incapacitated"
	ConditionInvisible     Condition = "invisible"
	ConditionParalyzed     Condition = "paralyzed"
	ConditionPetrified     Condition = "petrified"
	ConditionPoisoned      Condition = "poisoned"
	ConditionProne         Condition = "prone"
	ConditionRestrained    Condition = "restrained"
	ConditionStunned       Condition = "stunned"
	ConditionUnconscious   Condition = "unconscious"
	ConditionExhaustion1   Condition = "exhaustion1"
	ConditionExhaustion2   Condition = "exhaustion2"
	ConditionExhaustion3   Condition = "exhaustion3"
	ConditionExhaustion4   Condition = "exhaustion4"
	ConditionExhaustion5   Condition = "exhaustion5"
	ConditionExhaustion6   Condition = "exhaustion6"
	ConditionDodging       Condition = "dodging"
	ConditionStable        Condition = "stable"
	ConditionDead          Condition = "dead"
)

type DamageType string

const (
	DamageTypeAcid        DamageType = "acid"
	DamageTypeBludgeoning DamageType = "bludgeoning"
	DamageTypeCold        DamageType = "cold"
	DamageTypeFire        DamageType = "fire"
	DamageTypeForce       DamageType = "force"
	DamageTypeLightning   DamageType = "lightning"
	DamageTypeNecrotic    DamageType = "necrotic"
	DamageTypePiercing    DamageType = "piercing"
	DamageTypePoison      DamageType = "poison"
	DamageTypePsychic     DamageType = "psychic"
	DamageTypeRadiant     DamageType = "radiant"
	DamageTypeSlashing    DamageType = "slashing"
	DamageTypeThunder     DamageType = "thunder"
)

type CombatEffect struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	SourceID      string        `json:"sourceId"`
	TargetID      string        `json:"targetId"`
	Duration      int           `json:"duration"` // in rounds
	RemainingTime int           `json:"remainingTime"`
	EffectType    EffectType    `json:"effectType"`
	SaveDC        int           `json:"saveDc,omitempty"`
	SaveType      string        `json:"saveType,omitempty"`
}

type EffectType string

const (
	EffectTypeBuff   EffectType = "buff"
	EffectTypeDebuff EffectType = "debuff"
	EffectTypeDamage EffectType = "damage"
	EffectTypeHealing EffectType = "healing"
)

type CombatAction struct {
	ID             string          `json:"id"`
	CombatID       string          `json:"combatId"`
	Round          int             `json:"round"`
	ActorID        string          `json:"actorId"`
	ActionType     ActionType      `json:"actionType"`
	TargetID       string          `json:"targetId,omitempty"`
	Description    string          `json:"description"`
	Rolls          []Roll          `json:"rolls,omitempty"`
	Damage         []Damage        `json:"damage,omitempty"`
	Healing        int             `json:"healing,omitempty"`
	Effects        []string        `json:"effects,omitempty"`
	Timestamp      time.Time       `json:"timestamp"`
	
	// Additional fields for test compatibility
	Type           ActionType      `json:"type,omitempty"`
	WeaponName     string          `json:"weaponName,omitempty"`
	AttackBonus    int             `json:"attackBonus,omitempty"`
	DamageDice     string          `json:"damageDice,omitempty"`
	DamageBonus    int             `json:"damageBonus,omitempty"`
	DamageType     string          `json:"damageType,omitempty"`
	SpellName      string          `json:"spellName,omitempty"`
	SpellLevel     int             `json:"spellLevel,omitempty"`
	SpellDC        int             `json:"spellDC,omitempty"`
	SpellAttackBonus int          `json:"spellAttackBonus,omitempty"`
	SpellDamage    string          `json:"spellDamage,omitempty"`
	SpellDamageType string        `json:"spellDamageType,omitempty"`
	Movement       int             `json:"movement,omitempty"`
	NewPosition    Position        `json:"newPosition,omitempty"`
	Hit            bool            `json:"hit,omitempty"`
}

type ActionType string

const (
	ActionTypeAttack        ActionType = "attack"
	ActionTypeCast          ActionType = "cast"
	ActionTypeMove          ActionType = "move"
	ActionTypeDash          ActionType = "dash"
	ActionTypeDodge         ActionType = "dodge"
	ActionTypeHelp          ActionType = "help"
	ActionTypeHide          ActionType = "hide"
	ActionTypeReady         ActionType = "ready"
	ActionTypeSearch        ActionType = "search"
	ActionTypeUseItem       ActionType = "useItem"
	ActionTypeBonusAction   ActionType = "bonusAction"
	ActionTypeReaction      ActionType = "reaction"
	ActionTypeDeathSave     ActionType = "deathSave"
	ActionTypeConcentration ActionType = "concentration"
	ActionTypeSavingThrow   ActionType = "savingThrow"
	ActionTypeEndTurn       ActionType = "endTurn"
	ActionTypeCastSpell     ActionType = "castSpell"
)

type Roll struct {
	Type       RollType   `json:"type"`
	Dice       string     `json:"dice"`
	Modifier   int        `json:"modifier"`
	Result     int        `json:"result"`
	Individual []int      `json:"individual"`
	Advantage  bool       `json:"advantage"`
	Disadvantage bool     `json:"disadvantage"`
	Critical   bool       `json:"critical"`
	CriticalMiss bool     `json:"criticalMiss"`
}

type RollType string

// Position represents a location on the battle grid
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// CombatParticipant is an alias for Combatant to maintain compatibility
type CombatParticipant = Combatant

// GridPosition is an alias for Position to maintain compatibility
type GridPosition = Position

// ActionResult represents the result of a combat action
type ActionResult struct {
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
	Action        *CombatAction  `json:"action,omitempty"`
	DamageDealt   int            `json:"damageDealt,omitempty"`
	HealingDone   int            `json:"healingDone,omitempty"`
	TargetKilled  bool           `json:"targetKilled,omitempty"`
	NewConditions []Condition    `json:"newConditions,omitempty"`
	RemovedConditions []string   `json:"removedConditions,omitempty"`
}

const (
	RollTypeAttack       RollType = "attack"
	RollTypeDamage       RollType = "damage"
	RollTypeSavingThrow  RollType = "savingThrow"
	RollTypeAbilityCheck RollType = "abilityCheck"
	RollTypeInitiative   RollType = "initiative"
	RollTypeDeathSave    RollType = "deathSave"
	RollTypeConcentration RollType = "concentration"
)

type Damage struct {
	Amount int        `json:"amount"`
	Type   DamageType `json:"type"`
}

type CombatRequest struct {
	Action      ActionType      `json:"action"`
	ActorID     string          `json:"actorId"`
	TargetID    string          `json:"targetId,omitempty"`
	WeaponID    string          `json:"weaponId,omitempty"`
	SpellID     string          `json:"spellId,omitempty"`
	Movement    GridPosition    `json:"movement,omitempty"`
	Advantage   bool            `json:"advantage"`
	Disadvantage bool           `json:"disadvantage"`
	Description string          `json:"description,omitempty"`
}

type CombatUpdate struct {
	Type    UpdateType      `json:"type"`
	Combat  *Combat         `json:"combat,omitempty"`
	Action  *CombatAction   `json:"action,omitempty"`
	Message string          `json:"message,omitempty"`
}

type UpdateType string

const (
	UpdateTypeCombatStart    UpdateType = "combatStart"
	UpdateTypeCombatEnd      UpdateType = "combatEnd"
	UpdateTypeTurnStart      UpdateType = "turnStart"
	UpdateTypeTurnEnd        UpdateType = "turnEnd"
	UpdateTypeAction         UpdateType = "action"
	UpdateTypeCondition      UpdateType = "condition"
	UpdateTypeInitiative     UpdateType = "initiative"
	UpdateTypeHPChange       UpdateType = "hpChange"
	UpdateTypeDeathSave      UpdateType = "deathSave"
	UpdateTypeConcentration  UpdateType = "concentration"
)

// CombatantUpdate represents an update to a combatant's state
type CombatantUpdate struct {
	HP        int    `json:"hp,omitempty"`
	TempHP    int    `json:"tempHp,omitempty"`
	Condition string `json:"condition,omitempty"`
}

// RollDetails represents the details of a dice roll
type RollDetails struct {
	Total      int      `json:"total"`
	Rolls      []int    `json:"rolls"`
	Modifier   int      `json:"modifier"`
	Type       RollType `json:"type"`
	Advantage  bool     `json:"advantage"`
	Disadvantage bool   `json:"disadvantage"`
}

// DeathSaveResult represents the result of a death saving throw
type DeathSaveResult struct {
	Roll        int  `json:"roll"`
	Success     bool `json:"success"`
	CritSuccess bool `json:"critSuccess"`
	CritFailure bool `json:"critFailure"`
}