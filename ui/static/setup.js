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
        document.getElementById('footer-log').innerHTML = "------------<br/>";
    });

    document.getElementById('add-lines-button').addEventListener('click', () => {
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', "<br /><br /><br /><br /><br />");
    });
}

const wsSetup = () => {
    WSModule.createWSocket();
    const wSocket = WSModule.getWSocket();
    let tOut;

    wSocket.onmessage = (event) => {
        const evtJson = JSON.parse(event.data);
        let logMsg = '<span class="square"></span> ' + evtJson.message + '<br />';
        if(evtJson.type === 'mock') {
            logMsg = '<span class="square square-purple"></span> ' + evtJson.message + '<br />';
        } else if(evtJson.type === 'error') {
            logMsg = '<span class="square square-red"></span> ' + evtJson.message + '<br />';
        }

        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', logMsg);
    };

    wSocket.onopen = () => {
        document.getElementById("cover").style.display = "none";
        console.log("WebSocket Connected");
        document.getElementById('header-ws').innerHTML = 'connected <span class="bullet"></span>';

        WSModule.setWSAttempts(0);
    };

    wSocket.onerror = (error)  => {
        console.log("WebSocket Error: " + error);
        document.getElementById('header-ws').innerHTML = 'error <span class="bullet bullet-red"></span>';
    };

    wSocket.onclose = (event)  => {
        console.log("WebSocket Close: " + event.reason);
        if(event.reason === 'ws_takeover') {
            document.getElementById('header-ws').innerHTML = 'Taken over by other tab <span class="bullet bullet-red"></span>';
            clearTimeout(tOut);
            return;
        }

        document.getElementById('header-ws').innerHTML = 'reconnecting... <span class="bullet bullet-red"></span>';

        if (WSModule.getWSAttempts() < WSModule.getWSMaxAttempts()) {
            document.getElementById("cover").style.display = "block";
            tOut = setTimeout(reconnectWebSocket, WSModule.getWSReconnectDelay());
        } else {
            console.log("WebSocket reconnection failed after maximum attempts");
            document.getElementById('header-ws').innerHTML = 'closed <span class="bullet bullet-red"></span>';
        }
    };
}

const reconnectWebSocket = () => {
    WSModule.setWSAttempts(WSModule.getWSAttempts()+1);
    console.log(`Attempting to reconnect... (${WSModule.getWSAttempts()}/${WSModule.getWSMaxAttempts()})`);
    wsSetup();
}