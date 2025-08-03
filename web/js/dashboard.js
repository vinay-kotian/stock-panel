// Dashboard functionality
class Dashboard {
  constructor() {
    this.pnlChart = null;
    this.allPnlData = [];
    this.allTradesData = [];
    this.currentFilter = null;
    this.init();
  }
  
  async init() {
    console.log('Initializing dashboard...');
    await this.loadDashboardData();
    this.setupCharts();
    this.setupDateFilter();
  }
  
  async loadDashboardData() {
    try {
      // Load P&L data
      const pnlResponse = await fetch('/pnl');
      if (!pnlResponse.ok) throw new Error('Failed to load P&L data');
      this.allPnlData = await pnlResponse.json();
      
      // Load recent trades
      const tradesResponse = await fetch('/stocks');
      if (!tradesResponse.ok) throw new Error('Failed to load trades data');
      this.allTradesData = await tradesResponse.json();
      
      this.updateDashboard(this.allPnlData, this.allTradesData);
      
    } catch (error) {
      console.error('Error loading dashboard data:', error);
      this.showError('Failed to load dashboard data. Please try again.');
    }
  }
  
  setupDateFilter() {
    const startDateInput = document.getElementById('startDate');
    const endDateInput = document.getElementById('endDate');
    const applyFilterBtn = document.getElementById('applyFilter');
    const resetFilterBtn = document.getElementById('resetFilter');
    
    // Set default date range (last 30 days)
    const today = new Date();
    const thirtyDaysAgo = new Date(today);
    thirtyDaysAgo.setDate(today.getDate() - 30);
    
    endDateInput.value = today.toISOString().split('T')[0];
    startDateInput.value = thirtyDaysAgo.toISOString().split('T')[0];
    
    // Apply filter button
    applyFilterBtn.addEventListener('click', () => {
      this.applyDateFilter();
    });
    
    // Reset filter button
    resetFilterBtn.addEventListener('click', () => {
      this.resetDateFilter();
    });
    
    // Enter key support
    startDateInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.applyDateFilter();
    });
    
    endDateInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') this.applyDateFilter();
    });
  }
  
  applyDateFilter() {
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;
    const filterStatus = document.getElementById('filterStatus');
    
    if (!startDate || !endDate) {
      this.showFilterStatus('Please select both start and end dates.', 'error');
      return;
    }
    
    if (new Date(startDate) > new Date(endDate)) {
      this.showFilterStatus('Start date cannot be after end date.', 'error');
      return;
    }
    
    // Filter P&L data
    const filteredPnlData = this.allPnlData.filter(p => {
      const pnlDate = new Date(p.date);
      const start = new Date(startDate);
      const end = new Date(endDate);
      return pnlDate >= start && pnlDate <= end;
    });
    
    // Filter trades data
    const filteredTradesData = this.allTradesData.filter(t => {
      const tradeDate = new Date(t.timestamp.split('T')[0]);
      const start = new Date(startDate);
      const end = new Date(endDate);
      return tradeDate >= start && tradeDate <= end;
    });
    
    this.currentFilter = { startDate, endDate };
    this.updateDashboard(filteredPnlData, filteredTradesData);
    
    const dateRange = `${formatDate(startDate)} - ${formatDate(endDate)}`;
    this.showFilterStatus(`Showing data for: ${dateRange}`, 'success');
  }
  
  resetDateFilter() {
    document.getElementById('startDate').value = '';
    document.getElementById('endDate').value = '';
    this.currentFilter = null;
    this.updateDashboard(this.allPnlData, this.allTradesData);
    this.showFilterStatus('Showing all data', 'info');
  }
  
  showFilterStatus(message, type) {
    const filterStatus = document.getElementById('filterStatus');
    filterStatus.textContent = message;
    filterStatus.style.display = 'block';
    
    // Set color based on type
    switch (type) {
      case 'success':
        filterStatus.style.backgroundColor = '#d4edda';
        filterStatus.style.color = '#155724';
        filterStatus.style.border = '1px solid #c3e6cb';
        break;
      case 'error':
        filterStatus.style.backgroundColor = '#f8d7da';
        filterStatus.style.color = '#721c24';
        filterStatus.style.border = '1px solid #f5c6cb';
        break;
      case 'info':
        filterStatus.style.backgroundColor = '#d1ecf1';
        filterStatus.style.color = '#0c5460';
        filterStatus.style.border = '1px solid #bee5eb';
        break;
    }
    
    // Hide after 3 seconds
    setTimeout(() => {
      filterStatus.style.display = 'none';
    }, 3000);
  }
  
  updateDashboard(pnlData, tradesData) {
    this.updateMetrics(pnlData, tradesData);
    this.updatePerformanceSummary(pnlData);
    this.updateRecentActivity(tradesData);
    this.createPnLChart(pnlData);
  }
  
  updateMetrics(pnlData, tradesData) {
    if (pnlData.length === 0) {
      this.showNoData();
      return;
    }
    
    // Calculate metrics
    const totalPnl = pnlData.reduce((sum, p) => sum + p.pnl, 0);
    const profitableDays = pnlData.filter(p => p.pnl > 0).length;
    const tradingDays = pnlData.length;
    const winRate = tradingDays > 0 ? (profitableDays / tradingDays * 100) : 0;
    const avgDailyPnl = tradingDays > 0 ? (totalPnl / tradingDays) : 0;
    
    // Update DOM
    document.getElementById('totalPnl').textContent = `₹${totalPnl.toFixed(2)}`;
    document.getElementById('totalPnl').className = `metric-value ${totalPnl >= 0 ? 'pnl-positive' : 'pnl-negative'}`;
    
    document.getElementById('winRate').textContent = `${winRate.toFixed(1)}%`;
    document.getElementById('winRate').className = `metric-value ${winRate >= 50 ? 'pnl-positive' : 'pnl-negative'}`;
    
    document.getElementById('tradingDays').textContent = tradingDays;
    
    document.getElementById('avgDailyPnl').textContent = `₹${avgDailyPnl.toFixed(2)}`;
    document.getElementById('avgDailyPnl').className = `metric-value ${avgDailyPnl >= 0 ? 'pnl-positive' : 'pnl-negative'}`;
  }
  
  updatePerformanceSummary(pnlData) {
    if (pnlData.length === 0) return;
    
    const bestDay = pnlData.reduce((best, p) => p.pnl > best.pnl ? p : best);
    const worstDay = pnlData.reduce((worst, p) => p.pnl < worst.pnl ? p : worst);
    const totalTrades = pnlData.length;
    const profitableTrades = pnlData.filter(p => p.pnl > 0).length;
    const losingTrades = pnlData.filter(p => p.pnl < 0).length;
    const breakEvenTrades = pnlData.filter(p => p.pnl === 0).length;
    
    const summaryHtml = `
      <div class="performance-item">
        <div class="performance-value pnl-positive">${formatDate(bestDay.date)}</div>
        <div class="performance-label">Best Day</div>
        <div class="performance-value pnl-positive">₹${bestDay.pnl.toFixed(2)}</div>
      </div>
      
      <div class="performance-item">
        <div class="performance-value pnl-negative">${formatDate(worstDay.date)}</div>
        <div class="performance-label">Worst Day</div>
        <div class="performance-value pnl-negative">₹${worstDay.pnl.toFixed(2)}</div>
      </div>
      
      <div class="performance-item">
        <div class="performance-value">${profitableTrades}</div>
        <div class="performance-label">Profitable Days</div>
      </div>
      
      <div class="performance-item">
        <div class="performance-value">${losingTrades}</div>
        <div class="performance-label">Losing Days</div>
      </div>
      
      <div class="performance-item">
        <div class="performance-value">${breakEvenTrades}</div>
        <div class="performance-label">Break Even Days</div>
      </div>
      
      <div class="performance-item">
        <div class="performance-value">${totalTrades}</div>
        <div class="performance-label">Total Trading Days</div>
      </div>
    `;
    
    document.getElementById('performanceSummary').innerHTML = summaryHtml;
  }
  
  updateRecentActivity(tradesData) {
    if (tradesData.length === 0) {
      document.getElementById('recentActivity').innerHTML = '<div class="loading">No trading activity found.</div>';
      return;
    }
    
    // Get last 10 trades
    const recentTrades = tradesData
      .sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp))
      .slice(0, 10);
    
    const activityHtml = `
      <table style="width: 100%; border-collapse: collapse;">
        <thead>
          <tr style="background: #f8f9fa;">
            <th style="padding: 0.8em; text-align: left; border-bottom: 2px solid #dee2e6;">Symbol</th>
            <th style="padding: 0.8em; text-align: left; border-bottom: 2px solid #dee2e6;">Type</th>
            <th style="padding: 0.8em; text-align: left; border-bottom: 2px solid #dee2e6;">Side</th>
            <th style="padding: 0.8em; text-align: left; border-bottom: 2px solid #dee2e6;">Price</th>
            <th style="padding: 0.8em; text-align: left; border-bottom: 2px solid #dee2e6;">Date</th>
          </tr>
        </thead>
        <tbody>
          ${recentTrades.map(trade => `
            <tr style="border-bottom: 1px solid #f1f3f4;">
              <td style="padding: 0.8em;">${trade.symbol}</td>
              <td style="padding: 0.8em;">${trade.option_type}</td>
              <td style="padding: 0.8em;">
                <span style="color: ${trade.side === 'BUY' ? '#28a745' : '#dc3545'}; font-weight: bold;">
                  ${trade.side}
                </span>
              </td>
              <td style="padding: 0.8em;">₹${trade.price}</td>
              <td style="padding: 0.8em;">${formatDate(trade.timestamp.split('T')[0])}</td>
            </tr>
          `).join('')}
        </tbody>
      </table>
    `;
    
    document.getElementById('recentActivity').innerHTML = activityHtml;
  }
  
  createPnLChart(pnlData) {
    if (pnlData.length === 0) return;
    
    // Sort data by date
    const sortedData = pnlData.sort((a, b) => new Date(a.date) - new Date(b.date));
    
    // Calculate cumulative P&L
    let cumulativePnl = 0;
    const chartData = sortedData.map(p => {
      cumulativePnl += p.pnl;
      return {
        date: formatDate(p.date),
        pnl: p.pnl,
        cumulative: cumulativePnl
      };
    });
    
    const ctx = document.getElementById('pnlChart').getContext('2d');
    
    // Destroy existing chart if it exists
    if (this.pnlChart) {
      this.pnlChart.destroy();
    }
    
    this.pnlChart = new Chart(ctx, {
      type: 'line',
      data: {
        labels: chartData.map(d => d.date),
        datasets: [
          {
            label: 'Daily P&L',
            data: chartData.map(d => d.pnl),
            borderColor: '#3498db',
            backgroundColor: 'rgba(52, 152, 219, 0.1)',
            borderWidth: 2,
            fill: false,
            tension: 0.4
          },
          {
            label: 'Cumulative P&L',
            data: chartData.map(d => d.cumulative),
            borderColor: '#2ecc71',
            backgroundColor: 'rgba(46, 204, 113, 0.1)',
            borderWidth: 3,
            fill: false,
            tension: 0.4
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: {
            position: 'top',
            labels: {
              boxWidth: 12,
              padding: 10,
              font: {
                size: 12
              }
            }
          },
          title: {
            display: true,
            text: 'P&L Performance Over Time',
            font: {
              size: 14,
              weight: 'bold'
            },
            padding: {
              bottom: 10
            }
          }
        },
        scales: {
          x: {
            ticks: {
              maxRotation: 45,
              minRotation: 0,
              font: {
                size: 11
              }
            }
          },
          y: {
            beginAtZero: true,
            ticks: {
              callback: function(value) {
                return '₹' + value.toFixed(2);
              },
              font: {
                size: 11
              }
            }
          }
        },
        interaction: {
          intersect: false,
          mode: 'index'
        },
        elements: {
          point: {
            radius: 3,
            hoverRadius: 5
          }
        },
        layout: {
          padding: {
            top: 10,
            bottom: 10
          }
        }
      }
    });
  }
  
  setupCharts() {
    // Additional chart setup if needed
    console.log('Charts setup complete');
  }
  
  showError(message) {
    const errorHtml = `<div class="error">${message}</div>`;
    document.getElementById('totalPnl').parentElement.innerHTML = errorHtml;
    document.getElementById('performanceSummary').innerHTML = errorHtml;
    document.getElementById('recentActivity').innerHTML = errorHtml;
  }
  
  showNoData() {
    const noDataHtml = '<div class="loading">No trading data available. Start by adding some trades!</div>';
    document.getElementById('totalPnl').parentElement.innerHTML = noDataHtml;
    document.getElementById('performanceSummary').innerHTML = noDataHtml;
    document.getElementById('recentActivity').innerHTML = noDataHtml;
  }
}

// Helper function to format dates
function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { 
    month: 'short', 
    day: 'numeric' 
  });
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  console.log('Initializing dashboard...');
  new Dashboard();
});

// Export for potential use in other scripts
window.Dashboard = Dashboard; 