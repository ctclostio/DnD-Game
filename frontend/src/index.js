import './styles/main.css';
import { CharacterView } from './components/CharacterView.js';
import { CharacterBuilderView } from './components/CharacterBuilderView.js';
import { DiceRollerView } from './components/DiceRollerView.js';
import { GameSessionView } from './components/GameSessionView.js';
import { CombatView } from './components/CombatView.js';
import { Login } from './components/Login.js';
import { Register } from './components/Register.js';
import { SpellSlotManager } from './components/SpellSlotManager.js';
import { ExperienceTracker } from './components/ExperienceTracker.js';
import { SkillCheckView } from './components/SkillCheckView.js';
import { WebSocketService } from './services/websocket.js';
import { ApiService } from './services/api.js';
import authService from './services/auth.js';

class App {
    constructor() {
        this.currentView = null;
        this.api = new ApiService();
        this.ws = null;
        this.init();
    }

    init() {
        // Check if user is authenticated
        if (!authService.isAuthenticated()) {
            this.showLogin();
            return;
        }

        this.setupApp();
    }

    setupApp() {
        this.setupNavigation();
        this.setupUserInfo();
        this.loadView('character');
        
        // Show main app content
        document.getElementById('app-container').style.display = 'block';
        document.getElementById('auth-container').style.display = 'none';
    }

    setupUserInfo() {
        const user = authService.getCurrentUser();
        const userInfo = document.getElementById('user-info');
        if (userInfo) {
            userInfo.innerHTML = `
                <span>Welcome, ${user.username} (${user.role})</span>
                <button onclick="app.logout()">Logout</button>
            `;
        }
    }

    showLogin() {
        document.getElementById('app-container').style.display = 'none';
        document.getElementById('auth-container').style.display = 'block';
        
        const loginComponent = new Login({
            onLoginSuccess: () => this.setupApp()
        });
        
        loginComponent.updateUI();
    }

    showRegister() {
        document.getElementById('app-container').style.display = 'none';
        document.getElementById('auth-container').style.display = 'block';
        
        const registerComponent = new Register({
            onRegisterSuccess: () => this.setupApp()
        });
        
        registerComponent.updateUI();
    }

    logout() {
        authService.logout().then(() => {
            window.location.reload();
        });
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
            case 'character-builder':
                this.currentView = new CharacterBuilderView(mainContent);
                break;
            case 'dice':
                this.currentView = new DiceRollerView(mainContent, this.api);
                break;
            case 'game':
                this.currentView = new GameSessionView(mainContent, this.api);
                // Initialize WebSocket for game session
                if (!this.ws) {
                    this.ws = new WebSocketService();
                    // Connect with room ID (you might want to get this from the game session)
                    const roomId = 'default-room'; // TODO: Get actual room ID from game session
                    this.ws.connect(roomId);
                }
                document.getElementById('chat-panel').classList.remove('hidden');
                break;
            case 'combat':
                this.currentView = new CombatView(mainContent, this.api);
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
    window.app = new App();
    
    // Create global instances for components that need to be accessed from other components
    window.skillCheckView = new SkillCheckView();
});