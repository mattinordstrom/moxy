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
    data.forEach((mockEntityData) => {
        let mockEntity = '<div class="proxymock-content">' +
            '<div class="mock-obj"><label for="active_mock1">active:</label><input class="cbox" type="checkbox" name="active_mock1" id="active_mock1" ' + (mockEntityData['active'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="freezetime1">freezetimems:</label><input type="text" name="freezetime1" id="freezetime1" value="' + mockEntityData['freezetimems'] + '"></input></div>' +
            '<div class="mock-obj"><label for="matchpathexact1">matchPathExact:</label><input class="cbox" type="checkbox" name="matchpathexact1" id="matchpathexact1" ' + (mockEntityData['matchPathExact'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="method1">method:</label><select class="slct" name="method1" id="method1">' +
            '<option value="GET" ' + (mockEntityData['method'] === "GET" ? "selected" : "") + '>GET</option>' +
            '<option value="POST" ' + (mockEntityData['method'] === "POST" ? "selected" : "") + '>POST</option>' +
            '<option value="PUT" ' + (mockEntityData['method'] === "PUT" ? "selected" : "") + '>PUT</option>' +
            '<option value="DELETE" ' + (mockEntityData['method'] === "DELETE" ? "selected" : "") + '>DELETE</option>' + 
            '</select></div>';

            if(mockEntityData['payload']) {
                mockEntity += '<div class="mock-obj"><label for="payload1">payload:</label><textarea rows="8" cols="40" name="payload1" id="payload1">' + JSON.stringify(mockEntityData['payload']) + '</textarea></div>';
            } else if (mockEntityData['payloadFromFile']) {
                mockEntity += '<div class="mock-obj"><label for="payloadFromFile1">payloadFromFile:</label><input class="input-wide" type="text" name="payloadFromFile1" id="payloadFromFile1" value="' + mockEntityData['payloadFromFile'] + '"></input></div>';
            }

            mockEntity += '<div class="mock-obj"><label for="statuscode1">statuscode:</label><input type="text" name="statuscode1" id="statuscode1" value="' + mockEntityData['statuscode'] + '"></input></div>' +
            '<div class="mock-obj"><label for="urlpart_mock1">urlpart:</label><input class="input-wide" type="text" name="urlpart_mock1" id="urlpart_mock1" value="' + mockEntityData['urlpart'] + '"></input></div>' +
            '</div>';

            document.getElementById('mock-content-container').innerHTML += mockEntity + "<br />";
    });

    // Add some space to between last entity and the footer
    document.getElementById('mock-content-container').innerHTML += '<div style="margin-bottom:200px"></div>';
}

function renderProxydef(data) {
    data.forEach((proxyEntityData) => {
        let proxyEntity = '<div class="proxymock-content">' +
            '<div class="proxy-obj"><label for="active_proxy1">active:</label><input class="cbox" type="checkbox" name="active_proxy1" id="active_proxy1" ' + (proxyEntityData['active'] ? "checked" : "") + '></input></div>' +    
            '<div class="proxy-obj"><label for="target1">target:</label><input class="input-wide" type="text" name="target1" id="target1" value="' + proxyEntityData['target'] + '"></input></div>' +
            '<div class="proxy-obj"><label for="urlpart_proxy1">urlpart:</label><input class="input-wide" type="text" name="urlpart_proxy1" id="urlpart_proxy1" value="' + proxyEntityData['urlpart'] + '"></input></div>' +
            '<div class="proxy-obj"><label for="verbose1">verbose:</label><input class="cbox" type="checkbox" name="verbose1" id="verbose1" ' + (proxyEntityData['verbose'] ? "checked" : "") + '></input></div>' +    
            '</div>';

        document.getElementById('proxy-content-container').innerHTML += proxyEntity + "<br />";
    });

    // Add some space to between last entity and the footer
    document.getElementById('proxy-content-container').innerHTML += '<div style="margin-bottom:200px"></div>';
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

// TODO update json files when something changes (onfocus, onclick...)
function dostuff() {
    //console.log("dostuff: " + globalMockdefObj);

    //TODO
    globalMockdefObj[0].active = false;

    fetch('/moxyadminui/mockdef', {
        method: "POST",
        body: JSON.stringify(globalMockdefObj),
    })
    .then(data => {
        console.log("Success POST");
    })
    .catch(error => {
        console.error('POST error:', error);
    });
}