// Alerts functionality
class AlertsManager {
  constructor() {
    this.alerts = [];
    this.currentEditId = null;
    this.init();
  }
  
  async init() {
    console.log('Initializing alerts manager...');
    await this.loadAlerts();
    this.setupEventListeners();
  }
  
  setupEventListeners() {
    // Form submission
    const alertForm = document.getElementById('alertForm');
    if (alertForm) {
      alertForm.addEventListener('submit', (e) => this.handleFormSubmit(e));
    }
    
    // Modal close events
    window.addEventListener('click', (e) => {
      if (e.target.classList.contains('modal')) {
        this.closeAlertModal();
        this.closeConfirmModal();
      }
    });
    
    // Escape key to close modals
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        this.closeAlertModal();
        this.closeConfirmModal();
      }
    });
  }
  
  async loadAlerts() {
    try {
      this.showLoading(true);
      
      const token = localStorage.getItem('authToken');
      const headers = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch('/alerts', { headers });
      if (!response.ok) {
        if (response.status === 401) {
          // Redirect to login if unauthorized
          window.location.href = '/login';
          return;
        }
        throw new Error('Failed to load alerts');
      }
      
      const data = await response.json();
      this.alerts = data.alerts || [];
      this.renderAlerts();
      
    } catch (error) {
      console.error('Error loading alerts:', error);
      this.showError('Failed to load alerts. Please try again.');
    } finally {
      this.showLoading(false);
    }
  }
  
  renderAlerts() {
    const alertsGrid = document.getElementById('alertsGrid');
    const noAlerts = document.getElementById('noAlerts');
    
    if (!alertsGrid || !noAlerts) return;
    
    if (this.alerts.length === 0) {
      alertsGrid.style.display = 'none';
      noAlerts.style.display = 'block';
      return;
    }
    
    alertsGrid.style.display = 'grid';
    noAlerts.style.display = 'none';
    
    alertsGrid.innerHTML = this.alerts.map(alert => this.createAlertCard(alert)).join('');
  }
  
  createAlertCard(alert) {
    const statusClass = alert.is_active ? 'active' : 'inactive';
    const statusText = alert.is_active ? 'Active' : 'Inactive';
    const statusIcon = alert.is_active ? 'fa-bell' : 'fa-bell-slash';
    
    const toggleBtnText = alert.is_active ? 'Deactivate' : 'Activate';
    const toggleBtnClass = alert.is_active ? 'deactivate' : 'toggle';
    const toggleBtnIcon = alert.is_active ? 'fa-pause' : 'fa-play';
    
    const createdAt = new Date(alert.created_at).toLocaleDateString();
    const updatedAt = new Date(alert.updated_at).toLocaleDateString();
    
    // Determine if this is an option
    const isOption = alert.option_type && (alert.option_type === 'CALL' || alert.option_type === 'PUT');
    
    return `
      <div class="alert-card ${alert.is_active ? '' : 'inactive'}" data-alert-id="${alert.id}">
        <div class="alert-header">
          <h3 class="alert-symbol">
            ${this.escapeHtml(alert.symbol)}
            ${isOption ? `<span style="font-size: 0.8em; color: #7f8c8d; margin-left: 0.5em;">(${alert.option_type})</span>` : ''}
          </h3>
          <span class="alert-status ${statusClass}">
            <i class="fas ${statusIcon}"></i>
            ${statusText}
          </span>
        </div>
        
        <div class="alert-details">
          ${isOption ? `
            <div class="alert-detail">
              <span class="alert-detail-label">Underlying:</span>
              <span class="alert-detail-value">${this.escapeHtml(alert.underlying_symbol || 'N/A')}</span>
            </div>
            <div class="alert-detail">
              <span class="alert-detail-label">Strike:</span>
              <span class="alert-detail-value">${alert.strike_price || 'N/A'}</span>
            </div>
            <div class="alert-detail">
              <span class="alert-detail-label">Expiry:</span>
              <span class="alert-detail-value">${alert.expiry || 'N/A'}</span>
            </div>
          ` : ''}
          <div class="alert-detail">
            <span class="alert-detail-label">Alert Type:</span>
            <span class="alert-detail-value">${this.formatAlertType(alert.alert_type)}</span>
          </div>
          <div class="alert-detail">
            <span class="alert-detail-label">Condition:</span>
            <span class="alert-detail-value">${this.formatCondition(alert.condition)} ${alert.target_value}</span>
          </div>
          <div class="alert-detail">
            <span class="alert-detail-label">Created:</span>
            <span class="alert-detail-value">${createdAt}</span>
          </div>
          <div class="alert-detail">
            <span class="alert-detail-label">Updated:</span>
            <span class="alert-detail-value">${updatedAt}</span>
          </div>
        </div>
        
        ${alert.message ? `<div class="alert-message">"${this.escapeHtml(alert.message)}"</div>` : ''}
        
        <div class="alert-actions">
          <button class="alert-btn edit" onclick="alertsManager.editAlert(${alert.id})">
            <i class="fas fa-edit"></i>
            Edit
          </button>
          <button class="alert-btn ${toggleBtnClass}" onclick="alertsManager.toggleAlert(${alert.id})">
            <i class="fas ${toggleBtnIcon}"></i>
            ${toggleBtnText}
          </button>
          <button class="alert-btn delete" onclick="alertsManager.deleteAlert(${alert.id})">
            <i class="fas fa-trash"></i>
            Delete
          </button>
        </div>
      </div>
    `;
  }
  
  formatAlertType(type) {
    const types = {
      'PRICE_ABOVE': 'Price Above',
      'PRICE_BELOW': 'Price Below',
      'PERCENTAGE_CHANGE': 'Percentage Change'
    };
    return types[type] || type;
  }
  
  formatCondition(condition) {
    const conditions = {
      '>': 'Greater than',
      '<': 'Less than',
      '>=': 'Greater than or equal to',
      '<=': 'Less than or equal to',
      '==': 'Equal to'
    };
    return conditions[condition] || condition;
  }
  
  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }
  
  async handleFormSubmit(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    const instrumentType = document.getElementById('instrumentType').value;
    
    const alertData = {
      symbol: formData.get('symbol') || document.getElementById('symbol').value,
      alert_type: formData.get('alertType') || document.getElementById('alertType').value,
      target_value: parseFloat(formData.get('targetValue') || document.getElementById('targetValue').value),
      condition: formData.get('condition') || document.getElementById('condition').value,
      message: formData.get('message') || document.getElementById('message').value
    };
    
    // Add options data if instrument type is option
    if (instrumentType === 'OPTION') {
      alertData.underlying_symbol = document.getElementById('underlyingSymbol').value;
      alertData.option_type = document.getElementById('optionType').value;
      alertData.strike_price = parseFloat(document.getElementById('strikePrice').value) || 0;
      alertData.expiry = document.getElementById('expiry').value;
    }
    
    // Validate required fields
    if (!alertData.symbol || !alertData.alert_type || !alertData.condition) {
      this.showError('Please fill in all required fields.');
      return;
    }
    
    try {
      const submitBtn = document.getElementById('submitBtn');
      const originalText = submitBtn.textContent;
      submitBtn.textContent = this.currentEditId ? 'Updating...' : 'Creating...';
      submitBtn.disabled = true;
      
      if (this.currentEditId) {
        await this.updateAlert(this.currentEditId, alertData);
      } else {
        await this.createAlert(alertData);
      }
      
      this.closeAlertModal();
      await this.loadAlerts();
      
    } catch (error) {
      console.error('Error submitting alert:', error);
      this.showError('Failed to save alert. Please try again.');
    } finally {
      const submitBtn = document.getElementById('submitBtn');
      submitBtn.textContent = originalText;
      submitBtn.disabled = false;
    }
  }
  
  async createAlert(alertData) {
    const token = localStorage.getItem('authToken');
    const headers = {
      'Content-Type': 'application/json'
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch('/alerts', {
      method: 'POST',
      headers,
      body: JSON.stringify(alertData)
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || 'Failed to create alert');
    }
    
    const data = await response.json();
    this.showSuccess(data.message || 'Alert created successfully');
  }
  
  async updateAlert(alertId, alertData) {
    const token = localStorage.getItem('authToken');
    const headers = {
      'Content-Type': 'application/json'
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(`/alerts?id=${alertId}`, {
      method: 'PUT',
      headers,
      body: JSON.stringify(alertData)
    });
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || 'Failed to update alert');
    }
    
    const data = await response.json();
    this.showSuccess(data.message || 'Alert updated successfully');
  }
  
  async deleteAlert(alertId) {
    const alert = this.alerts.find(a => a.id === alertId);
    if (!alert) return;
    
    this.showConfirmModal(
      `Are you sure you want to delete the alert for ${alert.symbol}?`,
      async () => {
        try {
          const token = localStorage.getItem('authToken');
          const headers = {};
          if (token) {
            headers['Authorization'] = `Bearer ${token}`;
          }
          
          const response = await fetch(`/alerts?id=${alertId}`, {
            method: 'DELETE',
            headers
          });
          
          if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.message || 'Failed to delete alert');
          }
          
          const data = await response.json();
          this.showSuccess(data.message || 'Alert deleted successfully');
          await this.loadAlerts();
          
        } catch (error) {
          console.error('Error deleting alert:', error);
          this.showError('Failed to delete alert. Please try again.');
        }
      }
    );
  }
  
  async toggleAlert(alertId) {
    try {
      const token = localStorage.getItem('authToken');
      const headers = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      
      const response = await fetch(`/alerts/toggle?id=${alertId}`, {
        method: 'PATCH',
        headers
      });
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'Failed to toggle alert');
      }
      
      const data = await response.json();
      this.showSuccess(data.message || 'Alert status updated successfully');
      await this.loadAlerts();
      
    } catch (error) {
      console.error('Error toggling alert:', error);
      this.showError('Failed to update alert status. Please try again.');
    }
  }
  
  editAlert(alertId) {
    const alert = this.alerts.find(a => a.id === alertId);
    if (!alert) return;
    
    this.currentEditId = alertId;
    
    // Populate form
    document.getElementById('symbol').value = alert.symbol;
    document.getElementById('alertType').value = alert.alert_type;
    document.getElementById('targetValue').value = alert.target_value;
    document.getElementById('condition').value = alert.condition;
    document.getElementById('message').value = alert.message || '';
    
    // Handle options fields
    const isOption = alert.option_type && (alert.option_type === 'CALL' || alert.option_type === 'PUT');
    if (isOption) {
      document.getElementById('instrumentType').value = 'OPTION';
      document.getElementById('underlyingSymbol').value = alert.underlying_symbol || '';
      document.getElementById('optionType').value = alert.option_type || '';
      document.getElementById('strikePrice').value = alert.strike_price || '';
      document.getElementById('expiry').value = alert.expiry || '';
      toggleOptionsFields(); // Show options fields
    } else {
      document.getElementById('instrumentType').value = 'STOCK';
      toggleOptionsFields(); // Hide options fields
    }
    
    // Update modal title and button
    document.getElementById('modalTitle').textContent = 'Edit Alert';
    document.getElementById('submitBtn').textContent = 'Update Alert';
    
    this.openAlertModal();
  }
  
  openAddAlertModal() {
    this.currentEditId = null;
    
    // Reset form
    document.getElementById('alertForm').reset();
    
    // Show options fields by default (since options is default)
    document.getElementById('optionsFields').style.display = 'block';
    
    // Update modal title and button
    document.getElementById('modalTitle').textContent = 'Add New Alert';
    document.getElementById('submitBtn').textContent = 'Create Alert';
    
    this.openAlertModal();
  }
  
  openAlertModal() {
    document.getElementById('alertModal').style.display = 'block';
    document.body.style.overflow = 'hidden';
  }
  
  closeAlertModal() {
    document.getElementById('alertModal').style.display = 'none';
    document.body.style.overflow = 'auto';
    this.currentEditId = null;
  }
  
  showConfirmModal(message, onConfirm) {
    document.getElementById('confirmMessage').textContent = message;
    document.getElementById('confirmBtn').onclick = () => {
      onConfirm();
      this.closeConfirmModal();
    };
    document.getElementById('confirmModal').style.display = 'block';
    document.body.style.overflow = 'hidden';
  }
  
  closeConfirmModal() {
    document.getElementById('confirmModal').style.display = 'none';
    document.body.style.overflow = 'auto';
  }
  
  showLoading(show) {
    const loading = document.getElementById('loading');
    const alertsGrid = document.getElementById('alertsGrid');
    
    if (loading && alertsGrid) {
      loading.style.display = show ? 'block' : 'none';
      if (show) {
        alertsGrid.style.display = 'none';
      }
    }
  }
  
  showSuccess(message) {
    // Create a simple success notification
    const notification = document.createElement('div');
    notification.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #27ae60;
      color: white;
      padding: 1em 1.5em;
      border-radius: 6px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      z-index: 10000;
      animation: slideIn 0.3s ease;
    `;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
      notification.style.animation = 'slideOut 0.3s ease';
      setTimeout(() => {
        if (notification.parentNode) {
          notification.parentNode.removeChild(notification);
        }
      }, 300);
    }, 3000);
  }
  
  showError(message) {
    // Create a simple error notification
    const notification = document.createElement('div');
    notification.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      background: #e74c3c;
      color: white;
      padding: 1em 1.5em;
      border-radius: 6px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
      z-index: 10000;
      animation: slideIn 0.3s ease;
    `;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
      notification.style.animation = 'slideOut 0.3s ease';
      setTimeout(() => {
        if (notification.parentNode) {
          notification.parentNode.removeChild(notification);
        }
      }, 5000);
    }, 5000);
  }
}

