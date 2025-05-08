const hotKeysSetup = () => {
    document.getElementById('toggleHotKeys').addEventListener('click', toggleHotKeys);

    document.addEventListener('keydown', (event) => {
        if (localStorage.getItem('moxyUseHotKeys') === 'false') {
            return;
        }

        //Fullscreen log
        if (event.ctrlKey && event.code === 'Space') {
            event.preventDefault();
            toggleFullcreen();
        }

        //Show only mocks
        if (event.ctrlKey && event.altKey && event.code === 'KeyS') {
            event.preventDefault();
            showOnlyMocks();
        }

        //Show compact list
        if (event.ctrlKey && event.altKey && event.code === 'KeyC') {
            event.preventDefault();
            const compactIsActive = document.getElementById('toggle_compactlist_bullet').classList.contains('action-btn-bullet-active');
            const compactProxyIsActive = document.getElementById('toggle_compactlist_proxy_bullet').classList.contains('action-btn-bullet-active');
            showCompactList(compactIsActive !== compactProxyIsActive);
            showCompactListProxy(compactIsActive !== compactProxyIsActive);
        }

        //Payload files
        if (event.ctrlKey && event.altKey && event.code === 'KeyF') {
            event.preventDefault();
            listPayloadFiles('payload_files_btn');
        }
    });

    if (localStorage.getItem('moxyUseHotKeys') === 'false') {
        document.getElementById('toggleHotKeys_bullet').classList.toggle('toggle-btn-bullet-active');
        document.getElementsByClassName('hotkeys-info')[0].style.display = 'none';
    }
}

const darkModeSetup = () => {
    document.getElementById('toggleDarkMode').addEventListener('click', toggleDarkMode);

    if (localStorage.getItem('moxyDarkMode') === 'false') {
        toggleDarkMode();
    }
}

const toggleHotKeys = () => {
    document.getElementById('toggleHotKeys_bullet').classList.toggle('toggle-btn-bullet-active');

    if (localStorage.getItem('moxyUseHotKeys') === 'true' || localStorage.getItem('moxyUseHotKeys') === null) {
        localStorage.setItem('moxyUseHotKeys', 'false');

        document.getElementsByClassName('hotkeys-info')[0].style.display = 'none';
    } else {
        localStorage.setItem('moxyUseHotKeys', 'true');

        document.getElementsByClassName('hotkeys-info')[0].style.display = '';
    }
}

const toggleDarkMode = () => {
    document.getElementById('toggleDarkMode_bullet').classList.toggle('toggle-btn-bullet-active');

    let darkModeLink = document.getElementById('darkModeStylesheet');
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

const toggleFullcreen = () => {
    if (document.querySelector('footer').classList.contains('expanded')) {
        document.getElementById("expand-button").innerHTML = '<i class="fa-solid fa-angle-up"></i> Fullscreen log <i class="fa-solid fa-angle-up"></i>';
    } else {
        document.getElementById("expand-button").innerHTML = '<i class="fa-solid fa-angle-down"></i> Fullscreen log <i class="fa-solid fa-angle-down"></i>';
    }

    document.querySelector('footer').classList.toggle('expanded');
}

const clickEvtSetup = () => {
    document.getElementById('expand-button').addEventListener('click', () => {
        toggleFullcreen();
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

    return curlString;
}

const showCurl = (el, evtData) => {
    const evtJson = JSON.parse(decodeURIComponent(evtData));
    document.getElementById("curl-dialog").style.display = "flex";

    console.log(getCurlString(evtJson, true));

    document.getElementById("curl-dialog-content").innerHTML = getCurlString(evtJson, false);
}

const copyCurl = () => {
    const curlContent = document.getElementById('curl-dialog-content')
        .innerText
        .replace(/\\\n--/g, '--');

    navigator.clipboard.writeText(curlContent)
        .then(() => {
            document.getElementById("copy-curl-status").innerHTML = "Copied!";
        })
        .catch(err => {
            document.getElementById("copy-curl-status").innerHTML = "ERROR failed to copy";
        });
}

const closeCurlDialog = () => {
    document.getElementById("copy-curl-status").innerHTML = "";
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
        } else if(evtJson.type === 'wsmock') {
            logMsg = '<span class="square square-yellow"></span> ' + evtJson.message + '<br />';
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