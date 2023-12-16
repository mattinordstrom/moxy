const darkModeSetup = () => {
    document.getElementById('toggleDarkMode').addEventListener('click', toggleDarkMode);

    if (localStorage.getItem('moxyDarkMode') === 'false') {
        toggleDarkMode();
    }
}

const toggleDarkMode = () => {
    var darkModeLink = document.getElementById('darkModeStylesheet');
    if (darkModeLink) {
        darkModeLink.remove();
        localStorage.setItem('moxyDarkMode', 'false');
    } else {
        darkModeLink = document.createElement('link');
        darkModeLink.id = 'darkModeStylesheet';
        darkModeLink.rel = 'stylesheet';
        darkModeLink.href = '/ui/static/style/style_dark.css';
        document.head.appendChild(darkModeLink);
        localStorage.setItem('moxyDarkMode', 'true');
    }
}

const clickEvtSetup = () => {
    document.getElementById('expand-button').addEventListener('click', () => {
        document.querySelector('footer').classList.toggle('expanded');
    });

    document.getElementById('clear-log-button').addEventListener('click', () => {
        document.getElementById('footer-log').innerHTML = "";
    });

    document.getElementById('add-lines-button').addEventListener('click', () => {
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', "<br /><br /><br /><br /><br />");
    });
}

const wsSetup = () => {
   console.log('blööö');
}



let socket;
let retryCount = 0;
let maxRetries = 3;
let pingInterval;
let pongTimeout;
const reconnectAfterCloseDelay = 4000;
const pongTimeoutDuration = 5000; // 5 seconds to wait for a pong response
const pingIntervalDuration = pongTimeoutDuration + 2000; // 7 seconds for each ping interval
const serverUrl = "ws://localhost:"+Number(location.port)+"/moxyws";
let currentUUID = uuidv4();

function connectWebSocket() {
    socket = new WebSocket(serverUrl+"?id="+currentUUID);

    socket.onopen = function(event) {
        console.log("WebSocket is open now.");
        startPinging();
    };

    socket.onmessage = function(event) {
        if (event.data === "pong") {
            console.log("Pong received");
            retryCount = 0; // Reset retry count on successful communication
            clearTimeout(pongTimeout); // Clear the pong timeout as pong is received
        }
    };

    socket.onclose = function(event) {
        console.log("WebSocket is closed now.");
        stopPinging(); // Stop pinging on close
        if (retryCount < maxRetries) {
            retryCount++;
            console.log(`Retry attempt ${retryCount}`);
            setTimeout(connectWebSocket, reconnectAfterCloseDelay);
        } else {
            console.log("Max retry attempts reached. Stopping.");
        }
    };

    socket.onerror = function(error) {
        console.log("WebSocket error: " + error);
        stopPinging(); // Stop pinging on error
    };
}

function startPinging() {
    sendPing();
    pingInterval = setInterval(sendPing, pingIntervalDuration);
}

function sendPing() {
    if (socket.readyState === WebSocket.OPEN) {
        socket.send("ping_" + currentUUID);
        setPongTimeout();
    } else {
        console.error("WebSocket not open. Unable to send ping.");
    }
}

function setPongTimeout() {
    clearTimeout(pongTimeout); // Clear any previous pong timeout
    pongTimeout = setTimeout(function() {
        console.error("No pong received within the expected timeframe. Stopping everything.");
        stopPinging();
        if (socket) socket.close();
    }, pongTimeoutDuration);
}

function stopPinging() {
    if (pingInterval) clearInterval(pingInterval);
    if (pongTimeout) clearTimeout(pongTimeout);
    console.log("Stopped ping process.");
}

connectWebSocket();
