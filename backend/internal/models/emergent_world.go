package models

import (
	"time"
)

// WorldState represents the current state of the living world.
type WorldState struct {
	ID            string                 `json:"id" db:"id"`
	SessionID     string                 `json:"session_id" db:"session_id"`
	CurrentTime   time.Time              `json:"current_time" db:"current_time"`
	LastSimulated time.Time              `json:"last_simulated" db:"last_simulated"`
	WorldData     map[string]interface{} `json:"world_data" db:"world_data"`
	IsActive      bool                   `json:"is_active" db:"is_active"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// NPCGoal represents an autonomous goal for an NPC.
type NPCGoal struct {
	ID          string                 `json:"id" db:"id"`
	NPCID       string                 `json:"npc_id" db:"npc_id"`
	GoalType    string                 `json:"goal_type" db:"goal_type"`
	Priority    int                    `json:"priority" db:"priority"`
	Description string                 `json:"description" db:"description"`
	Progress    float64                `json:"progress" db:"progress"`
	Parameters  map[string]interface{} `json:"parameters" db:"parameters"`
	Status      string                 `json:"status" db:"status"` // active, completed, failed, abandoned
	StartedAt   time.Time              `json:"started_at" db:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// NPCSchedule represents daily routines and activities.
type NPCSchedule struct {
	ID         string                 `json:"id" db:"id"`
	NPCID      string                 `json:"npc_id" db:"npc_id"`
	TimeOfDay  string                 `json:"time_of_day" db:"time_of_day"`
	Activity   string                 `json:"activity" db:"activity"`
	Location   string                 `json:"location" db:"location"`
	Parameters map[string]interface{} `json:"parameters" db:"parameters"`
}

// FactionPersonality represents the AI personality of a faction.
type FactionPersonality struct {
	ID               string                 `json:"id" db:"id"`
	FactionID        string                 `json:"faction_id" db:"faction_id"`
	Traits           map[string]float64     `json:"traits" db:"traits"` // aggressive, diplomatic, isolationist, etc.
	Values           map[string]float64     `json:"values" db:"values"` // honor, wealth, knowledge, etc.
	Memories         []FactionMemory        `json:"memories" db:"memories"`
	CurrentMood      string                 `json:"current_mood" db:"current_mood"`
	DecisionWeights  map[string]float64     `json:"decision_weights" db:"decision_weights"`
	LearningData     map[string]interface{} `json:"learning_data" db:"learning_data"`
	LastLearningTime time.Time              `json:"last_learning_time" db:"last_learning_time"`
}

// FactionMemory represents a significant event in faction history.
type FactionMemory struct {
	ID           string                 `json:"id"`
	EventType    string                 `json:"event_type"`
	Description  string                 `json:"description"`
	Impact       float64                `json:"impact"` // -1 to 1, negative = bad, positive = good
	Participants []string               `json:"participants"`
	Context      map[string]interface{} `json:"context"`
	Timestamp    time.Time              `json:"timestamp"`
	Decay        float64                `json:"decay"` // How quickly the memory fades
}

// FactionAgenda represents long-term faction goals.
type FactionAgenda struct {
	ID          string                 `json:"id" db:"id"`
	FactionID   string                 `json:"faction_id" db:"faction_id"`
	AgendaType  string                 `json:"agenda_type" db:"agenda_type"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	Priority    int                    `json:"priority" db:"priority"`
	Stages      []AgendaStage          `json:"stages" db:"stages"`
	Progress    float64                `json:"progress" db:"progress"`
	Status      string                 `json:"status" db:"status"`
	Parameters  map[string]interface{} `json:"parameters" db:"parameters"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
}

// AgendaStage represents a step in achieving a faction agenda.
type AgendaStage struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Conditions  map[string]interface{} `json:"conditions"`
	Actions     []string               `json:"actions"`
	IsComplete  bool                   `json:"is_complete"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// ProceduralCulture represents a generated culture with unique characteristics.
type ProceduralCulture struct {
	ID                string                 `json:"id" db:"id"`
	Name              string                 `json:"name" db:"name"`
	Language          CultureLanguage        `json:"language" db:"language"`
	Customs           []CultureCustom        `json:"customs" db:"customs"`
	ArtStyle          CultureArtStyle        `json:"art_style" db:"art_style"`
	BeliefSystem      CultureBeliefSystem    `json:"belief_system" db:"belief_system"`
	Values            map[string]float64     `json:"values" db:"values"`
	Taboos            []string               `json:"taboos" db:"taboos"`
	Greetings         map[string]string      `json:"greetings" db:"greetings"`
	Architecture      ArchitectureStyle      `json:"architecture" db:"architecture"`
	Cuisine           []CuisineElement       `json:"cuisine" db:"cuisine"`
	MusicStyle        MusicStyle             `json:"music_style" db:"music_style"`
	ClothingStyle     ClothingStyle          `json:"clothing_style" db:"clothing_style"`
	NamingConventions NamingConventions      `json:"naming_conventions" db:"naming_conventions"`
	SocialStructure   SocialStructure        `json:"social_structure" db:"social_structure"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
}

