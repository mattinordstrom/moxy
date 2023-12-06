const MockDefModule = (() => {
    let mockDefs = [];

    const set = (data) => {
        data.forEach((mock, i) => {
            // We make shallow copies of the array. To prevent external modifications of objects from affecting our internal state, we require 'payload' to be a stringified version of any object.
            if(mock.payload && typeof mock.payload !== 'string' && typeof mock.payload !== 'number') {
                throw new Error('Payload must be stringified: ' + i);
            }
        });

        mockDefs = [...data];
    }

    const get = () => [...mockDefs];

    const remove = (index) => {
        if (index < 0 || index >= mockDefs.length) {
            throw new Error('Index out of bounds mockDefs');
        }

        mockDefs.splice(index, 1);
    }

    const stringifyPayload = (data) => {
        data.forEach((mock, i) => {
            if(mock.payload && typeof mock.payload === 'object') {
                mock.payload = JSON.stringify(mock.payload);
            }
        });

        return data;
    }

    const getWithJSONPayload = () => {
        return MockDefModule.get().map(mock => {
            if (mock.payload && typeof mock.payload !== 'number') {
                return {
                    ...mock,
                    payload: safelyParseJSON(mock.payload)
                };
            } else {
                return { ...mock };
            }
        });
    }

    return { set, get, remove, stringifyPayload, getWithJSONPayload };
})();

const ProxyDefModule = (() => {
    let proxyDefs = [];

    const set = (data) => {
        proxyDefs = [...data];
    }

    const get = () => [...proxyDefs];

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

        wSocket = new WebSocket("ws://localhost:"+Number(location.port)+"/moxyws");
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