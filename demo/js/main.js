function main() {
  var cols = 80;
  var rows = 24;
  var term = new Terminal({ cursorBlink: true });
  var terminalContainer = document.getElementById('terminal-container');

  // associate JS term object with HTML div
  term.open(terminalContainer);
  var charWidth = Math.ceil(term.element.offsetWidth / cols);
  var charHeight = Math.ceil(term.element.offsetHeight / rows);

  // create a backend container
  $.get('http://127.0.0.1:10411/create', function(containerID, status) {
    // create a websocket to that contianer
    var socketURL = `ws://127.0.0.1:10411/xterm/${containerID}`;
    var webSocket = new WebSocket(socketURL);
    webSocket.onopen = function() {
      term.attach(webSocket);
      console.log('Terminal attached');
    };
    webSocket.onerror = function(event) {
      console.log('Socket error: ' + event.data);
    };
    webSocket.onclose = function() {
      console.log('Socket closed');
    };
  });
}

main();
