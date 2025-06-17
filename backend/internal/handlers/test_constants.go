package handlers

// API endpoint constants
const (
	APIv1Prefix       = "/api/v1"
	APISessionsPath   = "/api/v1/sessions/"
	APICharactersPath = "/api/v1/characters/"
	APICombatPath     = "/api/v1/combat/"
	APIAuthPath       = "/api/v1/auth/"
	APIAuthMePath     = "/api/v1/auth/me"
	APIInventoryPath  = "/api/v1/inventory/"
	APISessionsBase   = "/api/sessions/"
	APICharactersBase = "/api/characters/"
)

// HTTP header constants
const (
	ContentTypeHeader = "Content-Type"
	ContentTypeJSON   = "application/json"
	AuthHeader        = "Authorization"
	BearerPrefix      = "Bearer "
)

// Test data constants
const (
	TestEmail         = "test@example.com"
	TestPassword      = "securepassword123"
	TestUsername      = "testuser"
	TestSessionName   = "Test Session"
	TestCharacterName = "Test Character"
	TestItemName      = "Test Item"
	DefaultPassword   = "password123"
	EmailDomain       = "@example.com"
)

// Test campaign names and descriptions
const (
	FellowshipCampaignName = "The Fellowship Campaign"
	FellowshipCampaignDesc = "A journey to destroy the One Ring"
	FellowshipCampaignUpd  = "The Fellowship - Book II"
	FellowshipDescUpd      = "The journey continues through Moria"
	AnotherCampaignName    = "Another Campaign"
	AnotherCampaignDesc    = "A different adventure"
	WebSocketSessionName   = "WebSocket Test Session"
	WebSocketSessionDesc   = "Testing real-time features"
	ConcurrentSessionName  = "Concurrent Session"
	SecureSession1Name     = "Secure Session 1"
	SecureSession2Name     = "Secure Session 2"
	SessionToEndName       = "Session to End"
	SessionToEndDesc       = "This will be ended"
)

// Test character names
const (
	CharacterAragorn = "Aragorn"
	CharacterLegolas = "Legolas"
	CharacterGimli   = "Gimli"
	CharacterWSHero  = "WSHero"
	CharacterSecHero = "SecHero"
)

// Test ID constants
const (
	TestUserID        = "user-123"
	TestSessionID     = "session-123"
	TestCharacterID   = "char-123"
	TestItemID        = "item-123"
	TestCombatID      = "combat-123"
	TestEncounterID   = "encounter-123"
)

// Error messages
const (
	ErrInvalidInput    = "invalid input"
	ErrNotFound        = "not found"
	ErrUnauthorized    = "unauthorized"
	ErrForbidden       = "forbidden"
	ErrInternalServer  = "internal server error"
	ErrDontHaveAccess  = "don't have access"
	ErrAlreadyInSession = "already"
	ErrCharacterRequired = "character"
	ErrParticipantNotFound = "participant not found"
)

// JSON field names
const (
	CharacterIDField = "character_id"
	NameField        = "name"
	DescriptionField = "description"
	MaxPlayersField  = "max_players"
	IsActiveField    = "is_active"
	CodeField        = "code"
)

// Test user names
const (
	TestDMUsername      = "dm"
	TestPlayer1Username = "player1"
	TestPlayer2Username = "player2"
	TestPlayer3Username = "player3"
	TestOtherDMUsername = "otherdm"
	TestWSDMUsername    = "wsdm"
	TestWSPlayerUsername = "wsplayer"
	TestSecDM1Username  = "secdm1"
	TestSecDM2Username  = "secdm2"
	TestSecPlayerUsername = "secplayer"
	TestHackerUsername  = "hacker"
	TestConcDMUsername  = "concdm"
)

// Test session codes
const (
	SessionCodeSEC001 = "SEC001"
	SessionCodeSEC002 = "SEC002"
	SessionCodeCONC123 = "CONC123"
)

// Test email addresses
const (
	TestDMEmail        = "dm@example.com"
	TestPlayer1Email   = "player1@example.com"
	TestPlayer2Email   = "player2@example.com"
	TestPlayer3Email   = "player3@example.com"
	TestOtherDMEmail   = "otherdm@example.com"
	TestWSDMEmail      = "wsdm@example.com"
	TestWSPlayerEmail  = "wsplayer@example.com"
	TestSecDM1Email    = "secdm1@example.com"
	TestSecDM2Email    = "secdm2@example.com"
	TestSecPlayerEmail = "secplayer@example.com"
	TestHackerEmail    = "hacker@example.com"
	TestConcDMEmail    = "concdm@example.com"
)