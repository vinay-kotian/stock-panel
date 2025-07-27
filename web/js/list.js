document.getElementById('loadBtn').addEventListener('click', async function() {
  const date = document.getElementById('datePicker').value;
  const resultDiv = document.getElementById('listResult');
  const tbody = document.querySelector('#stocksTable tbody');
  tbody.innerHTML = '';
  resultDiv.textContent = 'Loading...';
  try {
    const res = await fetch('/stocks');
    if (!res.ok) throw new Error(await res.text());
    const stocks = await res.json();
    // Filter by date (YYYY-MM-DD)
    const filtered = date ? stocks.filter(s => s.timestamp.startsWith(date)) : stocks;
    if (filtered.length === 0) {
      resultDiv.textContent = 'No trades found for this date.';
      return;
    }
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
  } catch (err) {
    resultDiv.textContent = 'Error: ' + err;
  }
}); 