(function(){
  let scanning = false;

  document.getElementById('scan-btn').onclick = function() {
    if(scanning) return;
    scanning = true;
    this.disabled = true;
    document.getElementById('scan-progress').style.display = 'block';
    document.getElementById('scan-status').textContent = 'Scanning...';
    document.getElementById('scan-text').textContent = 'Probing local network...';

    let pct = 0;
    const bar = document.getElementById('scan-bar');
    const interval = setInterval(() => { pct += Math.random()*3; if(pct>95) pct=95; bar.style.width=pct+'%'; }, 200);

    fetch('/api/scanner/scan', {method:'POST'}).then(r=>r.json()).then(() => {
      const poll = setInterval(() => {
        fetch('/api/scanner/results').then(r=>r.json()).then(d => {
          if(d.status === 'complete') {
            clearInterval(poll);
            clearInterval(interval);
            bar.style.width = '100%';
            scanning = false;
            document.getElementById('scan-btn').disabled = false;
            document.getElementById('scan-status').textContent = 'Complete';
            renderResults(d);
          }
        });
      }, 1000);
    });
  };

  function renderResults(data) {
    const results = data.results || [];
    let totalPorts = 0, services = new Set();
    results.forEach(h => { h.ports.forEach(p => { totalPorts++; services.add(p.service); }); });

    document.getElementById('vuln-hosts').textContent = results.length;
    document.getElementById('vuln-ports').textContent = totalPorts;
    document.getElementById('vuln-services').textContent = services.size;
    document.getElementById('vuln-scantime').textContent = data.time ? new Date(data.time).toLocaleTimeString() : 'Just now';

    const tbody = document.getElementById('host-table');
    tbody.innerHTML = results.map(h => '<tr>'
      + '<td>'+h.ip+'</td>'
      + '<td>'+(h.hostname||'—')+'</td>'
      + '<td>'+h.ports.map(p=>p.number).join(', ')+'</td>'
      + '<td>'+h.ports.map(p=>p.service).join(', ')+'</td>'
      + '</tr>').join('');
  }
})();
