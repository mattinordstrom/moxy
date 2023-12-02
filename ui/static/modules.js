const MockDefModule = (() => {
    let mockdefObj = [];

    const setMockdefObj = (data) => {
        mockdefObj = data;
    }

    const getMockdefObj = () => mockdefObj;

    const removeFromMockdefObj = (index) => {
        let newArray = [...mockdefObj];
        newArray.splice(index, 1);
        setMockdefObj(newArray);
    }

    return {
        set: setMockdefObj,
        get: getMockdefObj,
        remove: removeFromMockdefObj
    };
})();

const ProxyDefModule = (() => {
    let proxydefObj = [];

    const setProxydefObj = (data) => {
        proxydefObj = data;
    }

    const getProxydefObj = () => proxydefObj;

    const removeFromProxydefObj = (index) => {
        let newArray = [...proxydefObj];
        newArray.splice(index, 1);
        setProxydefObj(newArray);
    }

    return {
        set: setProxydefObj,
        get: getProxydefObj,
        remove: removeFromProxydefObj
    };
})();

const PayloadFromFileModule = (() => {
    let payloadPath = '';
    let payloadFiles = [];

    const setPayloadFiles = (data) => {
        payloadFiles = data;
    }

    const getPayloadFiles = () => payloadFiles;
    
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