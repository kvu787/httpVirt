# httpVirt

httpVirt is a tool for creating virtual machines and using them through a HTTP and WebSocket interface.

You can

* Deploy httpVirt to a physical machine
* Create resource-constrained VMs using physical machine resources
* Send shell commands to these VMs via HTTP and WebSocket requests

## Installation

1. Install [VirtualBox](https://www.virtualbox.org/wiki/Downloads).
2. Install [Vagrant 1.8.6](https://releases.hashicorp.com/vagrant/1.8.6) or [1.8.5](https://releases.hashicorp.com/vagrant/1.8.5).
3. Download the repo and run the httpVirt server:

```bash
$ git clone https://github.com/kvu787/httpVirt.git
```

## Start the httpVirt server

```bash
$ cd httpVirt
$ vagrant up
```

The latter command starts a server on http://127.0.0.1:10411 .
It takes a long time to run.

## Create a VM and run a command

### With an HTTP GET

Use cURL with the HTTP API:

```bash
$ id=`curl -w '\n' 'http://127.0.0.1:10411/create'`
8ebc3265a715ab0854fcb74c8305911e74c740471fa50f3a1b99949de43a9c0a
$ curl "http://127.0.0.1:10411/command/$id?command=ls%20-a"
.  ..  .bashrc  .profile
```

### With a WebSocket

Use client-side Javascript with the WebSocket API:

1. Visit http://127.0.0.1:10411 in your browser.
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
httpGetAsync('http://127.0.0.1:10411/create', function (containerID) {
  // open a WebSocket to the VM
  var webSocket = new WebSocket(`ws://127.0.0.1:10411/session/${containerID}`);

  // handler called for each incoming message
  webSocket.onmessage = function(e){
     var message = e.data;
     console.log('Message received: ' + message);
  }

  webSocket.onopen = function () {
    // when websocket is ready, send command as outgoing message
    webSocket.send('ls -a');
  }

  // onmessage handler will print output
});
```

4. You should see `Message received: <command output>` in the console.

## Shutdown httpVirt and clean up associated resources

```bash
$ cd httpVirt
$ vagrant destroy -f
```

## Run the xterm demo

Make sure httpVirt is running and open `demo/index.html` in your web browser.
