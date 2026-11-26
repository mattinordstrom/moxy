const MockDefModule = (() => {
    let mockDefs = [];

    const set = (data) => {
        data.forEach((mock, i) => {
            // We make shallow copies of the array. To prevent external modifications of objects from affecting our internal state, we require 'payload' to be a stringified version of any object.
            if(mock.payload && typeof mock.payload !== 'string' && typeof mock.payload !== 'number') {
                throw new Error('Payload must be stringified: ' + i);
            }
        });

        mockDefs = data.map(obj => ({ ...obj }));
    }

    const get = () => mockDefs.map(obj => ({ ...obj }));

    const remove = (index) => {
        if (index < 0 || index >= mockDefs.length) {
            throw new Error('Index out of bounds mockDefs');
        }

        mockDefs.splice(index, 1);
    }

    // Payloads are kept as their source text, since parsing them would corrupt big ints, 1.10 and key order
    const parse = (text) => {
        const data = JSON.parse(text);
        const sources = JSONText.elementMembers(text, 'payload');

        data.forEach((mock, i) => {
            if(mock.payload !== null && typeof mock.payload === 'object') {
                mock.payload = JSONText.pretty(sources[i] ?? JSON.stringify(mock.payload));
            } else if(typeof mock.payload === 'boolean') {
                mock.payload = String(mock.payload);
            }
        });

        return data;
    }

    const payloadJSON = (payload) => {
        const parsed = safelyParseJSON(payload);
        if(parsed !== null && typeof parsed === 'object') {
            return JSONText.compact(payload);
        }

        return JSON.stringify(parsed);
    }

    const toJSON = () => '[' + mockDefs.map(({ payload, ...rest }) => {
        const fields = JSON.stringify(rest);
        if(payload === undefined) {
            return fields;
        }

        return fields.slice(0, -1) + (fields === '{}' ? '' : ',') + '"payload":' + payloadJSON(payload) + '}';
    }).join(',') + ']';

    return { set, get, remove, parse, toJSON };
})();

const ProxyDefModule = (() => {
    let proxyDefs = [];

    const set = (data) => {
        proxyDefs = data.map(obj => ({ ...obj }));
    }

    const get = () => proxyDefs.map(obj => ({ ...obj }));

    const remove = (index) => {
        if (index < 0 || index >= proxyDefs.length) {
            throw new Error('Index out of bounds proxyDefs');
        }

        proxyDefs.splice(index, 1);
    }

    return { set, get, remove };
})();

const PayloadFromFileModule = (() => {
    let payloadPath = '';
    let payloadFiles = [];

    const setPayloadFiles = (data) => {
        payloadFiles = [...data];
    }

    const getPayloadFiles = () => [...payloadFiles];
    
    const setPayloadPath = (str) => {
        payloadPath = str;
    }

    const getPayloadPath = () => payloadPath;

    return { setPayloadFiles, getPayloadFiles, setPayloadPath, getPayloadPath };
})();

const WSModule = (() => {
    let wSocket = null;
    let wsAttempts = 0;

    const wsMaxAttempts = 3;
    const wsReconnectDelay = 5000;

    const createWSocket = () => {
        if(wSocket !== null) {
            wSocket.close();
            
            wSocket.onclose = null;
            wSocket.onerror = null;
            wSocket.onmessage = null;
            wSocket.onopen = null;

            wSocket = null;
        }

        const scheme = location.protocol.startsWith("https") ? "wss://" : "ws://";
        wSocket = new WebSocket(scheme + location.host + "/moxyws");
    }

    const getWSocket = () => wSocket;

    const setWSAttempts = (wsa) => {
        wsAttempts = wsa;
    }

    const getWSAttempts = () => wsAttempts;

    const getWSMaxAttempts = () => wsMaxAttempts;

    const getWSReconnectDelay = () => wsReconnectDelay;

    return { createWSocket, getWSocket, setWSAttempts, getWSAttempts, getWSMaxAttempts, getWSReconnectDelay };
})();
