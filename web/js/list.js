document.getElementById('loadBtn').addEventListener('click', async function() {
  const date = document.getElementById('datePicker').value;
  const resultDiv = document.getElementById('listResult');
  const tbody = document.querySelector('#stocksTable tbody');
  tbody.innerHTML = '';
  resultDiv.textContent = 'Loading...';
  const dailyPnlDiv = document.getElementById('dailyPnl');
  if (dailyPnlDiv) dailyPnlDiv.innerHTML = '';
  try {
    const token = localStorage.getItem('authToken');
    const headers = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const res = await fetch('/stocks', { headers });
    if (!res.ok) throw new Error(await res.text());
    const stocks = await res.json();
    // Filter by date (YYYY-MM-DD)
    const filtered = date ? stocks.filter(s => s.timestamp.startsWith(date)) : stocks;
    if (filtered.length === 0) {
      resultDiv.textContent = 'No trades found for this date.';
    } else {
      resultDiv.textContent = '';
      for (const s of filtered) {
        const row = document.createElement('tr');
        row.innerHTML = `
          <td data-label="Symbol">${s.symbol}</td>
          <td data-label="Underlying">${s.underlying_symbol}</td>
          <td data-label="Type">${s.option_type}</td>
          <td data-label="Strike">${s.strike_price}</td>
          <td data-label="Expiry">${s.expiry}</td>
          <td data-label="Price">${s.price}</td>
          <td data-label="Side">${s.side}</td>
          <td data-label="Timestamp">${s.timestamp.replace('T', ' ').slice(0, 19)}</td>
        `;
        tbody.appendChild(row);
      }
    }
    // Fetch daily P&L and display for selected date
    if (date) {
      const pnlRes = await fetch('/pnl', { headers });
      if (pnlRes.ok) {
        const pnls = await pnlRes.json();
        const pnlForDate = pnls.find(p => p.date === date);
        if (pnlForDate) {
          const pnlClass = pnlForDate.pnl >= 0 ? 'pnl-positive' : 'pnl-negative';
          const trendClass = pnlForDate.pnl >= 0 ? 'pnl-trend-up' : 'pnl-trend-down';
          const trendIcon = pnlForDate.pnl >= 0 ? '📈' : '📉';
          
          dailyPnlDiv.innerHTML = `
            <div class="daily-pnl-card">
              <div class="daily-pnl-label">Daily P&L for ${formatDate(date)}</div>
              <div class="daily-pnl-amount ${pnlClass}">₹${pnlForDate.pnl.toFixed(2)}</div>
              <div class="pnl-trend ${trendClass}">
                ${trendIcon} ${pnlForDate.pnl >= 0 ? 'Profit' : 'Loss'}
              </div>
            </div>
          `;
        } else {
          dailyPnlDiv.innerHTML = `
            <div class="daily-pnl-card">
              <div class="daily-pnl-label">Daily P&L for ${formatDate(date)}</div>
              <div class="daily-pnl-amount pnl-neutral">₹0.00</div>
              <div class="pnl-trend pnl-trend-neutral">📊 No Data</div>
            </div>
          `;
        }
      } else {
        dailyPnlDiv.innerHTML = `
          <div class="daily-pnl-card">
            <div class="daily-pnl-label">Daily P&L for ${formatDate(date)}</div>
            <div class="daily-pnl-amount pnl-neutral">Error</div>
            <div class="pnl-trend pnl-trend-neutral">⚠️ Error Loading</div>
          </div>
        `;
      }
    } else {
      dailyPnlDiv.innerHTML = '';
    }
  } catch (err) {
    resultDiv.textContent = 'Error: ' + err;
    if (dailyPnlDiv) dailyPnlDiv.innerHTML = '';
  }
}); 

