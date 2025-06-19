package security

import (
	"testing"
)

func TestGenerateSecureID(t *testing.T) {
	// Test successful generation
	id, err := GenerateSecureID()
	if err != nil {
		t.Fatalf("GenerateSecureID failed: %v", err)
	}
	
	// Should be 32 characters (16 bytes hex encoded)
	if len(id) != 32 {
		t.Errorf("Expected ID length 32, got %d", len(id))
	}
	
	// Test uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := GenerateSecureID()
		if err != nil {
			t.Fatalf("GenerateSecureID failed on iteration %d: %v", i, err)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name      string
		length    int
		expectErr bool
	}{
		{"valid length", 16, false},
		{"zero length", 0, true},
		{"negative length", -1, true},
		{"large length", 64, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateSecureToken(tt.length)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("GenerateSecureToken failed: %v", err)
			}
			if len(token) == 0 && tt.length > 0 {
				t.Errorf("Expected non-empty token")
			}
		})
	}
}

func TestGenerateSecureInt(t *testing.T) {
	tests := []struct {
		name      string
		max       int64
		expectErr bool
	}{
		{"valid max", 100, false},
		{"max 1", 1, false},
		{"zero max", 0, true},
		{"negative max", -1, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testGenerateSecureIntCase(t, tt.max, tt.expectErr)
		})
	}
	
	testGenerateSecureIntDistribution(t)
}

func testGenerateSecureIntCase(t *testing.T, max int64, expectErr bool) {
	n, err := GenerateSecureInt(max)
	if expectErr {
		if err == nil {
			t.Errorf("Expected error but got none")
		}
		return
	}
	if err != nil {
		t.Fatalf("GenerateSecureInt failed: %v", err)
	}
	if n < 0 || n >= max {
		t.Errorf("Generated number %d outside range [0, %d)", n, max)
	}
}

func testGenerateSecureIntDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping distribution test in short mode")
	}
	
	const max = 10
	counts := generateDistributionCounts(t, max, 10000)
	verifyAllValuesGenerated(t, counts, max)
}

func generateDistributionCounts(t *testing.T, max int64, iterations int) map[int64]int {
	counts := make(map[int64]int)
	for i := 0; i < iterations; i++ {
		n, err := GenerateSecureInt(max)
		if err != nil {
			t.Fatalf("GenerateSecureInt failed: %v", err)
		}
		counts[n]++
	}
	return counts
}

func verifyAllValuesGenerated(t *testing.T, counts map[int64]int, max int64) {
	for i := int64(0); i < max; i++ {
		if counts[i] == 0 {
			t.Errorf("Value %d was never generated", i)
		}
	}
}

func TestGenerateSessionID(t *testing.T) {
	id, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("GenerateSessionID failed: %v", err)
	}
	
	// Should be non-empty
	if len(id) == 0 {
		t.Errorf("Expected non-empty session ID")
	}
	
	// Test uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := GenerateSessionID()
		if err != nil {
			t.Fatalf("GenerateSessionID failed: %v", err)
		}
		if ids[id] {
			t.Errorf("Duplicate session ID generated: %s", id)
		}
		ids[id] = true
	}
}