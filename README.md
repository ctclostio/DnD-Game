# D&D Online Game

A web-based Dungeons & Dragons game platform with real-time multiplayer support.

## Features

- **Character Management**: Create and manage D&D characters with full stats, skills, and equipment
- **Real-time Gameplay**: WebSocket-based multiplayer sessions with live dice rolls and chat
- **Dice Roller**: Comprehensive dice rolling system with all standard D&D dice
- **Game Sessions**: Create or join game sessions as a player or Dungeon Master
- **Combat System**: Initiative tracking, attack rolls, and damage calculation
- **Rule Integration**: Built-in D&D 5e rules for classes, races, and spells

## Tech Stack

### Backend
- Go 1.21+
- Gorilla Mux (HTTP routing)
- Gorilla WebSocket (Real-time communication)
- In-memory storage (easily replaceable with database)

### Frontend
- Vanilla JavaScript with modern ES6+ features
- Webpack for bundling
- WebSocket API for real-time features
- Responsive CSS design

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Node.js 16 or higher
- npm or yarn

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd DnD-Game
```

2. Install backend dependencies:
```bash
go mod download
```

3. Install frontend dependencies:
```bash
cd frontend
npm install
```

### Running the Application

#### Development Mode

1. Start the backend server:
```bash
make run-backend
```

2. In a separate terminal, start the frontend development server:
```bash
make run-frontend
```

3. Open your browser and navigate to `http://localhost:3000`

#### Production Mode

1. Build the frontend:
```bash
make build-frontend
```

2. Run the full application:
```bash
make run
```

3. Open your browser and navigate to `http://localhost:8080`

### Using Docker

```bash
docker-compose up
```

## Project Structure

```
DnD-Game/
├── backend/
│   ├── cmd/server/       # Application entry point
│   ├── internal/         # Internal packages
│   │   ├── handlers/     # HTTP request handlers
│   │   ├── models/       # Data models
│   │   ├── services/     # Business logic
│   │   ├── game/         # Game mechanics
│   │   └── websocket/    # WebSocket handling
│   └── pkg/              # Public packages
│       ├── dice/         # Dice rolling logic
│       └── rules/        # D&D rules engine
├── frontend/
│   ├── public/           # Static assets
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── services/     # API and WebSocket services
│   │   └── game/         # Game UI logic
│   └── build/            # Production build output
├── data/                 # Game data (classes, races, spells)
├── scripts/              # Utility scripts
├── docker-compose.yml    # Docker configuration
├── Makefile             # Build commands
└── README.md            # This file
```

## API Endpoints

### Characters
- `GET /api/v1/characters` - List all characters
- `GET /api/v1/characters/{id}` - Get character details
- `POST /api/v1/characters` - Create new character
- `PUT /api/v1/characters/{id}` - Update character

### Dice
- `POST /api/v1/dice/roll` - Roll dice with notation (e.g., "2d6+3")

### Game Sessions
- `POST /api/v1/game/session` - Create new game session
- `GET /api/v1/game/session/{id}` - Get session details

### WebSocket
- `WS /ws?room={roomId}` - Connect to game session (authentication via message after connection)

## WebSocket Events

### Client to Server
- `chat` - Send chat message
- `dice_roll` - Broadcast dice roll result
- `join` - Join game session

### Server to Client
- `chat` - Receive chat message
- `dice_roll` - Receive dice roll from another player
- `player_joined` - Player joined notification
- `player_left` - Player left notification

## Security

The application implements comprehensive security measures:

### Security Features
- **Security Headers**: CSP, HSTS, X-Frame-Options, and more
- **WebSocket Security**: Origin validation and post-connection authentication
- **Token Security**: JWT tokens never transmitted in URLs
- **CORS Protection**: Strict origin validation with environment-based configuration
- **Rate Limiting**: Configurable limits for auth (5/min) and API (100/min) endpoints
- **CSRF Protection**: Token-based protection for state-changing operations

### Configuration
See [SECURITY.md](./SECURITY.md) for detailed security configuration and best practices.

### Reporting Security Issues
Please report security vulnerabilities privately to security@yourdomain.com

## Production Deployment

### ⚠️ Important Security Notice
**Never use development Docker stages or configurations in production!**

### Quick Start
1. Copy environment template:
   ```bash
   cp .env.production.template .env.production
   # Edit .env.production with your values
   ```

2. Build production images:
   ```bash
   ./scripts/build-production.sh
   ```

3. Deploy with production compose:
   ```bash
   docker-compose -f docker-compose.production.yml up -d
   ```

### Security Requirements
- Source maps are disabled in production builds
- Use `--target production` for frontend builds
- Use `--target final` for backend builds
- Set `ENV=production` (never `development`)
- JWT secrets must be 64+ characters
- Database SSL is required

For detailed production deployment instructions, see [DOCKER_PRODUCTION_GUIDE.md](./DOCKER_PRODUCTION_GUIDE.md).

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- D&D 5e SRD for game rules and content
- The Go community for excellent libraries
- All contributors and testers