// Add event listener for Show P&L button
document.getElementById('showPnlBtn').addEventListener('click', async function() {
  const pnlDisplay = document.getElementById('pnlDisplay');
  pnlDisplay.innerHTML = '<div style="text-align: center; padding: 2em;">Loading P&L data...</div>';
  try {
    const token = localStorage.getItem('authToken');
    const headers = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    const res = await fetch('/pnl', { headers });
    if (!res.ok) throw new Error(await res.text());
    const pnls = await res.json();
    if (pnls.length === 0) {
      pnlDisplay.innerHTML = `
        <div class="pnl-card">
          <h2>📊 P&L Summary</h2>
          <div class="pnl-amount pnl-neutral">No Data Available</div>
          <p>No P&L data has been recorded yet.</p>
        </div>
      `;
      return;
    }
    
    // Calculate summary statistics
    const totalPnl = pnls.reduce((sum, p) => sum + p.pnl, 0);
    const profitableDays = pnls.filter(p => p.pnl > 0).length;
    const losingDays = pnls.filter(p => p.pnl < 0).length;
    const breakEvenDays = pnls.filter(p => p.pnl === 0).length;
    const winRate = (profitableDays / pnls.length * 100).toFixed(1);
    const avgPnl = (totalPnl / pnls.length).toFixed(2);
    const bestDay = pnls.reduce((best, p) => p.pnl > best.pnl ? p : best);
    const worstDay = pnls.reduce((worst, p) => p.pnl < worst.pnl ? p : worst);
    
    const pnlClass = totalPnl >= 0 ? 'pnl-positive' : 'pnl-negative';
    const trendClass = totalPnl >= 0 ? 'pnl-trend-up' : 'pnl-trend-down';
    const trendIcon = totalPnl >= 0 ? '📈' : '📉';
    
    let html = `
      <div class="pnl-card">
        <h2>📊 Trading Performance Summary</h2>
        <div class="pnl-amount ${pnlClass}">₹${totalPnl.toFixed(2)}</div>
        <div class="pnl-trend ${trendClass}">
          ${trendIcon} ${totalPnl >= 0 ? 'Total Profit' : 'Total Loss'}
        </div>
        
        <div class="pnl-stats">
          <div class="pnl-stat">
            <div class="pnl-stat-label">📅 Total Days</div>
            <div class="pnl-stat-value">${pnls.length}</div>
          </div>
          <div class="pnl-stat">
            <div class="pnl-stat-label">✅ Profitable Days</div>
            <div class="pnl-stat-value pnl-positive">${profitableDays}</div>
          </div>
          <div class="pnl-stat">
            <div class="pnl-stat-label">❌ Losing Days</div>
            <div class="pnl-stat-value pnl-negative">${losingDays}</div>
          </div>
          <div class="pnl-stat">
            <div class="pnl-stat-label">📊 Win Rate</div>
            <div class="pnl-stat-value">${winRate}%</div>
          </div>
          <div class="pnl-stat">
            <div class="pnl-stat-label">📈 Avg Daily P&L</div>
            <div class="pnl-stat-value ${avgPnl >= 0 ? 'pnl-positive' : 'pnl-negative'}">₹${avgPnl}</div>
          </div>
          <div class="pnl-stat">
            <div class="pnl-stat-label">🎯 Break Even</div>
            <div class="pnl-stat-value pnl-neutral">${breakEvenDays}</div>
          </div>
        </div>
      </div>
      
      <div class="pnl-summary">
        <div class="pnl-summary-info">
          <div class="pnl-summary-label">🏆 Best Day</div>
          <div class="pnl-summary-value pnl-positive">${formatDate(bestDay.date)}: ₹${bestDay.pnl.toFixed(2)}</div>
        </div>
      </div>
      
      <div class="pnl-summary">
        <div class="pnl-summary-info">
          <div class="pnl-summary-label">📉 Worst Day</div>
          <div class="pnl-summary-value pnl-negative">${formatDate(worstDay.date)}: ₹${worstDay.pnl.toFixed(2)}</div>
        </div>
      </div>
      
      <div class="pnl-table">
        <table>
          <thead>
            <tr>
              <th>📅 Date</th>
              <th>💰 P&L</th>
              <th>📊 Status</th>
            </tr>
          </thead>
          <tbody>
    `;
    
    // Sort by date (newest first)
    const sortedPnls = pnls.sort((a, b) => new Date(b.date) - new Date(a.date));
    
    for (const p of sortedPnls) {
      const pnlClass = p.pnl >= 0 ? 'pnl-positive' : 'pnl-negative';
      const statusIcon = p.pnl > 0 ? '✅' : p.pnl < 0 ? '❌' : '➖';
      const statusText = p.pnl > 0 ? 'Profit' : p.pnl < 0 ? 'Loss' : 'Break Even';
      
      html += `
        <tr>
          <td>${formatDate(p.date)}</td>
          <td class="${pnlClass}">₹${p.pnl.toFixed(2)}</td>
          <td>${statusIcon} ${statusText}</td>
        </tr>
      `;
    }
    
    html += `
          </tbody>
        </table>
      </div>
    `;
    
    pnlDisplay.innerHTML = html;
  } catch (err) {
    pnlDisplay.innerHTML = `
      <div class="pnl-card">
        <h2>⚠️ Error</h2>
        <div class="pnl-amount pnl-negative">Failed to Load</div>
        <p>Error loading P&L data: ${err}</p>
      </div>
    `;
  }
});

// Helper function to format dates nicely
function formatDate(dateString) {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', { 
    weekday: 'short', 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  });
} 