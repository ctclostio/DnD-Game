package testutil

// Test data constants to avoid duplication in tests
const (
	// Test user data
	TestEmail        = "test@example.com"
	TestUsername     = "testuser"
	TestUserID       = "user-456"
	TestUserID2      = "user-42"
	TestUserID3      = "user-123"
	TestPassword     = "SecurePass123!"
	TestPasswordWeak = "weak"
	TestPasswordHash = "$2a$10$hashedpassword"
	
	// Test IDs
	TestRequestID   = "request-123"
	TestTraceID     = "trace-789"
	TestSessionID   = "session-123"
	TestCharacterID = "char-123"
	TestCampaignID  = "campaign-123"
	TestItemID      = "item-123"
	
	// Test names
	TestCharacterName = "Test Character"
	TestCampaignName  = "Test Campaign"
	TestItemName      = "Test Item"
	TestRaceName      = "Shadow Elf"
	TestClassName     = "Test Class"
	
	// Common test strings
	TestNonexistent = "nonexistent"
	TestInvalid     = "invalid"
	TestMessage     = "test message"
	
	// Item names used in tests
	TestHealingPotion    = "Healing Potion"
	TestRingOfProtection = "Ring of Protection"
	TestSwordOfTesting   = "Sword of Testing"
	
	// HTTP test data
	TestHTTPMethod = "GET"
	TestHTTPPath   = "/api/test"
	
	// Settlement test data
	TestSettlementType = "Settlement type: %v"
	
	// API endpoints
	APICharacters   = "/api/v1/characters"
	APICharacterByID = "/api/v1/characters/{id}"
	APICharacterBase = "/api/v1/characters/"
	APIInventory    = "/inventory"
)