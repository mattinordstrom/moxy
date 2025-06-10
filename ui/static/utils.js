const safelyParseJSON = (data) => {
    if (data === "null") return null;
    if (data === "true") return true;
    if (data === "false") return false;

    const couldBeJSON = typeof data === 'string' && (data.trim().startsWith('{') || data.trim().startsWith('['));
    if (typeof data === 'number' || data === '' || !couldBeJSON) {
        return data;
    }

    try {
        return JSON.parse(data);
    } catch (e) {
        const errorMsg = "FRONTEND ERROR! Could not parse json. " + typeof data + " ("+e+")";

        console.error(errorMsg);
        
        const logMsg = '<span class="square square-red"></span> ' + errorMsg + '<br />';
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', logMsg);

        return data;
    }
}