// Global functions for HTML onclick handlers
function openAddAlertModal() {
  alertsManager.openAddAlertModal();
}

function closeAlertModal() {
  alertsManager.closeAlertModal();
}

function closeConfirmModal() {
  alertsManager.closeConfirmModal();
}

// Bulk alert functions
function initializeBulkAlertTable() {
  // Add first row by default when page loads
  addBulkAlertRow();
}

function addBulkAlertRow() {
  const tbody = document.getElementById('bulkAlertTableBody');
  const rowIndex = tbody.children.length;
  
  const row = document.createElement('tr');
  row.innerHTML = `
    <td>
      <input type="text" class="bulk-input" data-field="symbol" placeholder="RELIANCE24JAN2500CE">
    </td>
    <td>
      <input type="text" class="bulk-input" data-field="underlying" placeholder="RELIANCE">
    </td>
    <td>
      <select class="bulk-input" data-field="optionType">
        <option value="CALL">Call</option>
        <option value="PUT">Put</option>
      </select>
    </td>
    <td>
      <input type="number" class="bulk-input" data-field="strike" placeholder="2500" step="0.01">
    </td>
    <td>
      <input type="date" class="bulk-input" data-field="expiry">
    </td>
    <td>
      <select class="bulk-input" data-field="alertType">
        <option value="PRICE_ABOVE">Price Above</option>
        <option value="PRICE_BELOW">Price Below</option>
        <option value="PERCENTAGE_CHANGE">Percentage Change</option>
      </select>
    </td>
    <td>
      <input type="number" class="bulk-input" data-field="target" placeholder="100" step="0.01">
    </td>
    <td>
      <select class="bulk-input" data-field="condition">
        <option value=">">></option>
        <option value="<"><</option>
        <option value=">=">>=</option>
        <option value="<="><=</option>
        <option value="==">==</option>
      </select>
    </td>
    <td>
      <input type="text" class="bulk-input" data-field="message" placeholder="Alert message">
    </td>
    <td style="text-align: center;">
      <button type="button" onclick="removeBulkAlertRow(this)" class="remove-row-btn">
        <i class="fas fa-trash"></i>
      </button>
    </td>
  `;
  
  tbody.appendChild(row);
}

