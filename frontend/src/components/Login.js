import authService from '../services/auth';

export class Login {
    constructor({ onLoginSuccess }) {
        this.onLoginSuccess = onLoginSuccess;
        this.username = '';
        this.password = '';
        this.error = '';
        this.loading = false;
    }

    async handleSubmit(e) {
        e.preventDefault();
        this.error = '';
        this.loading = true;
        this.updateUI();

        try {
            await authService.login(this.username, this.password);
            this.onLoginSuccess();
        } catch (err) {
            this.error = err.message || 'Login failed';
            this.updateUI();
        } finally {
            this.loading = false;
            this.updateUI();
        }
    }

    updateUI() {
        const container = document.getElementById('auth-container');
        if (container) {
            container.innerHTML = this.render();
            this.attachEventListeners();
        }
    }

    attachEventListeners() {
        const form = document.getElementById('login-form');
        if (form) {
            form.addEventListener('submit', (e) => this.handleSubmit(e));
        }

        const usernameInput = document.getElementById('username');
        if (usernameInput) {
            usernameInput.addEventListener('input', (e) => {
                this.username = e.target.value;
            });
        }

        const passwordInput = document.getElementById('password');
        if (passwordInput) {
            passwordInput.addEventListener('input', (e) => {
                this.password = e.target.value;
            });
        }

        const registerLink = document.getElementById('register-link');
        if (registerLink) {
            registerLink.addEventListener('click', (e) => {
                e.preventDefault();
                window.app.showRegister();
            });
        }
    }

    render() {
        return `
            <div class="login-container">
                <form id="login-form" class="login-form">
                    <h2>Login to D&D Game</h2>
                    
                    ${this.error ? `
                        <div class="error-message">
                            ${this.error}
                        </div>
                    ` : ''}

                    <div class="form-group">
                        <label for="username">Username</label>
                        <input
                            type="text"
                            id="username"
                            value="${this.username}"
                            required
                            ${this.loading ? 'disabled' : ''}
                        />
                    </div>

                    <div class="form-group">
                        <label for="password">Password</label>
                        <input
                            type="password"
                            id="password"
                            value="${this.password}"
                            required
                            ${this.loading ? 'disabled' : ''}
                        />
                    </div>

                    <button type="submit" ${this.loading ? 'disabled' : ''}>
                        ${this.loading ? 'Logging in...' : 'Login'}
                    </button>

                    <p class="register-link">
                        Don't have an account? <a href="#" id="register-link">Register here</a>
                    </p>
                </form>
            </div>
        `;
    }
}