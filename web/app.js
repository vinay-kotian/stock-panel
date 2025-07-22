document.getElementById('optionForm').addEventListener('submit', async function(e) {
  e.preventDefault();
  const form = e.target;
  const data = {
    symbol: form.symbol.value,
    underlying_symbol: form.underlying_symbol.value,
    option_type: form.option_type.value,
    strike_price: parseFloat(form.strike_price.value),
    expiry: form.expiry.value,
    price: parseFloat(form.price.value),
    side: form.side.value
  };
  const resultDiv = document.getElementById('result');
  resultDiv.textContent = 'Submitting...';
  try {
    const res = await fetch('/stocks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    });
    if (res.ok) {
      resultDiv.textContent = 'Submitted successfully!';
      form.reset();
    } else {
      const err = await res.text();
      resultDiv.textContent = 'Error: ' + err;
    }
  } catch (err) {
    resultDiv.textContent = 'Network error: ' + err;
  }
});