function removeBulkAlertRow(button) {
  button.closest('tr').remove();
}

function clearBulkAlertRows() {
  document.getElementById('bulkAlertTableBody').innerHTML = '';
}

function loadSampleData() {
  clearBulkAlertRows();
  
  const sampleData = [
    {
      symbol: 'RELIANCE24JAN2500CE',
      underlying: 'RELIANCE',
      optionType: 'CALL',
      strike: '2500',
      expiry: '2024-01-25',
      alertType: 'PRICE_ABOVE',
      target: '50',
      condition: '>',
      message: 'RELIANCE Call above 50'
    },
    {
      symbol: 'RELIANCE24JAN2400PE',
      underlying: 'RELIANCE',
      optionType: 'PUT',
      strike: '2400',
      expiry: '2024-01-25',
      alertType: 'PRICE_ABOVE',
      target: '30',
      condition: '>',
      message: 'RELIANCE Put above 30'
    },
    {
      symbol: 'TCS24JAN4000CE',
      underlying: 'TCS',
      optionType: 'CALL',
      strike: '4000',
      expiry: '2024-01-25',
      alertType: 'PRICE_BELOW',
      target: '20',
      condition: '<',
      message: 'TCS Call below 20'
    }
  ];
  
  sampleData.forEach(data => {
    addBulkAlertRow();
    const lastRow = document.getElementById('bulkAlertTableBody').lastElementChild;
    const inputs = lastRow.querySelectorAll('.bulk-input');
    
    inputs.forEach(input => {
      const field = input.getAttribute('data-field');
      if (data[field]) {
        input.value = data[field];
      }
    });
  });
}

