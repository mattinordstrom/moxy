const initFunc = () => {
    fetch('/moxyadminui/mockdef')
    .then(response => response.json())
    .then(data => {
        MockDefModule.set(data);
        renderMockdef(data);
    })
    .catch(error => { console.error('Fetch error mockdef:', error); });

    fetch('/moxyadminui/proxydef')
    .then(response => response.json())
    .then(data => {
        ProxyDefModule.set(data);
        renderProxydef(data);
    })
    .catch(error => { console.error('Fetch error proxydef:', error); });

    fetch('/moxyadminui/settings')
    .then(response => {
        if (response.ok) {
            return response.json();
        } else {
            return response.text().then(text => {
                const logMsg = '<span class="square square-red"></span> ' + text + '<br />';
                document.getElementById("footer-log").insertAdjacentHTML('afterbegin', logMsg);

                throw new Error('Error ' + response.status + ': ' + text);
            });
        }
    })
    .then(data => {
        document.getElementById('header-port').innerHTML = data['port'];
        document.getElementById('header-route').innerHTML = data['defaultRoute'];

        PayloadFromFileModule.setPayloadPath(data['payloadPath']);
        PayloadFromFileModule.setPayloadFiles(data['payloadFiles']);
    })
    .catch(error => { 
        console.error('Fetch error settings:', error);
     });

    darkModeSetup();

    //websocket setup
    wsSetup();

    //click events setup
    clickEvtSetup();
}

