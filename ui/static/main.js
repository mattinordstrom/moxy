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
            '<div class="mock-obj"><label for="active_mock'+i+'">active:</label><input class="cbox" type="checkbox" name="active_mock'+i+'" id="active_mock'+i+'" ' + (mockEntityData['active'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="freezetime'+i+'">freezetimems:</label><input type="text" name="freezetime'+i+'" id="freezetime'+i+'" value="' + mockEntityData['freezetimems'] + '"></input></div>' +
            '<div class="mock-obj"><label for="matchpathexact'+i+'">matchPathExact:</label><input class="cbox" type="checkbox" name="matchpathexact'+i+'" id="matchpathexact'+i+'" ' + (mockEntityData['matchPathExact'] ? "checked" : "") + '></input></div>' +
            '<div class="mock-obj"><label for="method'+i+'">method:</label><select class="slct" name="method'+i+'" id="method'+i+'">' +
            '<option value="GET" ' + (mockEntityData['method'] === "GET" ? "selected" : "") + '>GET</option>' +
            '<option value="POST" ' + (mockEntityData['method'] === "POST" ? "selected" : "") + '>POST</option>' +
            '<option value="PUT" ' + (mockEntityData['method'] === "PUT" ? "selected" : "") + '>PUT</option>' +
            '<option value="DELETE" ' + (mockEntityData['method'] === "DELETE" ? "selected" : "") + '>DELETE</option>' + 
            '</select></div>';

            if(mockEntityData['payload']) {
                mockEntity += '<div class="mock-obj"><label for="payload'+i+'">payload:</label><textarea rows="8" cols="40" name="payload'+i+'" id="payload'+i+'">' + JSON.stringify(mockEntityData['payload'], null, 2) + '</textarea></div>';
            } else if (mockEntityData['payloadFromFile']) {
                mockEntity += '<div class="mock-obj"><label for="payloadFromFile'+i+'">payloadFromFile:</label><input class="input-wide" type="text" name="payloadFromFile'+i+'" id="payloadFromFile'+i+'" value="' + mockEntityData['payloadFromFile'] + '"></input></div>';
            }

            mockEntity += '<div class="mock-obj"><label for="statuscode'+i+'">statuscode:</label><input type="text" name="statuscode'+i+'" id="statuscode'+i+'" value="' + mockEntityData['statuscode'] + '"></input></div>' +
            '<div class="mock-obj"><label for="urlpart_mock'+i+'">urlpart:</label><input class="input-wide" type="text" name="urlpart_mock'+i+'" id="urlpart_mock'+i+'" value="' + mockEntityData['urlpart'] + '"></input></div>' +
            '</div>';

            document.getElementById('mock-content-container').innerHTML += mockEntity + "<br />";
    };

    // Add some space to between last entity and the footer
    document.getElementById('mock-content-container').innerHTML += '<div style="margin-bottom:200px"></div>';
}

function renderProxydef(data) {
    for (let i = 0; i < data.length; i++) {
        const proxyEntityData = data[i];
        let proxyEntity = '<div class="proxymock-content">' +
            '<div class="proxy-obj"><label for="active_proxy'+i+'">active:</label><input class="cbox" type="checkbox" name="active_proxy'+i+'" id="active_proxy'+i+'" ' + (proxyEntityData['active'] ? "checked" : "") + '></input></div>' +    
            '<div class="proxy-obj"><label for="target'+i+'">target:</label><input class="input-wide" type="text" name="target'+i+'" id="target'+i+'" value="' + proxyEntityData['target'] + '"></input></div>' +
            '<div class="proxy-obj"><label for="urlpart_proxy'+i+'">urlpart:</label><input class="input-wide" type="text" name="urlpart_proxy'+i+'" id="urlpart_proxy'+i+'" value="' + proxyEntityData['urlpart'] + '"></input></div>' +
            '<div class="proxy-obj"><label for="verbose'+i+'">verbose:</label><input class="cbox" type="checkbox" name="verbose'+i+'" id="verbose'+i+'" ' + (proxyEntityData['verbose'] ? "checked" : "") + '></input></div>' +    
            '</div>';

        document.getElementById('proxy-content-container').innerHTML += proxyEntity + "<br />";
    };

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