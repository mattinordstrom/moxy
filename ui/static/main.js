let uiSettings = { darkMode: true, useHotkeys: true };

const fetchUISettings = async () => {
    try {
        const response = await fetch('/moxyadminui/uisettings');
        if (response.ok) {
            uiSettings = await response.json();
        }
    } catch (error) {
        console.error('Fetch error uisettings:', error);
    }
}

const saveUISettings = () => postJSON('/moxyadminui/uisettings', JSON.stringify(uiSettings), 'UI settings');

const logError = (message) => {
    console.error(message);
    appendLogEntry('<span class="square square-red"></span> ' + escapeHTML(message) + '<br />');
}

const logInfo = (message) => {
    appendLogEntry('<span class="square"></span> ' + escapeHTML(message) + '<br />');
}

const postJSON = async (url, body, what) => {
    try {
        const response = await fetch(url, {
            method: "POST",
            headers: { 'Content-Type': 'application/json' },
            body: body,
        });

        if (!response.ok) {
            logError('Saving ' + what + ' failed (' + response.status + '): ' + (await response.text()).trim());
            return false;
        }

        return true;
    } catch (error) {
        logError('Saving ' + what + ' failed: ' + error.message);
        return false;
    }
}

const initFunc = async () => {
    console.log("initFunc");

    await loadDefs('mock');
    await loadDefs('proxy');
    await fetchSettings();
    await fetchUISettings();

    hotKeysSetup();
    darkModeSetup();

    if(localStorage.getItem('moxyMockCompactActive') === 'true') {
        showCompactList(true);
    }
    if(localStorage.getItem('moxyProxyCompactActive') === 'true') {
        showCompactListProxy(true);
    }
    if(localStorage.getItem('moxyShowOnlyMocks') === 'true') {
        showOnlyMocks(true);
    }

    const urlParams = new URLSearchParams(window.location.search);
    const mode = urlParams.get('mode');

    //websocket setup
    if (mode !== 'test') {
        wsSetup();
    } else {
        document.getElementById("cover").style.display = "none";
    }

    //click events setup
    clickEvtSetup();
}

let lastDefsText = { mock: null, proxy: null };

const defsFile = (type) => type + 'def.json';

const fetchDefsText = async (type) => {
    const response = await fetch('/moxyadminui/' + type + 'def', { cache: 'no-store' });
    const text = await response.text();

    if (!response.ok) {
        throw new Error('HTTP ' + response.status + ': ' + text.trim());
    }

    return text;
}

const applyDefsText = (type, text) => {
    if(type === 'mock') {
        MockDefModule.set(MockDefModule.parse(text));
    } else {
        ProxyDefModule.set(JSON.parse(text));
    }

    lastDefsText[type] = text;
}

const getDefs = (type) => type === 'mock' ? MockDefModule.get() : ProxyDefModule.get();

const setDefs = (type, defs) => type === 'mock' ? MockDefModule.set(defs) : ProxyDefModule.set(defs);

const loadDefs = async (type) => {
    try {
        applyDefsText(type, await fetchDefsText(type));
    } catch (error) {
        logError('Could not load ' + defsFile(type) + ', changes will not be saved until it loads: ' + error.message);
    }

    rerenderDefs(type);
}

// Saving always starts from what is on disk, so a stale tab or an external edit is never overwritten
const syncWithDisk = async (type) => {
    try {
        const text = await fetchDefsText(type);
        if(text === lastDefsText[type]) {
            return 'unchanged';
        }

        applyDefsText(type, text);
        rerenderDefs(type);

        return 'changed';
    } catch (error) {
        logError('Not saved, could not read ' + defsFile(type) + ': ' + error.message);
        return null;
    }
}

const defsUpToDate = async (type) => {
    const state = await syncWithDisk(type);
    if(state === 'changed') {
        logInfo(defsFile(type) + ' changed on disk \u2014 reloaded, please try again');
    }

    return state === 'unchanged';
}

const saveDefs = async (type) => {
    const body = type === 'mock' ? MockDefModule.toJSON() : JSON.stringify(ProxyDefModule.get());

    if(!await postJSON('/moxyadminui/' + type + 'def', body, defsFile(type))) {
        return;
    }

    try {
        lastDefsText[type] = await fetchDefsText(type);
    } catch (error) {
        lastDefsText[type] = null;
    }

    console.log('Success POST ' + type + 'def');
}

