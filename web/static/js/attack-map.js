// Attack Map — Real-time with detailed attack info
(function(){
  const map = L.map('map').setView([20, 0], 2);
  L.tileLayer('https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png', {
    attribution: '&copy; CartoDB', maxZoom: 19
  }).addTo(map);

  const sevColors = {critical:'#ef4444',high:'#f97316',medium:'#eab308',low:'#22c55e'};
  let totalAttacks = 0, critCount = 0, typeCounts = {}, targetCountries = {};
  let minuteCount = 0;
  const MAX_MARKERS = 300;
  const markers = [];

  const attackInfo = {
    'DDoS':       {desc:'Distributed Denial of Service — flooding target with traffic',icon:'💥'},
    'Phishing':   {desc:'Social engineering — tricking users into revealing credentials',icon:'🎣'},
    'Ransomware': {desc:'Malware encrypting files, demanding payment',icon:'🔒'},
    'Malware':    {desc:'Malicious software targeting systems',icon:'🦠'},
    'Exploit':    {desc:'Exploiting known vulnerabilities in software/services',icon:'⚡'},
    'BruteForce': {desc:'Credential stuffing / brute force login attempts',icon:'🔨'},
    'SQLi':       {desc:'SQL Injection — manipulating database queries',icon:'💉'},
    'XSS':        {desc:'Cross-Site Scripting — injecting malicious scripts',icon:'🕸️'},
    'Recon':      {desc:'Reconnaissance — port scanning, enumeration',icon:'🔍'},
    'Privilege':  {desc:'Privilege escalation attempts',icon:'⬆️'},
  };

  function addAttack(a) {
    totalAttacks++; minuteCount++;
    if(a.severity==='critical') critCount++;
    typeCounts[a.type] = (typeCounts[a.type]||0)+1;
    targetCountries[a.country] = (targetCountries[a.country]||0)+1;
    const color = sevColors[a.severity]||'#7c3aed';
    const info = attackInfo[a.type]||{desc:a.type,icon:'⚠️'};

    const src = L.circleMarker([a.src_lat,a.src_lng], {radius:4,fillColor:color,color:color,opacity:0.9,fillOpacity:0.7,weight:1});
    const tgt = L.circleMarker([a.tgt_lat,a.tgt_lng], {radius:7,fillColor:color,color:'#fff',opacity:1,fillOpacity:0.9,weight:2});
    const line = L.polyline([[a.src_lat,a.src_lng],[a.tgt_lat,a.tgt_lng]], {color:color,opacity:0.5,weight:1.5,dashArray:'6,4'});

    const popup = '<div style="font-family:monospace;min-width:240px;">'
      + '<div style="display:flex;align-items:center;gap:6px;margin-bottom:8px;">'
      + '<span style="font-size:18px;">' + info.icon + '</span>'
      + '<b style="color:' + color + ';font-size:13px;">' + a.type + '</b>'
      + '<span style="background:' + color + '22;color:' + color + ';border:1px solid ' + color + '44;padding:1px 6px;border-radius:3px;font-size:10px;margin-left:auto;">' + a.severity + '</span>'
      + '</div>'
      + '<p style="font-size:11px;color:#aaa;margin-bottom:8px;">' + info.desc + '</p>'
      + '<div style="display:grid;grid-template-columns:1fr 1fr;gap:4px;font-size:11px;color:#ccc;">'
      + '<div><span style="color:#888;">FROM:</span> <b>' + (a.src_city||'?') + ', ' + (a.src_country||'??') + '</b></div>'
      + '<div><span style="color:#888;">TO:</span> <b>' + a.city + ', ' + a.country + '</b></div>'
      + '<div><span style="color:#888;">SRC IP:</span> <code>' + (a.src_ip||'N/A') + '</code></div>'
      + '<div><span style="color:#888;">TGT IP:</span> <code>' + (a.tgt_ip||'N/A') + '</code></div>'
      + '<div><span style="color:#888;">PORT:</span> <code>' + (a.port||'N/A') + '</code></div>'
      + '<div><span style="color:#888;">PROTO:</span> <code>' + (a.protocol||'TCP') + '</code></div>'
      + '<div style="grid-column:span 2;"><span style="color:#888;">TARGET SERVICE:</span> ' + (a.target_service||'Unknown') + '</div>'
      + '<div style="grid-column:span 2;"><span style="color:#888;">TIME:</span> ' + new Date(a.timestamp).toLocaleTimeString() + '</div>'
      + '<div style="grid-column:span 2;"><span style="color:#888;">ATTACK ID:</span> <code>' + a.id + '</code></div>'
      + '</div></div>';

    tgt.bindPopup(popup);
    line.bindPopup(popup);
    tgt.addTo(map); line.addTo(map); src.addTo(map);
    markers.push(src, tgt, line);

    addLogRow(a, info, color);
    updateStats();
    updateTopLists();

    while(markers.length > MAX_MARKERS*3) { map.removeLayer(markers.shift()); }
    setTimeout(() => { try{map.removeLayer(src);map.removeLayer(line);}catch(e){} }, 8000);
  }

  function addLogRow(a, info, color) {
    const log = document.getElementById('attack-log');
    if(!log) return;
    const row = document.createElement('div');
    row.className = 'log-row';
    row.innerHTML = '<span class="log-sev" style="background:'+color+'22;color:'+color+';">'+a.severity+'</span>'
      + '<span class="log-type">'+info.icon+' '+a.type+'</span>'
      + '<span class="log-target">'+a.city+', '+a.country+'</span>'
      + '<span class="log-time">'+new Date(a.timestamp).toLocaleTimeString()+'</span>';
    row.onclick = () => { map.flyTo([a.tgt_lat,a.tgt_lng],6); };
    log.insertBefore(row, log.firstChild);
    while(log.children.length > 80) log.removeChild(log.lastChild);
  }

  function updateStats() {
    const el = id => document.getElementById(id);
    if(el('attack-count')) el('attack-count').textContent = totalAttacks;
    if(el('stat-live')) el('stat-live').textContent = minuteCount;
    if(el('stat-critical')) el('stat-critical').textContent = critCount;
    if(el('stat-types')) el('stat-types').textContent = Object.keys(typeCounts).length;
    if(el('stat-targets')) el('stat-targets').textContent = Object.keys(targetCountries).length;
  }

  function updateTopLists() {
    const sorted = Object.entries(typeCounts).sort((a,b)=>b[1]-a[1]).slice(0,5);
    const el = document.getElementById('top-attacks');
    if(el) el.innerHTML = sorted.map(([t,c]) => '<div class="mini-stat"><span>'+(attackInfo[t]?attackInfo[t].icon:'⚠️')+' '+t+'</span><b>'+c+'</b></div>').join('');
    const sorted2 = Object.entries(targetCountries).sort((a,b)=>b[1]-a[1]).slice(0,5);
    const el2 = document.getElementById('top-targets');
    if(el2) el2.innerHTML = sorted2.map(([c,n]) => '<div class="mini-stat"><span>🏴 '+c+'</span><b>'+n+'</b></div>').join('');
  }

  setInterval(() => {
    minuteCount = 0;
  }, 60000);

  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const ws = new WebSocket(proto + '//' + window.location.host + '/api/attack-map/stream');
  ws.onmessage = e => { const d=JSON.parse(e.data); addAttack(d); };
})();
