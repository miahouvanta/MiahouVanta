(function(){
  const uploadArea = document.getElementById('upload-area');
  const fileInput = document.getElementById('file-input');

  uploadArea.onclick = () => fileInput.click();
  uploadArea.ondragover = e => { e.preventDefault(); uploadArea.style.borderColor = 'var(--primary)'; };
  uploadArea.ondragleave = () => { uploadArea.style.borderColor = 'var(--border)'; };
  uploadArea.ondrop = e => {
    e.preventDefault();
    uploadArea.style.borderColor = 'var(--border)';
    if(e.dataTransfer.files.length) uploadFile(e.dataTransfer.files[0]);
  };
  fileInput.onchange = e => { if(e.target.files.length) uploadFile(e.target.files[0]); };

  function uploadFile(file) {
    const formData = new FormData();
    formData.append('logfile', file);
    analyze(formData);
  }

  document.getElementById('analyze-btn').onclick = function() {
    const text = document.getElementById('log-paste').value.trim();
    if(!text) return;
    fetch('/api/logs/analyze', {
      method:'POST',
      headers:{'Content-Type':'application/json'},
      body: JSON.stringify({data:text})
    }).then(r=>r.json()).then(displayResults);
  };

  document.getElementById('clear-btn').onclick = function() {
    document.getElementById('log-paste').value = '';
    document.getElementById('analysis-results').style.display = 'none';
  };

  function analyze(formData) {
    fetch('/api/logs/upload', {method:'POST', body:formData})
      .then(r=>r.json()).then(displayResults);
  }

  function displayResults(data) {
    document.getElementById('analysis-results').style.display = 'block';
    const totalAttacks = Object.values(data.attack_types).reduce((a,b)=>a+b,0);
    document.getElementById('log-lines').textContent = data.total_lines;
    document.getElementById('log-attacks').textContent = totalAttacks;
    document.getElementById('log-severity').textContent = data.severity;
    document.getElementById('log-ips').textContent = Object.keys(data.source_ips).length;

    const tbody = document.querySelector('#attack-table tbody');
    tbody.innerHTML = Object.entries(data.attack_types).map(([t,c]) => {
      const sev = c > 20 ? 'critical' : c > 10 ? 'high' : c > 5 ? 'medium' : 'low';
      return '<tr><td>'+t.replace(/_/g,' ')+'</td><td>'+c+'</td><td><span class="badge badge-'+sev+'">'+sev+'</span></td></tr>';
    }).join('');

    document.getElementById('recommendations').innerHTML = data.recommendations.length
      ? data.recommendations.map(r => '<p>• '+r+'</p>').join('')
      : '<p style="color:var(--text-dim);">No specific recommendations.</p>';

    const ipsBody = document.querySelector('#ips-table tbody');
    ipsBody.innerHTML = Object.entries(data.source_ips).sort((a,b)=>b[1]-a[1]).slice(0,10).map(([ip,c]) =>
      '<tr><td><code>'+ip+'</code></td><td>'+c+'</td></tr>'
    ).join('');
  }
})();
