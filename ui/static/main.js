let globalMockdefObj = {};
let globalProxydefObj = {};

let wSocket;
let wsAttempts = 0;
const wsMaxAttempts = 3;
const wsReconnectDelay = 5000;

function initFunc() {
    fetch('/moxyadminui/mockdef')
    .then(response => response.json())
    .then(data => {
        globalMockdefObj = data;
        renderMockdef(data);
    })
    .catch(error => { console.error('Fetch error mockdef:', error); });

    fetch('/moxyadminui/proxydef')
    .then(response => response.json())
    .then(data => {
        globalProxydefObj = data;
        renderProxydef(data);
    })
    .catch(error => { console.error('Fetch error proxydef:', error); });

    fetch('/moxyadminui/settings')
    .then(response => response.json())
    .then(data => {
        document.getElementById('header-port').innerHTML = data['port'];
        document.getElementById('header-route').innerHTML = data['defaultRoute'];
    })
    .catch(error => { console.error('Fetch error settings:', error); });

    //click events setup
    clickEvtSetup();

    //websocket setup
    wsSetup();
}

function renderMockdef() {
    const data = globalMockdefObj;
    for (let i = 0; i < data.length; i++) {
        const mockEntityData = data[i];
        let mockEntity = '<div class="proxymock-content"><div style="color:#999999; display: flex; justify-content: space-between;">'+(i+1) + 
            '<div style="display:flex">' +
            '<div style="margin-right:4px"><button onclick="moveMock(this)" id="movemock_up_btn_'+i+'">&#8593;</button></div>' +
            '<div style="margin-right:4px"><button onclick="moveMock(this)" id="movemock_down_btn_'+i+'">&#8595;</button></div>' +
            '<div><button onclick="removeMock(this)" id="x_btn_'+i+'">X</button></div>' +
            '</div></div>' +
            '<div class="mock-obj"><label for="active_mock_'+i+'">active:</label><input onclick="updateMockdef(this)" class="cbox" type="checkbox" name="active_mock_'+i+'" id="active_mock_'+i+'" ' + (mockEntityData['active'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="freezetimems_'+i+'">freezetimems:</label><input onchange="updateMockdef(this)" type="number" name="freezetimems_'+i+'" id="freezetimems_'+i+'" value="' + mockEntityData['freezetimems'] + '"></input></div>' +
            '<div class="mock-obj"><label for="method_'+i+'">method:</label><select onchange="updateMockdef(this)" class="slct" name="method_'+i+'" id="method_'+i+'">' +
            '<option value="GET" ' + (mockEntityData['method'] === "GET" ? "selected" : "") + '>GET</option>' +
            '<option value="POST" ' + (mockEntityData['method'] === "POST" ? "selected" : "") + '>POST</option>' +
            '<option value="PUT" ' + (mockEntityData['method'] === "PUT" ? "selected" : "") + '>PUT</option>' +
            '<option value="DELETE" ' + (mockEntityData['method'] === "DELETE" ? "selected" : "") + '>DELETE</option>' + 
            '</select></div>';

            let payload = mockEntityData['payload'];
            if (typeof payload === 'object') {
                payload = JSON.stringify(mockEntityData['payload'], null, 2);
            }

            mockEntity += '<div class="mock-obj"><label for="payload_'+i+'">payload:</label><textarea onchange="updateMockdef(this)" spellcheck="false" rows="8" cols="32" name="payload_'+i+'" id="payload_'+i+'">' + 
                payload + '</textarea></div>';

            mockEntity += '<div class="mock-obj"><label for="payloadFromFile_'+i+'">payloadFromFile:</label><textarea onchange="updateMockdef(this)" spellcheck="false" rows="4" cols="32" class="fixed-textarea" name="payloadFromFile_'+i+'" id="payloadFromFile_'+i+'">' + mockEntityData['payloadFromFile'] + '</textarea></div>';

            mockEntity += '<div class="mock-obj"><label for="statuscode_'+i+'">statuscode:</label><input onchange="updateMockdef(this)" type="number" name="statuscode_'+i+'" id="statuscode_'+i+'" value="' + mockEntityData['statuscode'] + '"></input></div>' +
            '<div class="mock-obj"><label for="urlpart_mock_'+i+'"><b>urlpart:</b></label><input onchange="updateMockdef(this)" spellcheck="false" class="input-wide" type="text" name="urlpart_mock_'+i+'" id="urlpart_mock_'+i+'" value="' + mockEntityData['urlpart'] + '"></input></div>' +
            '</div>';

            document.getElementById('mock-content-container').innerHTML += mockEntity + "<br />";
    };

}

function renderProxydef() {
    const data = globalProxydefObj;
    for (let i = 0; i < data.length; i++) {
        const proxyEntityData = data[i];
        let proxyEntity = '<div class="proxymock-content"><div style="color:#999999; display: flex; justify-content: space-between;">'+(i+1) + 
            '<div style="display:flex">' +
            '<div style="margin-right:4px"><button onclick="moveProxy(this)" id="moveproxy_up_btn_'+i+'">&#8593;</button></div>' +
            '<div style="margin-right:4px"><button onclick="moveProxy(this)" id="moveproxy_down_btn_'+i+'">&#8595;</button></div>' +
            '<div><button onclick="removeProxy(this)" id="x_btn_'+i+'">X</button></div>' +
            '</div></div>' +
            '<div class="proxy-obj"><label for="active_proxy_'+i+'">active:</label><input onclick="updateProxydef(this)" class="cbox" type="checkbox" name="active_proxy_'+i+'" id="active_proxy_'+i+'" ' + (proxyEntityData['active'] ? "checked" : "") + '></input></div>' +    
            '<div class="proxy-obj"><label for="target_'+i+'">target:</label><input onchange="updateProxydef(this)" spellcheck="false" class="input-wide" type="text" name="target_'+i+'" id="target_'+i+'" value="' + proxyEntityData['target'] + '"></input></div>' +
            '<div class="proxy-obj"><label for="urlpart_proxy_'+i+'"><b>urlpart:</b></label><input onchange="updateProxydef(this)" spellcheck="false" class="input-wide" type="text" name="urlpart_proxy_'+i+'" id="urlpart_proxy_'+i+'" value="' + proxyEntityData['urlpart'] + '"></input></div>' +
            '<div class="proxy-obj"><label for="verbose_'+i+'">verbose:</label><input onclick="updateProxydef(this)" class="cbox" type="checkbox" name="verbose_'+i+'" id="verbose_'+i+'" ' + (proxyEntityData['verbose'] ? "checked" : "") + '></input></div>' +    
            '</div>';

        document.getElementById('proxy-content-container').innerHTML += proxyEntity + "<br />";
    };

}

function clickEvtSetup() {
    document.getElementById('expand-button').addEventListener('click', function() {
        document.querySelector('footer').classList.toggle('expanded');
    });

    document.getElementById('clear-log-button').addEventListener('click', function() {
        document.getElementById('footer-log').innerHTML = "";
    });

    document.getElementById('add-lines-button').addEventListener('click', function() {
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', "<br /><br /><br /><br /><br />");
    });
}

function wsSetup() {
    wSocket = new WebSocket("ws://localhost:9097/ws");

    wSocket.onmessage = function(event) {
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', event.data + "<br />");
    };

    wSocket.onopen = function() {
        console.log("WebSocket Connected");
        document.getElementById('header-ws').innerHTML = 'connected <span class="bullet"></span>';

        wsAttempts = 0;
    };

    wSocket.onerror = function(error) {
        console.log("WebSocket Error: " + error);
        document.getElementById('header-ws').innerHTML = 'error <span class="bullet bullet-red"></span>';
    };

    wSocket.onclose = function(event) {
        console.log("WebSocket Close: " + event);
        document.getElementById('header-ws').innerHTML = 'reconnecting... <span class="bullet bullet-red"></span>';

        if (wsAttempts < wsMaxAttempts) {
            setTimeout(reconnectWebSocket, wsReconnectDelay);
        } else {
            console.log("WebSocket reconnection failed after maximum attempts");
            document.getElementById('header-ws').innerHTML = 'closed <span class="bullet bullet-red"></span>';
        }
    };
}

function reconnectWebSocket() {
    wsAttempts++;
    console.log(`Attempting to reconnect... (${wsAttempts}/${wsMaxAttempts})`);
    wsSetup();
}

function addMock() {
    const mock = {
        "active": true,
        "freezetimems": 0,
        "method": "GET",
        "payload": {
            "response": "abc123"
        },
        "payloadFromFile": "",
        "statuscode": 200,
        "urlpart": "/api/whatever/someendpoint"
    };

    globalMockdefObj.push(mock);

    resetAndSync("mock");
}

function addProxy() {
    const proxy = {
        "active": true,
        "target": "http://localhost:8080",
        "urlpart": "/api/test123",
        "verbose": false
    };

      globalProxydefObj.push(proxy);

      resetAndSync("proxy");
}

function removeMock(evt) {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    globalMockdefObj.splice(index, 1);

    resetAndSync("mock");
}

function removeProxy(evt) {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    globalProxydefObj.splice(index, 1);

    resetAndSync("proxy");
}

function updateMockdef(evt) {
    if(evt){
        const index = Number(evt.id.split('_').slice(-1)[0]);
        const name = evt.id.split('_')[0];

        if(evt.tagName.toLowerCase() === "input") {
            if(evt.type === "checkbox") {
                globalMockdefObj[index][name] = evt.checked;
            } else if(evt.type === "number") { 
                globalMockdefObj[index][name] = Number(evt.value);
            } else if(evt.type === "text") {
                globalMockdefObj[index][name] = evt.value;
            } 
        } else if(evt.tagName.toLowerCase() === "select") {
            globalMockdefObj[index][name] = evt.value;
        } else if(evt.tagName.toLowerCase() === "textarea") {
            if(name === "payload") {
                let value = evt.value;
                try {
                    value = JSON.parse(value);
                } catch (e) {
                    value = evt.value;
                }

                globalMockdefObj[index][name] = value;
            } else {
                globalMockdefObj[index][name] = evt.value;
            }
        }
    }

    fetch('/moxyadminui/mockdef', {
        method: "POST",
        body: JSON.stringify(globalMockdefObj),
    })
    .then(data => {
        console.log("Success POST mockdef");
    })
    .catch(error => {
        console.error('POST error mockdef:', error);
    });
}

function updateProxydef(evt) {
    if(evt){
        const index = Number(evt.id.split('_').slice(-1)[0]);
        const name = evt.id.split('_')[0];

        if(evt.tagName.toLowerCase() === "input") {
            if(evt.type === "checkbox") {
                globalProxydefObj[index][name] = evt.checked;
            } else if(evt.type === "text") {
                globalProxydefObj[index][name] = evt.value;
            } 
        }
    }

    fetch('/moxyadminui/proxydef', {
        method: "POST",
        body: JSON.stringify(globalProxydefObj),
    })
    .then(data => {
        console.log("Success POST proxydef");
    })
    .catch(error => {
        console.error('POST error proxydef:', error);
    });
}

function moveMock(evt) {
    if(moveEntity(evt, globalMockdefObj)) {
        resetAndSync("mock");
    }
}

function moveProxy(evt) {
    if(moveEntity(evt, globalProxydefObj)) {
        resetAndSync("proxy");
    }
}

function moveEntity(evt, arr) {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    const way = evt.id.split('_')[1];
    let changed = false;

    if(way === 'up') {
        if (index > 0 && index < arr.length) {
            [arr[index], arr[index - 1]] = [arr[index - 1], arr[index]];
            changed = true;
        }
    } else if (way === 'down') {
        if (index >= 0 && index < arr.length - 1) {
            [arr[index], arr[index + 1]] = [arr[index + 1], arr[index]];
            changed = true;
        }
    }

    return changed;
}

function resetAndSync(type) {
    if(type === "mock") {
        document.getElementById('mock-content-container').innerHTML = "";
        renderMockdef();
        updateMockdef();

        return;
    }

    document.getElementById('proxy-content-container').innerHTML = "";
    renderProxydef();
    updateProxydef();
}