// Authentication functionality
class Auth {
  constructor() {
    this.init();
  }
  
  init() {
    this.setupLoginForm();
    this.checkAuthStatus();
  }
  
  setupLoginForm() {
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
      loginForm.addEventListener('submit', (e) => {
        e.preventDefault();
        this.handleLogin();
      });
    }
  }
  
  async handleLogin() {
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value;
    const loginButton = document.getElementById('loginButton');
    const buttonText = document.getElementById('buttonText');
    const loadingSpinner = document.getElementById('loadingSpinner');
    
    // Validate input
    if (!username || !password) {
      this.showError('Please enter both username and password.');
      return;
    }
    
    // Show loading state
    this.setLoadingState(true, loginButton, buttonText, loadingSpinner);
    
    try {
      const response = await fetch('/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: username,
          password: password
        })
      });
      
      const data = await response.json();
      
      if (response.ok) {
        this.showSuccess('Login successful! Redirecting...');
        // Store auth token if provided
        if (data.token) {
          localStorage.setItem('authToken', data.token);
        }
        // Redirect to dashboard after a short delay
        setTimeout(() => {
          window.location.href = '/web/dashboard/';
        }, 1000);
      } else {
        this.showError(data.message || 'Login failed. Please check your credentials.');
      }
    } catch (error) {
      console.error('Login error:', error);
      this.showError('Network error. Please try again.');
    } finally {
      this.setLoadingState(false, loginButton, buttonText, loadingSpinner);
    }
  }
  
  setLoadingState(isLoading, button, buttonText, spinner) {
    if (isLoading) {
      button.disabled = true;
      buttonText.textContent = 'Signing In...';
      spinner.style.display = 'inline-block';
    } else {
      button.disabled = false;
      buttonText.textContent = 'Sign In';
      spinner.style.display = 'none';
    }
  }
  
  showError(message) {
    const errorDiv = document.getElementById('errorMessage');
    const successDiv = document.getElementById('successMessage');
    
    if (errorDiv) {
      errorDiv.textContent = message;
      errorDiv.style.display = 'block';
    }
    
    if (successDiv) {
      successDiv.style.display = 'none';
    }
  }
  
  showSuccess(message) {
    const successDiv = document.getElementById('successMessage');
    const errorDiv = document.getElementById('errorMessage');
    
    if (successDiv) {
      successDiv.textContent = message;
      successDiv.style.display = 'block';
    }
    
    if (errorDiv) {
      errorDiv.style.display = 'none';
    }
  }
  
  async checkAuthStatus() {
    const token = localStorage.getItem('authToken');
    if (token) {
      try {
        const response = await fetch('/auth/verify', {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        
        if (response.ok) {
          // User is already authenticated, redirect to dashboard
          if (window.location.pathname === '/login') {
            window.location.href = '/web/dashboard/';
          }
        } else {
          // Token is invalid, remove it
          localStorage.removeItem('authToken');
        }
      } catch (error) {
        console.error('Auth verification error:', error);
        localStorage.removeItem('authToken');
      }
    }
  }
  
  static logout() {
    localStorage.removeItem('authToken');
    window.location.href = '/login';
  }
  
  static isAuthenticated() {
    return !!localStorage.getItem('authToken');
  }
  
  static getAuthToken() {
    return localStorage.getItem('authToken');
  }
}

// Initialize auth when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  new Auth();
});

// Export for use in other scripts
window.Auth = Auth; 