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

const getCurlString = (evtJson, consolelog) => {
    const NEWLINE = consolelog ? "\n" : "<br/>";
    /*
    const curlString = 
        "curl --request PUT --location 'localhost:9097/api/test' \\"  + NEWLINE +
        "--header 'content-type: application/json' \\"  + NEWLINE +
        "--data '{"  + NEWLINE +
        "    \"matti123\": 5555666777" + NEWLINE +
        "}'";
    */

    const targetUrl = evtJson.message.substring(evtJson.message.lastIndexOf(' http') + 1)
    let curlString = "curl --request " + evtJson.extras.httpMethod + " --location '" + targetUrl + "' \\" + NEWLINE;

    let headers = evtJson.extras.headers.split(" | ");
    for(let i = 0; i < headers.length; i++) {
        if(!consolelog) {
            if(headers[i].toLowerCase().startsWith("content-length") || headers[i].toLowerCase().startsWith("accept-encoding")) {
                continue;
            }
        }

        curlString += "--header '" + headers[i] + "' \\" + NEWLINE;
    }

    curlString += "--data '" + evtJson.extras.body.replace(/'/g, "'\\''") + "'";

    if(!consolelog) {
        curlString += "<br/><br/><hr/><b>Some headers might not be included in the curl command above. See console for all</b><br/><br/>";
    }

    return curlString;
}

const showCurl = (el, evtData) => {
    const evtJson = JSON.parse(decodeURIComponent(evtData));
    document.getElementById("curl-dialog").style.display = "flex";

    console.log(getCurlString(evtJson, true));

    document.getElementById("curl-dialog-content").innerHTML = getCurlString(evtJson, false);
}

const closeCurlDialog = () => {
    document.getElementById("curl-dialog").style.display = "none";
}

const wsSetup = () => {
    WSModule.createWSocket();
    const wSocket = WSModule.getWSocket();
    let tOut;

    wSocket.onmessage = (event) => {
        const evtJson = JSON.parse(event.data);
        let logMsg = '';
        if(evtJson.type === 'mock' || evtJson.type === 'proxy') {
            let output = evtJson.message;
            if(evtJson.type === 'mock') {
                logMsg = '<span class="square square-purple"></span> ' + output + '<br />';
            } else {
                if(evtJson.extras && evtJson.extras.httpMethod !== "") {
                    output = "<br/>"+evtJson.extras.httpMethod + " " + evtJson.message +
                        "<br/><i>"+evtJson.extras.headers+"</i><br/>"+evtJson.extras.body+"<br/>";

                        const dataString = encodeURIComponent(event.data).replace(/'/g, '%27');
                    logMsg = `<span style="cursor:pointer" class="square" onclick="showCurl(this, '${dataString}')"></span> ${output}<br />`;
                } else {
                    logMsg = '<span class="square"></span> ' + output + '<br />'; 
                }
            }
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