async function submitBulkAlerts() {
  const tbody = document.getElementById('bulkAlertTableBody');
  const rows = tbody.querySelectorAll('tr');
  
  if (rows.length === 0) {
    alertsManager.showError('Please add at least one alert row.');
    return;
  }
  
  const alerts = [];
  const errors = [];
  
  rows.forEach((row, index) => {
    const inputs = row.querySelectorAll('.bulk-input');
    const alertData = {};
    
    inputs.forEach(input => {
      const field = input.getAttribute('data-field');
      alertData[field] = input.value;
    });
    
    // Validate required fields
    if (!alertData.symbol || !alertData.alertType || !alertData.target || !alertData.condition) {
      errors.push(`Row ${index + 1}: Missing required fields`);
      return;
    }
    
    // Convert to proper format
    const alert = {
      symbol: alertData.symbol,
      underlying_symbol: alertData.underlying,
      option_type: alertData.optionType,
      strike_price: parseFloat(alertData.strike) || 0,
      expiry: alertData.expiry,
      alert_type: alertData.alertType,
      target_value: parseFloat(alertData.target),
      condition: alertData.condition,
      message: alertData.message
    };
    
    alerts.push(alert);
  });
  
  if (errors.length > 0) {
    alertsManager.showError('Validation errors:\n' + errors.join('\n'));
    return;
  }
  
  try {
    // Create alerts one by one
    let successCount = 0;
    let errorCount = 0;
    
    for (const alert of alerts) {
      try {
        await alertsManager.createAlert(alert);
        successCount++;
      } catch (error) {
        errorCount++;
        console.error('Failed to create alert:', error);
      }
    }
    
    if (successCount > 0) {
      alertsManager.showSuccess(`Successfully created ${successCount} alerts${errorCount > 0 ? `, ${errorCount} failed` : ''}`);
      clearBulkAlertRows();
      addBulkAlertRow(); // Add a new empty row
      await alertsManager.loadAlerts();
    } else {
      alertsManager.showError('Failed to create any alerts');
    }
    
  } catch (error) {
    alertsManager.showError('Failed to create alerts: ' + error.message);
  }
}