const normalizeFieldValue = (value) => {
    const parsed = safelyParseJSON(value);
    return (parsed !== null && typeof parsed === 'object') ? JSONText.compact(value) : String(parsed);
}

// Only keystrokes that are not saved yet are worth protecting from a rebuild
const focusedFieldHasUnsavedEdit = (type) => {
    const container = document.getElementById(type + '-content-container');
    const focused = document.activeElement;

    if(!container || !focused || !container.contains(focused)) {
        return false;
    }

    const tag = focused.tagName.toLowerCase();
    if(tag !== 'textarea' && !(tag === 'input' && focused.type !== 'checkbox')) {
        return false;
    }

    const name = focused.id.split('_')[0];
    const index = Number(focused.id.split('_').slice(-1)[0]);
    const entity = (type === 'mock' ? MockDefModule.get() : ProxyDefModule.get())[index];

    if(!entity) {
        return false;
    }

    if(name === 'payload') {
        return normalizeFieldValue(focused.value) !== normalizeFieldValue(entity.payload);
    }

    return focused.value !== String(entity[name] === undefined || entity[name] === null ? '' : entity[name]);
}

let pendingReload = { mock: false, proxy: false };

const schedulePendingReload = (type) => {
    if(pendingReload[type]) {
        return;
    }

    pendingReload[type] = true;
    const container = document.getElementById(type + '-content-container');

    const onFocusOut = () => {
        container.removeEventListener('focusout', onFocusOut);
        pendingReload[type] = false;

        // let the change handler's POST land first, otherwise we refetch the pre-save content
        setTimeout(() => {
            reloadDefsFromDisk(type + 'def.json').then((result) => {
                if(result) {
                    appendLogEntry('<span class="square"></span> ' + result + '<br />');
                }
            });
        }, 300);
    };

    container.addEventListener('focusout', onFocusOut);
}

const reloadDefsFromDisk = async (file) => {
    const type = file === 'proxydef.json' ? 'proxy' : 'mock';

    let text;
    try {
        text = await fetchDefsText(type);
    } catch (error) {
        console.error('Fetch error on reload:', error);
        return null;
    }

    if(text === lastDefsText[type]) {
        return null;
    }

    if(focusedFieldHasUnsavedEdit(type)) {
        schedulePendingReload(type);
        return type + 'def.json changed on disk &mdash; will reload when you finish editing';
    }

    try {
        applyDefsText(type, text);
    } catch (error) {
        logError(defsFile(type) + ' changed on disk but could not be loaded: ' + error.message);
        return null;
    }

    rerenderDefs(type);

    return type + 'def.json changed on disk &mdash; reloaded';
}

const fetchSettings = async () => {
    try {
        const response = await fetch('/moxyadminui/settings');
        if (!response.ok) {
            const text = await response.text();
            
            const logMsg = '<span class="square square-red"></span> ' + escapeHTML(text) + '<br />';
            document.getElementById("footer-log").insertAdjacentHTML('afterbegin', logMsg);

            throw new Error('Error ' + response.status + ': ' + text);
        }

        const data = await response.json();

        document.getElementById('header-version').textContent = data['version'] ? 'v' + data['version'] : '';
        document.getElementById('header-port').textContent = data['port'];
        document.getElementById('header-route').textContent = data['defaultRoute'] + (data['defaultInsecure'] ? ' (insecure TLS)' : '');

        PayloadFromFileModule.setPayloadPath(data['payloadPath']);
        PayloadFromFileModule.setPayloadFiles(data['payloadFiles']);

        const maxLogEntries = data['maxLogEntries'];
        window.MoxyMaxLogEntries = (typeof maxLogEntries === 'number') ? maxLogEntries : null;
        const maxLogLabel = document.getElementById('max-log-entries-label');
        if (window.MoxyMaxLogEntries !== null) {
            maxLogLabel.textContent = 'Max log entries: ' + window.MoxyMaxLogEntries;
            maxLogLabel.style.display = '';
        } else {
            maxLogLabel.style.display = 'none';
        }
    } catch (error) {
        console.error('Fetch error settings:', error);
    }
}

