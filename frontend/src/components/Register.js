import authService from '../services/auth';

export class Register {
    constructor({ onRegisterSuccess }) {
        this.onRegisterSuccess = onRegisterSuccess;
        this.formData = {
            username: '',
            email: '',
            password: '',
            confirmPassword: ''
        };
        this.error = '';
        this.loading = false;
    }

    async handleSubmit(e) {
        e.preventDefault();
        this.error = '';

        // Validate passwords match
        if (this.formData.password !== this.formData.confirmPassword) {
            this.error = 'Passwords do not match';
            this.updateUI();
            return;
        }

        // Validate password length
        if (this.formData.password.length < 8) {
            this.error = 'Password must be at least 8 characters long';
            this.updateUI();
            return;
        }

        this.loading = true;
        this.updateUI();

        try {
            await authService.register(
                this.formData.username,
                this.formData.email,
                this.formData.password
            );
            this.onRegisterSuccess();
        } catch (err) {
            this.error = err.message || 'Registration failed';
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
        const form = document.getElementById('register-form');
        if (form) {
            form.addEventListener('submit', (e) => this.handleSubmit(e));
        }

        // Attach input listeners
        ['username', 'email', 'password', 'confirmPassword'].forEach(field => {
            const input = document.getElementById(field);
            if (input) {
                input.addEventListener('input', (e) => {
                    this.formData[field] = e.target.value;
                });
            }
        });

        const loginLink = document.getElementById('login-link');
        if (loginLink) {
            loginLink.addEventListener('click', (e) => {
                e.preventDefault();
                window.app.showLogin();
            });
        }
    }

    render() {
        return `
            <div class="register-container">
                <form id="register-form" class="register-form">
                    <h2>Create Your D&D Account</h2>
                    
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
                            name="username"
                            value="${this.formData.username}"
                            required
                            ${this.loading ? 'disabled' : ''}
                        />
                    </div>

                    <div class="form-group">
                        <label for="email">Email</label>
                        <input
                            type="email"
                            id="email"
                            name="email"
                            value="${this.formData.email}"
                            required
                            ${this.loading ? 'disabled' : ''}
                        />
                    </div>

                    <div class="form-group">
                        <label for="password">Password</label>
                        <input
                            type="password"
                            id="password"
                            name="password"
                            value="${this.formData.password}"
                            required
                            ${this.loading ? 'disabled' : ''}
                            minlength="8"
                        />
                    </div>

                    <div class="form-group">
                        <label for="confirmPassword">Confirm Password</label>
                        <input
                            type="password"
                            id="confirmPassword"
                            name="confirmPassword"
                            value="${this.formData.confirmPassword}"
                            required
                            ${this.loading ? 'disabled' : ''}
                        />
                    </div>

                    <button type="submit" ${this.loading ? 'disabled' : ''}>
                        ${this.loading ? 'Creating Account...' : 'Register'}
                    </button>

                    <p class="login-link">
                        Already have an account? <a href="#" id="login-link">Login here</a>
                    </p>
                </form>
            </div>
        `;
    }
}