// Toggle options fields visibility
function toggleOptionsFields() {
  const instrumentType = document.getElementById('instrumentType').value;
  const optionsFields = document.getElementById('optionsFields');
  
  if (instrumentType === 'OPTION') {
    optionsFields.style.display = 'block';
    // Make option fields required
    document.getElementById('underlyingSymbol').required = true;
    document.getElementById('optionType').required = true;
    document.getElementById('strikePrice').required = true;
    document.getElementById('expiry').required = true;
  } else {
    optionsFields.style.display = 'none';
    // Remove required from option fields
    document.getElementById('underlyingSymbol').required = false;
    document.getElementById('optionType').required = false;
    document.getElementById('strikePrice').required = false;
    document.getElementById('expiry').required = false;
  }
}

// Initialize alerts manager when DOM is loaded
let alertsManager;
document.addEventListener('DOMContentLoaded', () => {
  alertsManager = new AlertsManager();
  initializeBulkAlertTable(); // Initialize the bulk alert table
});

// Add CSS animations for notifications
const style = document.createElement('style');
style.textContent = `
  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }
  
  @keyframes slideOut {
    from {
      transform: translateX(0);
      opacity: 1;
    }
    to {
      transform: translateX(100%);
      opacity: 0;
    }
  }
`;
document.head.appendChild(style);