const renderMockdefs = () => {
    MockDefModule.get().forEach((mockEntityData, i) => {
        let mockEntity = `
            <div class="proxymock-content">
                <div class="proxymock-entity">
                    <div style="display:flex">
                        <div><input placeholder="comment here" onchange="updateMockdef(this)" class="comment-input" type="text" spellcheck="false" name="comment_mock_${i}" id="comment_mock_${i}" value="${escapeHTML(mockEntityData['comment'] || "")}"></input></div>
                        <div><button class="small-btn" onclick="moveMock(this)" id="movemock_first_btn_${i}"><i class="fa-solid fa-angles-up"></i></button></div>
                        <div><button class="small-btn" onclick="moveMock(this)" id="movemock_up_btn_${i}"><i class="fa-solid fa-angle-up"></i></button></div>
                        <div><button class="small-btn" onclick="moveMock(this)" id="movemock_down_btn_${i}"><i class="fa-solid fa-angle-down"></i></button></div>
                        <div><button class="small-btn" onclick="moveMock(this)" id="movemock_last_btn_${i}"><i class="fa-solid fa-angles-down"></i></button></div>
                        <div><button class="small-btn small-btn-close" onclick="removeMock(this)" id="x_btn_${i}"><i class="fa-solid fa-xmark"></i></button></div>
                    </div>
                </div>
                <div class="mock-obj"><label for="active_mock_${i}">active:</label><input onclick="updateMockdef(this)" class="cbox" type="checkbox" name="active_mock_${i}" id="active_mock_${i}" ${mockEntityData['active'] ? "checked" : ""}></input></div>
                <div class="mock-obj mock_compact_exclude"><label for="freezetimems_${i}">freezetimems:</label><input onchange="updateMockdef(this)" type="number" name="freezetimems_${i}" id="freezetimems_${i}" value="${escapeHTML(mockEntityData['freezetimems'])}"></input></div>
                <div class="mock-obj"><label for="method_${i}">method:</label><select onchange="updateMockdef(this)" class="slct" name="method_${i}" id="method_${i}">
                    <option value="GET" ${mockEntityData['method'] === "GET" ? "selected" : ""}>GET</option>
                    <option value="POST" ${mockEntityData['method'] === "POST" ? "selected" : ""}>POST</option>
                    <option value="PUT" ${mockEntityData['method'] === "PUT" ? "selected" : ""}>PUT</option>
                    <option value="DELETE" ${mockEntityData['method'] === "DELETE" ? "selected" : ""}>DELETE</option>
                    <option value="PATCH" ${mockEntityData['method'] === "PATCH" ? "selected" : ""}>PATCH</option>
                    </select>
                </div>`;

        let payload = mockEntityData['payload'];
        const jsonPayload = safelyParseJSON(payload);
        if (jsonPayload === null) {
            payload = 'null';
        } else if (typeof jsonPayload === 'object') {
            payload = JSONText.pretty(payload);
        }

        mockEntity += `
                <div class="mock-obj mock_compact_exclude">
                    <label for="payload_${i}">payload:</label>
                    <textarea onchange="updateMockdef(this)" spellcheck="false" rows="20" cols="60" name="payload_${i}" id="payload_${i}">${escapeHTML(payload)}</textarea>
                </div>`;

        mockEntity += `
                <div class="mock-obj mock_compact_exclude">
                    <label style="display:flex; flex-direction: column" for="payloadFromFile_${i}">
                        payloadFromFile:
                    </label>
                    <textarea onchange="updateMockdef(this)" spellcheck="false" rows="4" cols="45" class="fixed-textarea" name="payloadFromFile_${i}" id="payloadFromFile_${i}">${escapeHTML(mockEntityData['payloadFromFile'])}</textarea>
                    <button class="mock-file-edit payload-files-btn small-btn" id="mock_file_edit_${i}" onclick="editFileFromMock(this)">
                        <i class="fa-solid fa-pen"></i>
                    </button>
                </div>`;

        mockEntity += `
                <div class="mock-obj mock_compact_exclude">
                    <label for="statuscode_${i}">statuscode:</label>
                    <input onchange="updateMockdef(this)" type="number" name="statuscode_${i}" id="statuscode_${i}" value="${escapeHTML(mockEntityData['statuscode'])}"></input>
                </div>
                <div class="mock-obj">
                    <label for="urlpart_mock_${i}"><b>urlpart:</b></label>
                    <input onchange="updateMockdef(this)" spellcheck="false" class="input-wide" type="text" name="urlpart_mock_${i}" id="urlpart_mock_${i}" value="${escapeHTML(mockEntityData['urlpart'])}"></input>
                </div>
            </div>`;

        document.getElementById('mock-content-container').innerHTML += mockEntity + "<br />";
    });
}

