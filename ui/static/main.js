let globalMockdefObj = {};
let globalProxydefObj = {};

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

function renderMockdef(data) {
    for (let i = 0; i < data.length; i++) {
        const mockEntityData = data[i];
        let mockEntity = '<div class="proxymock-content">' +
            '<div class="mock-obj"><label for="active_mock_'+i+'">active:</label><input onclick="updateMockdef(this)" class="cbox" type="checkbox" name="active_mock_'+i+'" id="active_mock_'+i+'" ' + (mockEntityData['active'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="freezetime_'+i+'">freezetimems:</label><input onchange="updateMockdef(this)" type="text" name="freezetimems_'+i+'" id="freezetimems_'+i+'" value="' + mockEntityData['freezetimems'] + '"></input></div>' +
            '<div class="mock-obj"><label for="matchpathexact_'+i+'">matchPathExact:</label><input onclick="updateMockdef(this)" class="cbox" type="checkbox" name="matchPathExact_'+i+'" id="matchPathExact_'+i+'" ' + (mockEntityData['matchPathExact'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="method_'+i+'">method:</label><select onchange="updateMockdef(this)" class="slct" name="method_'+i+'" id="method_'+i+'">' +
            '<option value="GET" ' + (mockEntityData['method'] === "GET" ? "selected" : "") + '>GET</option>' +
            '<option value="POST" ' + (mockEntityData['method'] === "POST" ? "selected" : "") + '>POST</option>' +
            '<option value="PUT" ' + (mockEntityData['method'] === "PUT" ? "selected" : "") + '>PUT</option>' +
            '<option value="DELETE" ' + (mockEntityData['method'] === "DELETE" ? "selected" : "") + '>DELETE</option>' + 
            '</select></div>';

            if(mockEntityData['payload']) {
                mockEntity += '<div class="mock-obj"><label for="payload_'+i+'">payload:</label><textarea onchange="updateMockdef(this)" spellcheck="false" rows="8" cols="32" name="payload_'+i+'" id="payload_'+i+'">' + JSON.stringify(mockEntityData['payload'], null, 2) + '</textarea></div>';
            } else if (mockEntityData['payloadFromFile']) {
                mockEntity += '<div class="mock-obj"><label for="payloadFromFile_'+i+'">payloadFromFile:</label><textarea onchange="updateMockdef(this)" spellcheck="false" rows="4" cols="32" class="fixed-textarea" name="payloadFromFile_'+i+'" id="payloadFromFile_'+i+'">' + mockEntityData['payloadFromFile'] + '</textarea></div>';
            }

            mockEntity += '<div class="mock-obj"><label for="statuscode_'+i+'">statuscode:</label><input onchange="updateMockdef(this)" type="text" name="statuscode_'+i+'" id="statuscode_'+i+'" value="' + mockEntityData['statuscode'] + '"></input></div>' +
            '<div class="mock-obj"><label for="urlpart_mock_'+i+'"><b>urlpart:</b></label><input onchange="updateMockdef(this)" spellcheck="false" class="input-wide" type="text" name="urlpart_mock_'+i+'" id="urlpart_mock_'+i+'" value="' + mockEntityData['urlpart'] + '"></input></div>' +
            '</div>';

            document.getElementById('mock-content-container').innerHTML += mockEntity + "<br />";
    };

}

function renderProxydef(data) {
    for (let i = 0; i < data.length; i++) {
        const proxyEntityData = data[i];
        let proxyEntity = '<div class="proxymock-content">' +
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
    var socket = new WebSocket("ws://localhost:9097/ws");

    socket.onmessage = function(event) {
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', event.data + "<br />");
    };

    socket.onopen = function() {
        console.log("WebSocket Connected");
        document.getElementById('header-ws').innerHTML = 'connected <span class="bullet"></span>';
    };

    socket.onerror = function(error) {
        console.log("WebSocket Error: " + error);
        document.getElementById('header-ws').innerHTML = 'error <span class="bullet bullet-red"></span>';
    };

    socket.onclose = function(event) {
        console.log("WebSocket Close: " + event);
        document.getElementById('header-ws').innerHTML = 'closed <span class="bullet bullet-red"></span>';
    };
}

function updateMockdef(evt) {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    const name = evt.id.split('_')[0];

    if(evt.tagName.toLowerCase() === "input") {
        if(evt.type === "checkbox") {
            globalMockdefObj[index][name] = evt.checked;
        } else if(name === "freezetimems" || name === "statuscode") { // TODO set input type=number on these and css hide arrows
            globalMockdefObj[index][name] = Number(evt.value);
        } else if(evt.type === "text") {
            globalMockdefObj[index][name] = evt.value;
        } 
    } else if(evt.tagName.toLowerCase() === "select") {
        globalMockdefObj[index][name] = evt.value;
    } else if(evt.tagName.toLowerCase() === "textarea") {
        if(name === "payload") {
            globalMockdefObj[index][name] = JSON.parse(evt.value);
        } else {
            globalMockdefObj[index][name] = evt.value;
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
    const index = Number(evt.id.split('_').slice(-1)[0]);
    const name = evt.id.split('_')[0];

    if(evt.tagName.toLowerCase() === "input") {
        if(evt.type === "checkbox") {
            globalProxydefObj[index][name] = evt.checked;
        } else if(evt.type === "text") {
            globalProxydefObj[index][name] = evt.value;
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