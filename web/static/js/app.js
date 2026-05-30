// CyberDeck UI — Shared JS
document.addEventListener('DOMContentLoaded', function() {
  // Clock
  function updateClock() {
    var el = document.getElementById('clock');
    if (el) el.textContent = new Date().toLocaleString();
  }
  updateClock();
  setInterval(updateClock, 1000);
});