const renderProxydefs = () => {
    ProxyDefModule.get().forEach((proxyEntityData, i) => {
        let proxyEntity = `
            <div class="proxymock-content">
                <div class="proxymock-entity">
                    <div style="display:flex">
                        <div><input placeholder="comment here" onchange="updateProxydef(this)" class="comment-input" type="text" spellcheck="false" name="comment_proxy_${i}" id="comment_proxy_${i}" value="${escapeHTML(proxyEntityData['comment'] || "")}"></input></div>
                        <div><button class="small-btn" onclick="moveProxy(this)" id="moveproxy_first_btn_${i}"><i class="fa-solid fa-angles-up"></i></button></div>
                        <div><button class="small-btn" onclick="moveProxy(this)" id="moveproxy_up_btn_${i}"><i class="fa-solid fa-angle-up"></i></button></div>
                        <div><button class="small-btn" onclick="moveProxy(this)" id="moveproxy_down_btn_${i}"><i class="fa-solid fa-angle-down"></i></button></div>
                        <div><button class="small-btn" onclick="moveProxy(this)" id="moveproxy_last_btn_${i}"><i class="fa-solid fa-angles-down"></i></button></div>
                        <div><button class="small-btn small-btn-close" onclick="removeProxy(this)" id="x_btn_${i}"><i class="fa-solid fa-xmark"></i></button></div>
                    </div>
                </div>
                <div class="proxy-obj"><label for="active_proxy_${i}">active:</label><input onclick="updateProxydef(this)" class="cbox" type="checkbox" name="active_proxy_${i}" id="active_proxy_${i}" ${proxyEntityData['active'] ? "checked" : ""}></input></div>    
                <div class="proxy-obj proxy_compact_exclude"><label for="target_${i}">target:</label><input onchange="updateProxydef(this)" spellcheck="false" class="input-wide" type="text" name="target_${i}" id="target_${i}" value="${escapeHTML(proxyEntityData['target'])}"></input></div>
                <div class="proxy-obj"><label for="urlpart_proxy_${i}"><b>urlpart:</b></label><input onchange="updateProxydef(this)" spellcheck="false" class="input-wide" type="text" name="urlpart_proxy_${i}" id="urlpart_proxy_${i}" value="${escapeHTML(proxyEntityData['urlpart'])}"></input></div>
                <div class="proxy-obj proxy_compact_exclude"><label for="verbose_${i}">verbose:</label><input onclick="updateProxydef(this)" class="cbox" type="checkbox" name="verbose_${i}" id="verbose_${i}" ${proxyEntityData['verbose'] ? "checked" : ""}></input></div> 
                <div class="proxy-obj proxy_compact_exclude"><label for="insecure_${i}" title="Skip the TLS certificate check, e.g. for a self-signed dev server">insecure:</label><input onclick="updateProxydef(this)" class="cbox" type="checkbox" name="insecure_${i}" id="insecure_${i}" ${proxyEntityData['insecure'] ? "checked" : ""}></input></div>
            </div>`;

        document.getElementById('proxy-content-container').innerHTML += proxyEntity + "<br />";
    });
}

const maximizeMock = (index) => {
    const payload = document.getElementById('payload_'+index);
    if(!payload) {
        return;
    }

    payload.style.height = '470px';
    payload.style.width = '680px';
}

const showCompactList = (forceActive) => {
    const isActive = localStorage.getItem('moxyMockCompactActive') === 'true'

    const toggleButton = document.getElementById('toggle_compactlist_bullet');
    const elements = document.querySelectorAll('.mock_compact_exclude');

    const shouldHide = forceActive || !isActive;

    localStorage.setItem('moxyMockCompactActive', shouldHide.toString());

    if (forceActive || shouldHide) {
        toggleButton.classList.add('action-btn-bullet-active');
    } else {
        toggleButton.classList.remove('action-btn-bullet-active');
    }

    elements.forEach(el => {
        el.style.display = shouldHide ? 'none' : '';
    });
}