const renderMockdef = () => {
    const data = MockDefModule.get();
    for (let i = 0; i < data.length; i++) {
        const mockEntityData = data[i];
        let mockEntity = '<div class="proxymock-content"><div style="color:#999999; display: flex; justify-content: space-between;">'+(i+1) + 
            '<div style="display:flex">' +
            '<div><input onchange="updateMockdef(this)" class="comment-input" type="text" spellcheck="false" name="comment_mock_'+i+'" id="comment_mock_'+i+'" value="' + (mockEntityData['comment'] || "") + '"></input></div>' +
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

const renderProxydef = () => {
    const data = ProxyDefModule.get();
    for (let i = 0; i < data.length; i++) {
        const proxyEntityData = data[i];
        let proxyEntity = '<div class="proxymock-content"><div style="color:#999999; display: flex; justify-content: space-between;">'+(i+1) + 
            '<div style="display:flex">' +
            '<div><input onchange="updateProxydef(this)" class="comment-input" type="text" spellcheck="false" name="comment_proxy_'+i+'" id="comment_proxy_'+i+'" value="' + (proxyEntityData['comment'] || "") + '"></input></div>' +
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

const toggleDarkMode = () => {
    var darkModeLink = document.getElementById('darkModeStylesheet');
    if (darkModeLink) {
        darkModeLink.remove();
        localStorage.setItem('moxyDarkMode', 'false');
    } else {
        darkModeLink = document.createElement('link');
        darkModeLink.id = 'darkModeStylesheet';
        darkModeLink.rel = 'stylesheet';
        darkModeLink.href = '/ui/static/style_dark.css';
        document.head.appendChild(darkModeLink);
        localStorage.setItem('moxyDarkMode', 'true');
    }
}

const darkModeSetup = () => {
    document.getElementById('toggleDarkMode').addEventListener('click', toggleDarkMode);

    if (localStorage.getItem('moxyDarkMode') === 'false') {
        toggleDarkMode();
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
    WSModule.createWSocket();
    const wSocket = WSModule.getWSocket();

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
        console.log("WebSocket Connected");
        document.getElementById('header-ws').innerHTML = 'connected <span class="bullet"></span>';

        WSModule.setWSAttempts(0);
    };

    wSocket.onerror = (error)  => {
        console.log("WebSocket Error: " + error);
        document.getElementById('header-ws').innerHTML = 'error <span class="bullet bullet-red"></span>';
    };

    wSocket.onclose = (event)  => {
        console.log("WebSocket Close: " + event);
        document.getElementById('header-ws').innerHTML = 'reconnecting... <span class="bullet bullet-red"></span>';

        if (WSModule.getWSAttempts() < WSModule.getWSMaxAttempts()) {
            setTimeout(reconnectWebSocket, WSModule.getWSReconnectDelay());
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

const maximizeFirst = () => {
    document.getElementById('payload_0').style.height = '470px';
    document.getElementById('payload_0').style.width = '970px';
}

const listPayloadFiles = () => {
    document.getElementById('payloadFiles').style.display = 'block';
    document.getElementById('payloadFilesContent').innerHTML = '<br />' + PayloadFromFileModule.getPayloadPath() + '<hr /><br />';

    let filesListHtml = '';
    const files = PayloadFromFileModule.getPayloadFiles();
    for (let i = 0; i < files.length; i++) {
        filesListHtml += '<div style="display:flex"><div style="min-width:225px">' + files[i] + '</div>';
        filesListHtml += '&nbsp;&nbsp;<button onclick="navigator.clipboard.writeText(\'' + PayloadFromFileModule.getPayloadPath() + files[i] + '\')">Copy full path</button>';
        filesListHtml += '</div><br />';
    }

    document.getElementById('payloadFilesContent').innerHTML += filesListHtml;
}

const closeListPayloadFiles = () => {
    document.getElementById('payloadFiles').style.display = 'none';
}

const addMock = () => {
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

    MockDefModule.set([...MockDefModule.get(), mock]);

    resetAndSync("mock");
}

const addProxy = () => {
    const proxy = {
        "active": true,
        "target": "http://localhost:8080",
        "urlpart": "/api/test123",
        "verbose": false
    };

    ProxyDefModule.set([...ProxyDefModule.get(), proxy]);

    resetAndSync("proxy");
}

const removeMock = (evt) => {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    MockDefModule.remove(index);

    resetAndSync("mock");
}

const removeProxy = (evt) => {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    ProxyDefModule.remove(index);

    resetAndSync("proxy");
}

const updateMockdef = (evt) => {
    if(evt){
        const index = Number(evt.id.split('_').slice(-1)[0]);
        const name = evt.id.split('_')[0];

        let mocks = MockDefModule.get();

        if(evt.tagName.toLowerCase() === "input") {
            if(evt.type === "checkbox") {
                mocks[index][name] = evt.checked;
            } else if(evt.type === "number") { 
                mocks[index][name] = Number(evt.value);
            } else if(evt.type === "text") {
                mocks[index][name] = evt.value;
            } 
        } else if(evt.tagName.toLowerCase() === "select") {
            mocks[index][name] = evt.value;
        } else if(evt.tagName.toLowerCase() === "textarea") {
            if(name === "payload") {
                let value = evt.value;
                try {
                    value = JSON.parse(value);
                } catch (e) {
                    value = evt.value;
                }

                mocks[index][name] = value;
            } else if(name === "payloadFromFile" && evt.value.startsWith('~')) {
                alert('Absolute path cannot start with ~');
                return;
            } else {
                mocks[index][name] = evt.value;
            }
        }

        MockDefModule.set(mocks);
    }

    fetch('/moxyadminui/mockdef', {
        method: "POST",
        body: JSON.stringify(MockDefModule.get()),
    })
    .then(data => {
        console.log("Success POST mockdef");
    })
    .catch(error => {
        console.error('POST error mockdef:', error);
    });
}

const updateProxydef = (evt) => {
    if(evt){
        const index = Number(evt.id.split('_').slice(-1)[0]);
        const name = evt.id.split('_')[0];

        let proxies = ProxyDefModule.get();

        if(evt.tagName.toLowerCase() === "input") {
            if(evt.type === "checkbox") {
                proxies[index][name] = evt.checked;
            } else if(evt.type === "text") {
                proxies[index][name] = evt.value;
            } 
        }

        ProxyDefModule.set(proxies);
    }

    fetch('/moxyadminui/proxydef', {
        method: "POST",
        body: JSON.stringify(ProxyDefModule.get()),
    })
    .then(data => {
        console.log("Success POST proxydef");
    })
    .catch(error => {
        console.error('POST error proxydef:', error);
    });
}

const moveMock = (evt) => {
    let mocks = MockDefModule.get();
    if(moveEntity(evt, mocks)) {
        MockDefModule.set(mocks);
        resetAndSync("mock");
    }
}

const moveProxy = (evt) => {
    let proxies = ProxyDefModule.get();
    if(moveEntity(evt, proxies)) {
        ProxyDefModule.set(proxies);
        resetAndSync("proxy");
    }
}

const moveEntity = (evt, arr) => {
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

const resetAndSync = (type) => {
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