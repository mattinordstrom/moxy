const safelyParseJSON = (data) => {
    if (typeof data === 'number' || data === '') {
        return data;
    }

    try {
        return JSON.parse(data);
    } catch (e) {
        const errorMsg = "FRONTEND ERROR! Could not parse json. Returning data as " + typeof data + " ("+e+")";

        console.error(errorMsg);
        
        const logMsg = '<span class="square square-red"></span> ' + errorMsg + '<br />';
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', logMsg);

        return data;
    }
}