const showCompactListProxy = (forceActive) => {
    const isActive = localStorage.getItem('moxyProxyCompactActive') === 'true'

    const toggleButton = document.getElementById('toggle_compactlist_proxy_bullet');
    const elements = document.querySelectorAll('.proxy_compact_exclude');

    const shouldHide = forceActive || !isActive;

    localStorage.setItem('moxyProxyCompactActive', shouldHide.toString());

    if (forceActive || shouldHide) {
        toggleButton.classList.add('action-btn-bullet-active');
    } else {
        toggleButton.classList.remove('action-btn-bullet-active');
    }

    elements.forEach(el => {
        el.style.display = shouldHide ? 'none' : '';
    });
}

const showOnlyMocks = (forceActive) => {
    const isActive = localStorage.getItem('moxyShowOnlyMocks') === 'true'

    localStorage.setItem('moxyShowOnlyMocks', forceActive || !isActive);

    document.getElementById('toggle_only_mocks_bullet').classList.toggle('action-btn-bullet-active');

    if(forceActive || document.getElementsByClassName('right')[0].style.display === '') {
        document.getElementsByClassName('right')[0].style.display = 'none';

        maximizeMock(0);
        maximizeMock(1);
    } else {
        document.getElementsByClassName('right')[0].style.display = '';
    }
    
}

const editFileFromMock = (el) => {
    window.scrollTo({ top: 0, behavior: 'smooth' });

    if(document.getElementById('payloadFiles').style.display === 'block'){
        closeListPayloadFiles();
    }

    const index = Number(el.id.split('_').slice(-1)[0]);
    const fullPath = document.getElementById('payloadFromFile_'+index).value;

    if(listPayloadFiles('edit_file_btn')) {
        editFile(null, fullPath);
    }
}

const listPayloadFiles = async (evtSource) => {
    if(document.getElementById('payloadFiles').style.display === 'block' && evtSource === 'payload_files_btn') {
        closeListPayloadFiles();
        return false;
    }

    document.getElementById('payload_files_btn_bullet').classList.toggle('action-btn-bullet-active');

    const response = await fetch('/moxyadminui/settings');
    const data = await response.json();
    PayloadFromFileModule.setPayloadFiles(data['payloadFiles']);

    document.getElementById('payloadFiles').style.display = 'block';
    document.getElementById('payloadFilesPath').innerHTML = `<br />${escapeHTML(PayloadFromFileModule.getPayloadPath())}<hr /><br />`;

    let filesListHtml = '<div>';
    PayloadFromFileModule.getPayloadFiles().forEach((file) => {
        filesListHtml += `
            <div style="display:flex">
                <div style="min-width:275px">${escapeHTML(file)}</div>
                &nbsp;&nbsp;<button onclick="navigator.clipboard.writeText(${escapeHTML(JSON.stringify(PayloadFromFileModule.getPayloadPath() + file))})">Copy full path</button>
                &nbsp;&nbsp;<button onclick="editFile(this, ${escapeHTML(JSON.stringify(PayloadFromFileModule.getPayloadPath() + file))})">Edit file</button>
            </div><br />
        `;
    });
    filesListHtml += '</div>';

    document.getElementById('payloadFilesContent').innerHTML = filesListHtml;
   
    return true;
}

const editFile = async (btnEl, fullPath) => {
    const filename = fullPath.split('/').pop();
    
    if(!PayloadFromFileModule.getPayloadFiles().includes(filename)) {
        document.getElementById('payloadedit').disabled = true;
        document.getElementById('payloadedit').value = '';
        document.getElementById('editfiletitle').innerText = '---';

        //console.log('Invalid file ' + filename);
        return;
    }

    const response = await fetch('/moxyadminui/editpayloadfile?file=' + filename, { cache: 'no-store' });

    let data = await response.text();
    try {
        JSON.parse(data);
        data = JSONText.pretty(data);
    } catch (error) {
        console.warn("Received response could not be parsed as JSON");
    }

    document.getElementById('editfiletitle').textContent = filename;
    document.getElementById('editfilecopybtn').innerHTML = `<button onclick="navigator.clipboard.writeText(${escapeHTML(JSON.stringify(PayloadFromFileModule.getPayloadPath() + filename))})">Copy full path</button>`;

    document.getElementById('payloadedit').disabled = false;

    document.getElementById('payloadedit').value = data;
}

