fetch('/api/data')
    .then(r => r.json())
    .then(data => document.getElementById('api-response').textContent = JSON.stringify(data, null, 2))
    .catch(e => document.getElementById('api-response').textContent = 'Error: ' + e.message);
