// CyberDeck Chat - Miahou AI
(function() {
  var messagesDiv = document.getElementById('chat-messages');
  var input = document.getElementById('chat-input');
  var sendBtn = document.getElementById('chat-send');
  if (!messagesDiv || !input || !sendBtn) { return; }

  function appendMessage(role, content) {
    var time = new Date().toLocaleTimeString();
    var div = document.createElement('div');
    div.className = 'chat-msg ' + role;
    var formatted = (content || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/\n/g,'<br>');
    div.innerHTML = '<div class="chat-bubble">' + formatted + '</div><span class="chat-time">' + time + '</span>';
    messagesDiv.appendChild(div);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
  }

  function sendMessage() {
    var text = input.value.trim();
    if (!text) return;
    input.value = '';
    appendMessage('user', text);
    var typing = document.createElement('div');
    typing.id = 'typing-indicator';
    typing.className = 'chat-msg miahou';
    typing.innerHTML = '<div class="chat-bubble" style="color:#9b8ab8;font-style:italic;">Miahou is typing...</div>';
    messagesDiv.appendChild(typing);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;

    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/chat/send', true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.timeout = 15000;
    xhr.onreadystatechange = function() {
      if (xhr.readyState === 4) {
        var el = document.getElementById('typing-indicator');
        if (el) el.remove();
        if (xhr.status === 200) {
          try { var d = JSON.parse(xhr.responseText); appendMessage('miahou', d.content || 'Hmm, something went wrong.'); }
          catch(e) { appendMessage('miahou', 'Sorry, could not parse response.'); }
        } else { appendMessage('miahou', 'Connection error. Is CyberDeck running?'); }
      }
    };
    xhr.onerror = function() { var el = document.getElementById('typing-indicator'); if (el) el.remove(); appendMessage('miahou', 'Network error.'); };
    xhr.ontimeout = function() { var el = document.getElementById('typing-indicator'); if (el) el.remove(); appendMessage('miahou', 'Timeout. Try again.'); };
    xhr.send(JSON.stringify({message: text}));
  }

  sendBtn.addEventListener('click', sendMessage);
  input.addEventListener('keydown', function(e) { if (e.key === 'Enter') sendMessage(); });
})();
