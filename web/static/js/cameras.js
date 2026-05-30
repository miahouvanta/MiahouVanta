// Camera Viewer - with modal player
(function(){
  var cameras = [];
  var grid = document.getElementById('camera-grid');
  var filter = document.getElementById('country-filter');

  // Create modal
  var modal = document.createElement('div');
  modal.id = 'camera-modal';
  modal.style.cssText = 'display:none;position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.92);z-index:1000;align-items:center;justify-content:center;flex-direction:column;';
  modal.innerHTML = '<div style="position:relative;max-width:900px;width:90%;">'
    + '<div style="display:flex;justify-content:space-between;align-items:center;padding:0.75rem 1rem;background:#1a1025;border-radius:8px 8px 0 0;border:1px solid #3b1f6e;border-bottom:none;">'
    + '<div><span id="modal-cam-name" style="color:#e2d5f5;font-size:0.85rem;font-weight:bold;"></span>'
    + '<span id="modal-cam-loc" style="color:#9b8ab8;font-size:0.7rem;margin-left:0.75rem;"></span></div>'
    + '<button onclick="closeCameraModal()" style="background:none;border:none;color:#a78bfa;font-size:1.5rem;cursor:pointer;">&times;</button>'
    + '</div>'
    + '<div style="background:#000;border-radius:0 0 8px 8px;overflow:hidden;border:1px solid #3b1f6e;border-top:none;">'
    + '<video id="modal-video" controls autoplay playsinline style="width:100%;max-height:500px;background:#000;"></video>'
    + '</div>'
    + '<div id="modal-error" style="display:none;padding:1rem;text-align:center;color:#f97316;font-size:0.75rem;"></div>'
    + '</div>'
    + '<p style="color:#9b8ab8;font-size:0.65rem;margin-top:0.75rem;">Press ESC or click outside to close</p>';
  document.body.appendChild(modal);

  // Close on background click
  modal.addEventListener('click', function(e){ if(e.target === modal) closeCameraModal(); });
  document.addEventListener('keydown', function(e){ if(e.key === 'Escape') closeCameraModal(); });

  // Expose to global scope for onclick handlers
  window.closeCameraModal = function(){
    modal.style.display = 'none';
    var vid = document.getElementById('modal-video');
    vid.pause();
    vid.removeAttribute('src');
    vid.load();
  };

  window.openCameraModal = function(cam){
    document.getElementById('modal-cam-name').textContent = cam.name;
    document.getElementById('modal-cam-loc').textContent = cam.city + ', ' + cam.country;
    var vid = document.getElementById('modal-video');
    var err = document.getElementById('modal-error');
    err.style.display = 'none';
    modal.style.display = 'flex';

    // Try HLS playback
    if (cam.type === 'hls' || cam.url.indexOf('.m3u8') !== -1) {
      if (window.Hls && Hls.isSupported()) {
        var hls = new Hls();
        hls.loadSource(cam.url);
        hls.attachMedia(vid);
        hls.on(Hls.Events.MANIFEST_PARSED, function(){ vid.play(); });
        hls.on(Hls.Events.ERROR, function(e, data){
          err.textContent = 'Stream unavailable (HLS error). Camera may be offline.';
          err.style.display = 'block';
        });
      } else if (vid.canPlayType('application/vnd.apple.mpegurl')) {
        vid.src = cam.url;
        vid.play();
      } else {
        err.textContent = 'HLS not supported. Try Chrome or Safari.';
        err.style.display = 'block';
      }
    } else {
      vid.src = cam.url;
      vid.play().catch(function(){
        err.textContent = 'Stream unavailable. Camera may be offline.';
        err.style.display = 'block';
      });
    }
  };

  function loadCameras(){
    fetch('/api/cameras/list')
      .then(function(r){ return r.json(); })
      .then(function(data){
        cameras = data;
        document.getElementById('camera-count').textContent = cameras.length + ' cameras';
        renderFilter();
        renderGrid();
      })
      .catch(function(e){ console.error('Failed to load cameras:', e); });
  }

  function renderFilter(){
    var countries = [];
    for(var i=0;i<cameras.length;i++){
      if(countries.indexOf(cameras[i].country)===-1) countries.push(cameras[i].country);
    }
    countries.sort();
    filter.innerHTML = '<option value="">All Countries</option>'
      + countries.map(function(c){ return '<option value="'+c+'">'+c+'</option>'; }).join('');
    filter.onchange = renderGrid;
  }

  function renderGrid(){
    var f = filter ? filter.value : '';
    var filtered = f ? cameras.filter(function(c){ return c.country===f; }) : cameras;
    if(filtered.length === 0){
      grid.innerHTML = '<div style="grid-column:1/-1;text-align:center;color:var(--text-dim);padding:2rem;">No cameras found</div>';
      return;
    }
    grid.innerHTML = filtered.map(function(c){
      return '<div class="camera-card" style="cursor:pointer;" data-cam-id="'+c.id+'">'
        + '<div style="height:160px;background:linear-gradient(135deg,#1a1025,#241540);display:flex;align-items:center;justify-content:center;position:relative;">'
        + '<span style="font-size:2.5rem;">&#x1F4F9;</span>'
        + '<div style="position:absolute;bottom:0;left:0;right:0;padding:0.5rem;background:linear-gradient(transparent,rgba(0,0,0,0.8));">'
        + '<div style="color:#e2d5f5;font-size:0.7rem;font-weight:bold;">'+c.name+'</div>'
        + '<div style="color:#9b8ab8;font-size:0.55rem;">'+c.city+', '+c.country+'</div>'
        + '</div></div></div>';
    }).join('');

    // Attach click handlers
    var cards = grid.querySelectorAll('[data-cam-id]');
    for(var i=0;i<cards.length;i++){
      (function(cam){
        cards[i].onclick = function(){ window.openCameraModal(cam); };
      })(filtered[i]);
    }
  }

  loadCameras();
})();
