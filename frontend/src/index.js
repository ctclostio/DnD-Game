import './styles/main.css';
import { CharacterView } from './components/CharacterView.js';
import { DiceRollerView } from './components/DiceRollerView.js';
import { GameSessionView } from './components/GameSessionView.js';
import { WebSocketService } from './services/websocket.js';
import { ApiService } from './services/api.js';

class App {
    constructor() {
        this.currentView = null;
        this.api = new ApiService();
        this.ws = null;
        this.init();
    }

    init() {
        this.setupNavigation();
        this.loadView('character');
    }

    setupNavigation() {
        const navButtons = document.querySelectorAll('#main-nav button');
        navButtons.forEach(button => {
            button.addEventListener('click', (e) => {
                const view = e.target.dataset.view;
                this.loadView(view);
                
                // Update active button
                navButtons.forEach(btn => btn.classList.remove('active'));
                e.target.classList.add('active');
            });
        });
    }

    loadView(viewName) {
        const mainContent = document.getElementById('main-content');
        mainContent.innerHTML = '';

        switch(viewName) {
            case 'character':
                this.currentView = new CharacterView(mainContent, this.api);
                break;
            case 'dice':
                this.currentView = new DiceRollerView(mainContent, this.api);
                break;
            case 'game':
                this.currentView = new GameSessionView(mainContent, this.api);
                // Initialize WebSocket for game session
                if (!this.ws) {
                    this.ws = new WebSocketService();
                }
                document.getElementById('chat-panel').classList.remove('hidden');
                break;
            default:
                mainContent.innerHTML = '<h2>View not found</h2>';
        }

        if (viewName !== 'game') {
            document.getElementById('chat-panel').classList.add('hidden');
        }
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new App();
});