// CultureLanguage represents linguistic characteristics.
type CultureLanguage struct {
	Name           string            `json:"name"`
	Phonemes       []string          `json:"phonemes"`
	CommonWords    map[string]string `json:"common_words"`
	GrammarRules   []string          `json:"grammar_rules"`
	WritingSystem  string            `json:"writing_system"`
	Idioms         []LanguageIdiom   `json:"idioms"`
	HonorificRules []string          `json:"honorific_rules"`
}

// LanguageIdiom represents a cultural expression.
type LanguageIdiom struct {
	Expression string `json:"expression"`
	Meaning    string `json:"meaning"`
	Context    string `json:"context"`
	Formality  string `json:"formality"`
}

// CultureCustom represents a cultural practice or tradition.
type CultureCustom struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // ceremony, daily_practice, seasonal, lifecycle
	Description  string                 `json:"description"`
	Frequency    string                 `json:"frequency"`
	Participants string                 `json:"participants"`
	Significance float64                `json:"significance"`
	Requirements map[string]interface{} `json:"requirements"`
}

// CultureArtStyle represents artistic preferences.
type CultureArtStyle struct {
	PrimaryMediums   []string               `json:"primary_mediums"`
	CommonMotifs     []string               `json:"common_motifs"`
	ColorPalette     []string               `json:"color_palette"`
	StyleDescription string                 `json:"style_description"`
	SacredSymbols    []ArtSymbol            `json:"sacred_symbols"`
	Techniques       map[string]string      `json:"techniques"`
	Materials        []string               `json:"materials"`
	Influences       map[string]interface{} `json:"influences"`
}

// ArtSymbol represents a meaningful symbol in the culture.
type ArtSymbol struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Meaning      string `json:"meaning"`
	Usage        string `json:"usage"`
	Restrictions string `json:"restrictions"`
}

// CultureBeliefSystem represents religious/philosophical beliefs.
type CultureBeliefSystem struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // monotheistic, polytheistic, animistic, philosophical
	Deities      []CultureDeity         `json:"deities"`
	CoreBeliefs  []string               `json:"core_beliefs"`
	Practices    []ReligiousPractice    `json:"practices"`
	HolyDays     []HolyDay              `json:"holy_days"`
	Afterlife    string                 `json:"afterlife"`
	CreationMyth string                 `json:"creation_myth"`
	MoralCode    map[string]string      `json:"moral_code"`
	SacredTexts  []string               `json:"sacred_texts"`
	ClergyRanks  []string               `json:"clergy_ranks"`
	Miracles     map[string]interface{} `json:"miracles"`
}

// CultureDeity represents a deity in the belief system.
type CultureDeity struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Domain      []string `json:"domain"`
	Personality string   `json:"personality"`
	Symbol      string   `json:"symbol"`
	Alignment   string   `json:"alignment"`
	Followers   string   `json:"followers"`
}

// ReligiousPractice represents a religious ritual or practice.
type ReligiousPractice struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Frequency   string                 `json:"frequency"`
	Description string                 `json:"description"`
	Materials   []string               `json:"materials"`
	Duration    string                 `json:"duration"`
	Effects     map[string]interface{} `json:"effects"`
}

// HolyDay represents a religious holiday.
type HolyDay struct {
	Name         string   `json:"name"`
	Date         string   `json:"date"` // Can be seasonal or calendar-based
	Duration     string   `json:"duration"`
	Celebration  string   `json:"celebration"`
	Restrictions []string `json:"restrictions"`
	Traditions   []string `json:"traditions"`
}

// ArchitectureStyle represents building preferences.
type ArchitectureStyle struct {
	Name              string                   `json:"name"`
	Materials         []string                 `json:"materials"`
	CommonFeatures    []string                 `json:"common_features"`
	BuildingTypes     map[string]BuildingStyle `json:"building_types"`
	DefensiveElements []string                 `json:"defensive_elements"`
	Decorations       []string                 `json:"decorations"`
	TypicalLayout     string                   `json:"typical_layout"`
}

// BuildingStyle represents specific building characteristics.
type BuildingStyle struct {
	Purpose     string   `json:"purpose"`
	Size        string   `json:"size"`
	Features    []string `json:"features"`
	Materials   []string `json:"materials"`
	Inhabitants string   `json:"inhabitants"`
}

