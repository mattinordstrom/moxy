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

    const wsMaxAttempts = 5;
    //const wsReconnectDelay = 5000;

    let reconnectTimeout = null;

    let pingInterval = null;
    let pingTimeout = null;

    const createWSocket = () => {
        if(wSocket !== null) {
            wSocket.close();
            
            wSocket.onclose = null;
            wSocket.onerror = null;
            wSocket.onmessage = null;
            wSocket.onopen = null;

            wSocket = null;

            clearInterval(pingInterval);
            clearTimeout(pingTimeout);

            if(reconnectTimeout !== null) {
                clearTimeout(reconnectTimeout);
                reconnectTimeout = null;
            }
        }

        wSocket = new WebSocket("ws://localhost:"+Number(location.port)+"/moxyws");
        
        pingInterval = setInterval(sendPing, 5000);
    }

    const sendPing = () => {
        if (wSocket.readyState === WebSocket.OPEN) {
          wSocket.send("ping");
          
          pingTimeout = setTimeout(() => {
            //clearTimeout(reconnectTimeout);

            console.log("No ping pong response (open socket)");
            document.getElementById('header-ws').innerHTML = 'No ping response. UI opened in other tab? <span class="bullet bullet-red"></span>';
            clearPingTimeout();
          }, 3000);
        } else if(wSocket.readyState === WebSocket.CLOSED) {
            if (getWSAttempts() >= getWSMaxAttempts()) {
                document.getElementById('header-ws').innerHTML = 'Closed <span class="bullet bullet-red"></span>';
                clearInterval(pingInterval);
                return;
            }

            setWSAttempts(getWSAttempts() + 1);
            wsSetup();

            console.log("No ping pong response (closed socket)");
            document.getElementById('header-ws').innerHTML = 'Reconnecting... <span class="bullet bullet-red"></span>';

        }
    }

    const clearPingTimeout = () => {
        clearTimeout(pingTimeout);
    }

    const setReconnectTimeout = (rto) => {
        reconnectTimeout = rto;
    }

    const getWSocket = () => wSocket;

    const setWSAttempts = (wsa) => {
        wsAttempts = wsa;
    }

    const getWSAttempts = () => wsAttempts;

    const getWSMaxAttempts = () => wsMaxAttempts;

    const getWSReconnectDelay = () => wsReconnectDelay;

    return { createWSocket, getWSocket, setWSAttempts, getWSAttempts, getWSMaxAttempts, getWSReconnectDelay, clearPingTimeout, setReconnectTimeout };
})();

const SVGModule = (() => {
    const getPen = (width, height) => {
        return '<svg xmlns="http://www.w3.org/2000/svg" height="'+height+'px" width="'+width+'px" viewBox="0 0 512 512"><!--!Font Awesome Free 6.5.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2023 Fonticons, Inc.--><path d="M362.7 19.3L314.3 67.7 444.3 197.7l48.4-48.4c25-25 25-65.5 0-90.5L453.3 19.3c-25-25-65.5-25-90.5 0zm-71 71L58.6 323.5c-10.4 10.4-18 23.3-22.2 37.4L1 481.2C-1.5 489.7 .8 498.8 7 505s15.3 8.5 23.7 6.1l120.3-35.4c14.1-4.2 27-11.8 37.4-22.2L421.7 220.3 291.7 90.3z"/></svg>';
    }

    const getUpArrows = (width, height) => {
        return '<svg xmlns="http://www.w3.org/2000/svg" height="'+height+'px" width="'+width+'px" viewBox="0 0 448 512"><!--!Font Awesome Free 6.5.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2023 Fonticons, Inc.--><path d="M246.6 41.4c-12.5-12.5-32.8-12.5-45.3 0l-160 160c-12.5 12.5-12.5 32.8 0 45.3s32.8 12.5 45.3 0L224 109.3 361.4 246.6c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3l-160-160zm160 352l-160-160c-12.5-12.5-32.8-12.5-45.3 0l-160 160c-12.5 12.5-12.5 32.8 0 45.3s32.8 12.5 45.3 0L224 301.3 361.4 438.6c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3z"/></svg>';
    }

    const getDownArrows = (width, height) => {
        return '<svg xmlns="http://www.w3.org/2000/svg" height="'+height+'px" width="'+width+'px" viewBox="0 0 448 512"><!--!Font Awesome Free 6.5.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2023 Fonticons, Inc.--><path d="M246.6 470.6c-12.5 12.5-32.8 12.5-45.3 0l-160-160c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0L224 402.7 361.4 265.4c12.5-12.5 32.8-12.5 45.3 0s12.5 32.8 0 45.3l-160 160zm160-352l-160 160c-12.5 12.5-32.8 12.5-45.3 0l-160-160c-12.5-12.5-12.5-32.8 0-45.3s32.8-12.5 45.3 0L224 210.7 361.4 73.4c12.5-12.5 32.8-12.5 45.3 0s12.5 32.8 0 45.3z"/></svg>';
    }

    const getUpArrow = (width, height) => {
        return '<svg xmlns="http://www.w3.org/2000/svg" height="'+height+'px" width="'+width+'px" viewBox="0 0 448 512"><!--!Font Awesome Free 6.5.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2023 Fonticons, Inc.--><path d="M201.4 137.4c12.5-12.5 32.8-12.5 45.3 0l160 160c12.5 12.5 12.5 32.8 0 45.3s-32.8 12.5-45.3 0L224 205.3 86.6 342.6c-12.5 12.5-32.8 12.5-45.3 0s-12.5-32.8 0-45.3l160-160z"/></svg>';
    }

    const getDownArrow = (width, height) => {
        return '<svg xmlns="http://www.w3.org/2000/svg" height="'+height+'px" width="'+width+'px" viewBox="0 0 448 512"><!--!Font Awesome Free 6.5.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2023 Fonticons, Inc.--><path d="M201.4 342.6c12.5 12.5 32.8 12.5 45.3 0l160-160c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0L224 274.7 86.6 137.4c-12.5-12.5-32.8-12.5-45.3 0s-12.5 32.8 0 45.3l160 160z"/></svg>';
    }

    const getX = (width, height) => {
        return '<svg xmlns="http://www.w3.org/2000/svg" height="'+height+'px" width="'+width+'px" viewBox="0 0 384 512"><!--!Font Awesome Free 6.5.1 by @fontawesome - https://fontawesome.com License - https://fontawesome.com/license/free Copyright 2023 Fonticons, Inc.--><path d="M342.6 150.6c12.5-12.5 12.5-32.8 0-45.3s-32.8-12.5-45.3 0L192 210.7 86.6 105.4c-12.5-12.5-32.8-12.5-45.3 0s-12.5 32.8 0 45.3L146.7 256 41.4 361.4c-12.5 12.5-12.5 32.8 0 45.3s32.8 12.5 45.3 0L192 301.3 297.4 406.6c12.5 12.5 32.8 12.5 45.3 0s12.5-32.8 0-45.3L237.3 256 342.6 150.6z"/></svg>';
    }

    return { getPen, getUpArrows, getDownArrows, getUpArrow, getDownArrow, getX }
})();