const updatePayloadFile = async (el) => {
    const filename = document.getElementById('editfiletitle').innerText;

    try {
        JSON.parse(el.value);
    } catch (error) {
        logError('Not saved, ' + filename + ' is not valid JSON: ' + error.message);
        return;
    }

    const body = '{"file":' + JSON.stringify(PayloadFromFileModule.getPayloadPath() + filename) + ',"payload":' + JSONText.compact(el.value) + '}';

    if(await postJSON('/moxyadminui/editpayloadfile', body, filename)) {
        console.log("Success POST editpayloadfile");
    }
}

const closeListPayloadFiles = () => {
    document.getElementById('payloadFiles').style.display = 'none';
    document.getElementById('payload_files_btn_bullet').classList.toggle('action-btn-bullet-active');
}

const addDef = async (type, def) => {
    if(await syncWithDisk(type) === null) {
        return;
    }

    setDefs(type, [def, ...getDefs(type)]);
    rerenderDefs(type);
    await saveDefs(type);
}

const addMock = () => addDef('mock', {
    "active": true,
    "freezetimems": 0,
    "method": "GET",
    "payload": '{ "response": "abc123" }',
    "payloadFromFile": "",
    "statuscode": 200,
    "urlpart": "/api/whatever/someendpoint"
});

const addProxy = () => addDef('proxy', {
    "active": true,
    "target": "http://localhost:8080",
    "urlpart": "/api/test123",
    "verbose": false
});

const removeDef = async (type, el) => {
    if(!await defsUpToDate(type)) {
        return;
    }

    const index = Number(el.id.split('_').slice(-1)[0]);
    if(type === 'mock') {
        MockDefModule.remove(index);
    } else {
        ProxyDefModule.remove(index);
    }

    rerenderDefs(type);
    await saveDefs(type);
}

const removeMock = (el) => removeDef('mock', el);

const removeProxy = (el) => removeDef('proxy', el);

const updateDef = async (type, index, name, value) => {
    const state = await syncWithDisk(type);
    if(state === null) {
        return;
    }

    const defs = getDefs(type);
    if(!defs[index]) {
        logInfo(defsFile(type) + ' changed on disk \u2014 reloaded, please try again');
        return;
    }

    defs[index][name] = value;
    setDefs(type, defs);

    if(state === 'changed') {
        rerenderDefs(type);
        logInfo(defsFile(type) + ' changed on disk \u2014 your edit was applied on top of it');
    }

    await saveDefs(type);
}

const mockFieldValue = (el, name) => {
    const tag = el.tagName.toLowerCase();

    if(tag === "input") {
        if(el.type === "checkbox") {
            return el.checked;
        }

        if(name === "freezetimems" || name === "statuscode") {
            el.value = Math.abs(parseInt(el.value));
            if(name === "statuscode" && (!el.value || el.value === "0")) {
                el.value = "200";
            }

            return parseInt(el.value);
        }

        return el.value;
    }

    if(tag === "textarea") {
        if(name === "payload" && !isNaN(el.value) && el.value !== "") {
            return parseFloat(el.value);
        }

        if(name === "payloadFromFile" && el.value.startsWith('~')) {
            alert('Absolute path cannot start with ~');
            return undefined;
        }
    }

    return el.value;
}

const updateMockdef = (el) => {
    const name = el.id.split('_')[0];
    const value = mockFieldValue(el, name);
    if(value === undefined) {
        return;
    }

    return updateDef('mock', Number(el.id.split('_').slice(-1)[0]), name, value);
}

const updateProxydef = (el) => {
    const name = el.id.split('_')[0];
    const value = el.type === "checkbox" ? el.checked : el.value;

    return updateDef('proxy', Number(el.id.split('_').slice(-1)[0]), name, value);
}

const moving = { mock: false, proxy: false };

