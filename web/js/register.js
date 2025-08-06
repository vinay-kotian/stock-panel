// Registration functionality
class Register {
  constructor() {
    this.init();
  }
  
  init() {
    this.setupRegisterForm();
    this.setupPasswordValidation();
  }
  
  setupRegisterForm() {
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
      registerForm.addEventListener('submit', (e) => {
        e.preventDefault();
        this.handleRegister();
      });
    }
  }
  
  setupPasswordValidation() {
    const passwordInput = document.getElementById('password');
    const confirmPasswordInput = document.getElementById('confirmPassword');
    
    if (passwordInput) {
      passwordInput.addEventListener('input', () => {
        this.validatePassword(passwordInput.value);
      });
    }
    
    if (confirmPasswordInput) {
      confirmPasswordInput.addEventListener('input', () => {
        this.validateConfirmPassword(passwordInput.value, confirmPasswordInput.value);
      });
    }
  }
  
  validatePassword(password) {
    const requirements = {
      length: password.length >= 8,
      uppercase: /[A-Z]/.test(password),
      lowercase: /[a-z]/.test(password),
      number: /\d/.test(password)
    };
    
    // Update requirement indicators
    document.getElementById('req-length').className = `requirement ${requirements.length ? 'met' : 'not-met'}`;
    document.getElementById('req-uppercase').className = `requirement ${requirements.uppercase ? 'met' : 'not-met'}`;
    document.getElementById('req-lowercase').className = `requirement ${requirements.lowercase ? 'met' : 'not-met'}`;
    document.getElementById('req-number').className = `requirement ${requirements.number ? 'met' : 'not-met'}`;
    
    return Object.values(requirements).every(req => req);
  }
  
  validateConfirmPassword(password, confirmPassword) {
    const confirmPasswordInput = document.getElementById('confirmPassword');
    if (confirmPasswordInput) {
      if (password !== confirmPassword) {
        confirmPasswordInput.setCustomValidity('Passwords do not match');
      } else {
        confirmPasswordInput.setCustomValidity('');
      }
    }
    return password === confirmPassword;
  }
  
  async handleRegister() {
    const username = document.getElementById('username').value.trim();
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    const registerButton = document.getElementById('registerButton');
    const buttonText = document.getElementById('buttonText');
    const loadingSpinner = document.getElementById('loadingSpinner');
    
    // Validate input
    if (!username || !email || !password || !confirmPassword) {
      this.showError('Please fill in all fields.');
      return;
    }
    
    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      this.showError('Please enter a valid email address.');
      return;
    }
    
    // Validate password requirements
    if (!this.validatePassword(password)) {
      this.showError('Please ensure your password meets all requirements.');
      return;
    }
    
    // Validate password confirmation
    if (!this.validateConfirmPassword(password, confirmPassword)) {
      this.showError('Passwords do not match.');
      return;
    }
    
    // Show loading state
    this.setLoadingState(true, registerButton, buttonText, loadingSpinner);
    
    try {
      const response = await fetch('/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: username,
          email: email,
          password: password
        })
      });
      
      const data = await response.json();
      
      if (response.ok) {
        this.showSuccess('Account created successfully! Redirecting to login...');
        // Redirect to login page after a short delay
        setTimeout(() => {
          window.location.href = '/login';
        }, 2000);
      } else {
        this.showError(data.message || 'Registration failed. Please try again.');
      }
    } catch (error) {
      console.error('Registration error:', error);
      this.showError('Network error. Please try again.');
    } finally {
      this.setLoadingState(false, registerButton, buttonText, loadingSpinner);
    }
  }
  
  setLoadingState(isLoading, button, buttonText, spinner) {
    if (isLoading) {
      button.disabled = true;
      buttonText.textContent = 'Creating Account...';
      spinner.style.display = 'inline-block';
    } else {
      button.disabled = false;
      buttonText.textContent = 'Create Account';
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
}

// Initialize registration when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  new Register();
});

// Export for use in other scripts
window.Register = Register; 