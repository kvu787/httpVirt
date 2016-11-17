# httpVirt

httpVirt is a tool for creating virtual machines and using them through a HTTP and WebSocket interface.

You can

* deploy httpVirt to a physical machine
* create resource-constrained VMs using physical machine resources
* send shell commands to these VMs via HTTP and WebSocket requests

## Installation

1. Install VirtualBox.
2. Install Vagrant
3. Download the repo and run the httpVirt server:

```bash
$ git clone https://github.com/kvu787/httpVirt.git
```

## Start the httpVirt server

```bash
$ cd httpVirt
$ vagrant up
```

## Run a command

### With an HTTP GET

Use cURL with the HTTP API:

```bash
$ curl 'http://localhost:10411/create'
8ebc3265a715ab0854fcb74c8305911e74c740471fa50f3a1b99949de43a9c0a
$ curl 'http://localhost:10411/shell/8ebc3265a715ab0854fcb74c8305911e74c740471fa50f3a1b99949de43a9c0a?command=ls%20-a'
.  ..  .bashrc  .profile
```

### With a WebSocket

Use client-side Javascript with the WebSocket API:

1. Visit http://localhost:10411 in your browser.
2. Open the JavaScript console.
3. Copy/paste and run the following code:

```javascript

// Vanilla JS GET: http://stackoverflow.com/questions/247483/http-get-request-in-javascript
function httpGetAsync(theUrl, callback) {
    var xmlHttp = new XMLHttpRequest();
    xmlHttp.onreadystatechange = function() {
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
            callback(xmlHttp.responseText);
    }
    xmlHttp.open('GET', theUrl, true); // true for asynchronous
    xmlHttp.send(null);
}

// create a VM
httpGetAsync('http://localhost:10411/create', function (containerID) {
  // open a WebSocket to the VM
  var webSocket = new WebSocket(`ws://localhost:10411/session/${containerID}`);

  // handler called for each incoming message
  webSocket.onmessage = function(e){
     var message = e.data;
     console.log('Message received: ' + message);
  }

  // send command as outgoing message
  webSocket.onopen = function () {
    webSocket.send('ls -a');
  }

  // onmessage handler will print output
});
```