// CuisineElement represents food culture.
type CuisineElement struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"` // staple, delicacy, ceremonial, everyday
	Ingredients  []string `json:"ingredients"`
	Preparation  string   `json:"preparation"`
	Occasion     string   `json:"occasion"`
	Significance string   `json:"significance"`
	Taboos       []string `json:"taboos"`
}

// MusicStyle represents musical traditions.
type MusicStyle struct {
	Name        string   `json:"name"`
	Instruments []string `json:"instruments"`
	Scales      []string `json:"scales"`
	Rhythms     []string `json:"rhythms"`
	Occasions   []string `json:"occasions"`
	Themes      []string `json:"themes"`
	DanceStyles []string `json:"dance_styles"`
}

// ClothingStyle represents fashion and dress.
type ClothingStyle struct {
	EverydayWear   map[string]ClothingItem `json:"everyday_wear"`
	FormalWear     map[string]ClothingItem `json:"formal_wear"`
	CeremonialWear map[string]ClothingItem `json:"ceremonial_wear"`
	Colors         []string                `json:"colors"`
	Materials      []string                `json:"materials"`
	Jewelry        []string                `json:"jewelry"`
	StatusMarkers  map[string]string       `json:"status_markers"`
}

// ClothingItem represents a piece of clothing.
type ClothingItem struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	WornBy      string   `json:"worn_by"`
	Materials   []string `json:"materials"`
	Colors      []string `json:"colors"`
	Decorations []string `json:"decorations"`
}

// NamingConventions represents how people are named.
type NamingConventions struct {
	GivenNamePatterns  []string          `json:"given_name_patterns"`
	FamilyNamePatterns []string          `json:"family_name_patterns"`
	TitleFormats       []string          `json:"title_formats"`
	NicknameRules      []string          `json:"nickname_rules"`
	NameMeanings       map[string]string `json:"name_meanings"`
	TabooNames         []string          `json:"taboo_names"`
	NamingCeremonies   []string          `json:"naming_ceremonies"`
}

// SocialStructure represents societal organization.
type SocialStructure struct {
	Type        string            `json:"type"` // caste, class, egalitarian, etc.
	Classes     []SocialClass     `json:"classes"`
	Mobility    string            `json:"mobility"` // rigid, limited, fluid
	Leadership  string            `json:"leadership"`
	FamilyUnit  string            `json:"family_unit"`
	GenderRoles map[string]string `json:"gender_roles"`
	AgeRoles    map[string]string `json:"age_roles"`
	Outsiders   string            `json:"outsiders"` // How strangers are treated
}

// SocialClass represents a social stratum.
type SocialClass struct {
	Name         string   `json:"name"`
	Rank         int      `json:"rank"`
	Privileges   []string `json:"privileges"`
	Restrictions []string `json:"restrictions"`
	Occupations  []string `json:"occupations"`
	Markers      []string `json:"markers"` // Visual or behavioral markers
}

// EmergentWorldEvent represents a significant event in the living world simulation.
type EmergentWorldEvent struct {
	ID               string                 `json:"id" db:"id"`
	SessionID        string                 `json:"session_id" db:"session_id"`
	EventType        string                 `json:"event_type" db:"event_type"`
	Title            string                 `json:"title" db:"title"`
	Description      string                 `json:"description" db:"description"`
	Impact           map[string]interface{} `json:"impact" db:"impact"`
	AffectedEntities []string               `json:"affected_entities" db:"affected_entities"`
	Consequences     []EventConsequence     `json:"consequences" db:"consequences"`
	IsPlayerVisible  bool                   `json:"is_player_visible" db:"is_player_visible"`
	OccurredAt       time.Time              `json:"occurred_at" db:"occurred_at"`
}

// EventConsequence represents the outcome of a world event.
type EventConsequence struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Effect     string                 `json:"effect"`
	Magnitude  float64                `json:"magnitude"`
	Duration   string                 `json:"duration"`
	Parameters map[string]interface{} `json:"parameters"`
}

// SimulationLog tracks world simulation activities.
type SimulationLog struct {
	ID             string                 `json:"id" db:"id"`
	SessionID      string                 `json:"session_id" db:"session_id"`
	SimulationType string                 `json:"simulation_type" db:"simulation_type"`
	StartTime      time.Time              `json:"start_time" db:"start_time"`
	EndTime        time.Time              `json:"end_time" db:"end_time"`
	EventsCreated  int                    `json:"events_created" db:"events_created"`
	Details        map[string]interface{} `json:"details" db:"details"`
	Success        bool                   `json:"success" db:"success"`
	ErrorMessage   string                 `json:"error_message" db:"error_message"`
}