// A click before the previous move is saved would act on the old list and undo that move, so it is ignored
const moveDef = async (type, el) => {
    if(moving[type]) {
        return;
    }
    moving[type] = true;

    if(!await defsUpToDate(type)) {
        moving[type] = false;
        return;
    }

    const defs = getDefs(type);
    const index = Number(el.id.split('_').slice(-1)[0]);
    const way = el.id.split('_')[1];

    if(!moveEntity(el, defs)) {
        moving[type] = false;
        return;
    }

    setDefs(type, defs);

    const container = document.getElementById(type + '-content-container');
    const targetIndex = way === 'up' ? index - 1 : way === 'down' ? index + 1 : way === 'first' ? 0 : defs.length - 1;

    animateMove(container, index, targetIndex, async () => {
        rerenderDefs(type);
        await saveDefs(type);
        moving[type] = false;
    });
}

const moveMock = (el) => moveDef('mock', el);

const moveProxy = (el) => moveDef('proxy', el);

const animateMove = (container, fromIndex, toIndex, callback) => {
    const cards = container.querySelectorAll('.proxymock-content');
    const movedCard = cards[fromIndex];

    if (!movedCard || fromIndex === toIndex) {
        callback();
        return;
    }

    const rects = [];
    cards.forEach(card => rects.push(card.getBoundingClientRect()));

    const movedTarget = rects[toIndex].top;
    const movedDelta = movedTarget - rects[fromIndex].top;

    const minIdx = Math.min(fromIndex, toIndex);
    const maxIdx = Math.max(fromIndex, toIndex);
    const movingUp = toIndex < fromIndex;

    const affectedCards = [];
    for (let i = minIdx; i <= maxIdx; i++) {
        if (i === fromIndex) continue;
        const card = cards[i];
        const shiftTarget = movingUp ? rects[i + 1].top : rects[i - 1].top;
        const shiftDelta = shiftTarget - rects[i].top;
        affectedCards.push({ card, delta: shiftDelta });
    }

    movedCard.style.transition = 'none';
    movedCard.style.transform = '';
    movedCard.style.zIndex = '1';
    affectedCards.forEach(({ card }) => {
        card.style.transition = 'none';
        card.style.transform = '';
    });

    const duration = Math.abs(fromIndex - toIndex) === 1 ? '0.25s' : '0.35s';

    requestAnimationFrame(() => {
        requestAnimationFrame(() => {
            movedCard.style.transition = `transform ${duration} ease`;
            movedCard.style.transform = `translateY(${movedDelta}px)`;

            affectedCards.forEach(({ card, delta }) => {
                card.style.transition = `transform ${duration} ease`;
                card.style.transform = `translateY(${delta}px)`;
            });
        });
    });

    movedCard.addEventListener('transitionend', function handler(e) {
        if (e.propertyName !== 'transform') return;
        movedCard.removeEventListener('transitionend', handler);
        movedCard.style.transition = '';
        movedCard.style.transform = '';
        movedCard.style.zIndex = '';
        affectedCards.forEach(({ card }) => {
            card.style.transition = '';
            card.style.transform = '';
        });
        callback();
    });
}

const moveEntity = (evt, arr) => {
    const index = Number(evt.id.split('_').slice(-1)[0]);
    const way = evt.id.split('_')[1];
    let changed = false;

    if(way === 'up') {
        if (index > 0) {
            [arr[index], arr[index - 1]] = [arr[index - 1], arr[index]];
            changed = true;
        }
    } else if (way === 'down') {
        if (index < arr.length - 1) {
            [arr[index], arr[index + 1]] = [arr[index + 1], arr[index]];
            changed = true;
        }
    } else if (way === 'first') {
        if (index > 0) {
            const [item] = arr.splice(index, 1);
            arr.unshift(item);
            changed = true;
        }
    } else if (way === 'last'){
        if (index < arr.length - 1) {
            const [item] = arr.splice(index, 1);
            arr.push(item);
            changed = true;
        }
    }

    return changed;
}

const rerenderDefs = (type) => {
    if(type === "mock") {
        document.getElementById('mock-content-container').innerHTML = "";
        renderMockdefs();

        const compactIsActive = document.getElementById('toggle_compactlist_bullet').classList.contains('action-btn-bullet-active');
        if(compactIsActive) {
            showCompactList(true);
        }
        return;
    }

    document.getElementById('proxy-content-container').innerHTML = "";
    renderProxydefs();

    const compactIsActive = document.getElementById('toggle_compactlist_proxy_bullet').classList.contains('action-btn-bullet-active');
    if(compactIsActive) {
        showCompactListProxy(true);
    }
}
