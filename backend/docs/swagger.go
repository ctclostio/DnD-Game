// Package docs provides OpenAPI/Swagger documentation for the D&D Game API
package docs

import (
	"github.com/swaggo/swag"
)

// @title D&D Game API
// @version 1.0
// @description API for the D&D online game platform with real-time multiplayer support
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/ctclostio/DnD-Game/issues
// @contact.email support@dndgame.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @schemes http https
// @produce json
// @consumes json

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api/v1",
	Schemes:          []string{"http", "https"},
	Title:            "D&D Game API",
	Description:      "API for the D&D online game platform with real-time multiplayer support",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

// General API responses
// @Success 200 {object} map[string]interface{} "Success"
// @Success 201 {object} map[string]interface{} "Created"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 429 {object} map[string]string "Too Many Requests"
// @Failure 500 {object} map[string]string "Internal Server Error"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "API Support",
            "url": "https://github.com/ctclostio/DnD-Game/issues",
            "email": "support@dndgame.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {},
    "definitions": {},
    "securityDefinitions": {
        "Bearer": {
            "description": "Type \"Bearer\" followed by a space and JWT token.",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`
