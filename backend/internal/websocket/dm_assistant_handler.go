package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ctclostio/DnD-Game/backend/internal/models"
	"github.com/ctclostio/DnD-Game/backend/internal/services"
)

// DMAssistantMessage represents a DM Assistant WebSocket message
type DMAssistantMessage struct {
	Type      string                 `json:"type"`
	RequestID string                 `json:"requestId"`
	Data      map[string]interface{} `json:"data"`
}

// DMAssistantResponse represents a DM Assistant response
type DMAssistantResponse struct {
	Type      string      `json:"type"`
	RequestID string      `json:"requestId"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Streaming bool        `json:"streaming"`
	Complete  bool        `json:"complete"`
}

// HandleDMAssistantMessage processes DM Assistant messages over WebSocket
func (c *Client) HandleDMAssistantMessage(message []byte, dmAssistant *services.DMAssistantService) {
	var msg DMAssistantMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		c.sendError(msg.RequestID, "Invalid message format")
		return
	}

	// Ensure user is DM
	if c.role != "dm" {
		c.sendError(msg.RequestID, "DM privileges required")
		return
	}

	switch msg.Type {
	case "dm_assistant_request":
		c.handleDMAssistantRequest(msg, dmAssistant)
	case "dm_assistant_npc_dialog":
		c.handleNPCDialogue(msg, dmAssistant)
	case "dm_assistant_location":
		c.handleLocationGeneration(msg, dmAssistant)
	case "dm_assistant_combat":
		c.handleCombatNarration(msg, dmAssistant)
	case "dm_assistant_plot_twist":
		c.handlePlotTwist(msg, dmAssistant)
	case "dm_assistant_hazard":
		c.handleEnvironmentalHazard(msg, dmAssistant)
	default:
		c.sendError(msg.RequestID, fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

func (c *Client) handleDMAssistantRequest(msg DMAssistantMessage, dmAssistant *services.DMAssistantService) {
	// Parse the request
	requestType, _ := msg.Data["type"].(string)
	gameSessionID, _ := msg.Data["gameSessionId"].(string)
	parameters, _ := msg.Data["parameters"].(map[string]interface{})
	contextData, _ := msg.Data["context"].(map[string]interface{})

	req := models.DMAssistantRequest{
		Type:           requestType,
		GameSessionID:  gameSessionID,
		Parameters:     parameters,
		Context:        contextData,
		StreamResponse: true,
	}

	// Send initial response
	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "dm_assistant_response",
		RequestID: msg.RequestID,
		Streaming: true,
		Complete:  false,
		Data: map[string]string{
			"status": "processing",
		},
	})

	// Process the request
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userID, _ := uuid.Parse(c.id)
	result, err := dmAssistant.ProcessRequest(ctx, userID, req)
	if err != nil {
		c.sendError(msg.RequestID, err.Error())
		return
	}

	// Send the result
	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "dm_assistant_response",
		RequestID: msg.RequestID,
		Data:      result,
		Streaming: false,
		Complete:  true,
	})

	// Broadcast certain results to the game session
	if requestType == models.RequestTypeLocationDesc || requestType == models.RequestTypeEnvironmentalHazard {
		contentData, _ := json.Marshal(map[string]interface{}{
			"contentType": requestType,
			"content":     result,
		})
		c.broadcastToSession(gameSessionID, &Message{
			Type: "dm_content_update",
			Data: json.RawMessage(contentData),
		})
	}
}

func (c *Client) handleNPCDialogue(msg DMAssistantMessage, _ *services.DMAssistantService) {
	// Stream NPC dialog as it's generated
	npcName, _ := msg.Data["npcName"].(string)
	playerInput, _ := msg.Data["playerInput"].(string)

	// Send typing indicator
	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "npc_dialog_stream",
		RequestID: msg.RequestID,
		Streaming: true,
		Complete:  false,
		Data: map[string]interface{}{
			"npcName": npcName,
			"status":  "typing",
		},
	})

	// In a real implementation, you would stream the response
	// For now, we'll simulate with a complete response
	time.Sleep(500 * time.Millisecond) // Simulate thinking

	dialog := fmt.Sprintf("'%s? Well, that's an interesting question...'", playerInput)

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "npc_dialog_stream",
		RequestID: msg.RequestID,
		Streaming: false,
		Complete:  true,
		Data: map[string]interface{}{
			"npcName": npcName,
			"dialog":  dialog,
		},
	})
}

func (c *Client) handleLocationGeneration(msg DMAssistantMessage, _ *services.DMAssistantService) {
	locationType, _ := msg.Data["locationType"].(string)
	locationName, _ := msg.Data["locationName"].(string)

	// Send progress updates
	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "location_generation",
		RequestID: msg.RequestID,
		Streaming: true,
		Complete:  false,
		Data: map[string]interface{}{
			"status":   "generating",
			"progress": 0.25,
			"message":  "Creating location layout...",
		},
	})

	time.Sleep(500 * time.Millisecond)

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "location_generation",
		RequestID: msg.RequestID,
		Streaming: true,
		Complete:  false,
		Data: map[string]interface{}{
			"status":   "generating",
			"progress": 0.5,
			"message":  "Adding atmospheric details...",
		},
	})

	time.Sleep(500 * time.Millisecond)

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "location_generation",
		RequestID: msg.RequestID,
		Streaming: true,
		Complete:  false,
		Data: map[string]interface{}{
			"status":   "generating",
			"progress": 0.75,
			"message":  "Placing secrets and hazards...",
		},
	})

	// Final result would come from the DM Assistant service
	// This is just a placeholder
	location := map[string]interface{}{
		"name":        locationName,
		"type":        locationType,
		"description": "A mysterious location filled with wonder...",
	}

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "location_generation",
		RequestID: msg.RequestID,
		Streaming: false,
		Complete:  true,
		Data:      location,
	})
}

func (c *Client) handleCombatNarration(msg DMAssistantMessage, _ *services.DMAssistantService) {
	// Quick combat narration without streaming
	attackerName, _ := msg.Data["attackerName"].(string)
	targetName, _ := msg.Data["targetName"].(string)
	damage, _ := msg.Data["damage"].(float64)

	narration := fmt.Sprintf("%s strikes %s with devastating force, dealing %d damage!",
		attackerName, targetName, int(damage))

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "combat_narration",
		RequestID: msg.RequestID,
		Data: map[string]string{
			"narration": narration,
		},
		Complete: true,
	})

	// Broadcast to all players in the session
	if sessionID, ok := msg.Data["gameSessionId"].(string); ok {
		combatData, _ := json.Marshal(map[string]interface{}{
			"narration": narration,
			"attacker":  attackerName,
			"target":    targetName,
			"damage":    int(damage),
		})
		c.broadcastToSession(sessionID, &Message{
			Type: "combat_update",
			Data: json.RawMessage(combatData),
		})
	}
}

func (c *Client) handlePlotTwist(msg DMAssistantMessage, _ *services.DMAssistantService) {
	// Generate a plot twist based on current context
	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "plot_twist_generation",
		RequestID: msg.RequestID,
		Streaming: true,
		Complete:  false,
		Data: map[string]string{
			"status": "analyzing_story",
		},
	})

	// Simulate generation
	time.Sleep(1 * time.Second)

	plotTwist := map[string]interface{}{
		"title":       "The Betrayal of Trust",
		"description": "The helpful NPC the party has been relying on is revealed to be working for the villain...",
		"impact":      "major",
		"hints": []string{
			"They've been asking a lot of questions about the party's plans",
			"No one in town seems to know where they came from",
			"They always seem to disappear during combat",
		},
	}

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "plot_twist_generation",
		RequestID: msg.RequestID,
		Data:      plotTwist,
		Complete:  true,
	})
}

func (c *Client) handleEnvironmentalHazard(msg DMAssistantMessage, _ *services.DMAssistantService) {
	// Extract parameters (not used in this example implementation)
	// locationType, _ := msg.Data["locationType"].(string)
	// difficulty, _ := msg.Data["difficulty"].(float64)

	hazard := map[string]interface{}{
		"name":        "Unstable Floor",
		"description": "The ancient stonework gives way under pressure",
		"trigger":     "When more than 200 pounds of weight is applied",
		"effect":      "Fall 20 feet into a pit (2d6 damage)",
		"dc":          15,
		"detection":   "DC 15 Perception to notice the loose stones",
	}

	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "environmental_hazard",
		RequestID: msg.RequestID,
		Data:      hazard,
		Complete:  true,
	})
}

func (c *Client) sendDMAssistantResponse(response DMAssistantResponse) {
	data, _ := json.Marshal(response)
	c.send <- data
}

func (c *Client) sendError(requestID, errorMsg string) {
	c.sendDMAssistantResponse(DMAssistantResponse{
		Type:      "error",
		RequestID: requestID,
		Error:     errorMsg,
		Complete:  true,
	})
}

func (c *Client) broadcastToSession(sessionID string, message *Message) {
	// This would broadcast to all clients in the same game session
	// Implementation depends on your hub structure
	if message == nil {
		return
	}
	message.RoomID = sessionID
	data, _ := json.Marshal(message)
	c.hub.broadcast <- data
}
