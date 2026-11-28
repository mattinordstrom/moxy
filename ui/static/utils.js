const HTML_ESCAPES = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };

const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (c) => HTML_ESCAPES[c]);

// Single quotes keep everything literal in sh/bash/zsh, so a logged request can't run code when its curl is pasted
const shellQuote = (value) => "'" + String(value).replace(/'/g, "'\\''") + "'";

// Works on the JSON source text so big ints, number formatting and key order survive; input must be valid JSON
const JSONText = (() => {
    const WHITESPACE = ' \t\n\r';

    const skipWhitespace = (s, i) => {
        while (i < s.length && WHITESPACE.includes(s[i])) i++;
        return i;
    };

    const skipString = (s, i) => {
        for (i++; i < s.length && s[i] !== '"'; i++) {
            if (s[i] === '\\') i++;
        }
        return i + 1;
    };

    const skipValue = (s, i) => {
        if (s[i] === '"') return skipString(s, i);

        if (s[i] === '{' || s[i] === '[') {
            let depth = 0;
            for (; i < s.length; i++) {
                if (s[i] === '"') {
                    i = skipString(s, i) - 1;
                } else if (s[i] === '{' || s[i] === '[') {
                    depth++;
                } else if ((s[i] === '}' || s[i] === ']') && --depth === 0) {
                    return i + 1;
                }
            }
            return i;
        }

        while (i < s.length && !(',}]' + WHITESPACE).includes(s[i])) i++;
        return i;
    };

    const memberSource = (s, start, key) => {
        let found;
        let i = start + 1;
        while (i < s.length) {
            i = skipWhitespace(s, i);
            if (s[i] !== '"') break;

            const keyEnd = skipString(s, i);
            const name = JSON.parse(s.slice(i, keyEnd));
            const valueStart = skipWhitespace(s, skipWhitespace(s, keyEnd) + 1);
            const valueEnd = skipValue(s, valueStart);

            // JSON.parse keeps the last duplicate key, so do the same
            if (name === key) found = s.slice(valueStart, valueEnd);

            i = skipWhitespace(s, valueEnd);
            if (s[i] === ',') i++;
        }
        return found;
    };

    const elementMembers = (s, key) => {
        const sources = [];
        let i = skipWhitespace(s, 0);
        if (s[i] !== '[') return sources;

        i++;
        while (i < s.length) {
            i = skipWhitespace(s, i);
            if (s[i] === ']') break;

            const end = skipValue(s, i);
            sources.push(s[i] === '{' ? memberSource(s, i, key) : undefined);

            i = skipWhitespace(s, end);
            if (s[i] === ',') i++;
        }
        return sources;
    };

    const format = (s, indent) => {
        let out = '';
        let depth = 0;
        const newline = () => indent ? '\n' + indent.repeat(depth) : '';

        for (let i = 0; i < s.length; i++) {
            const c = s[i];
            if (c === '"') {
                const end = skipString(s, i);
                out += s.slice(i, end);
                i = end - 1;
            } else if (c === '{' || c === '[') {
                const next = skipWhitespace(s, i + 1);
                if (s[next] === '}' || s[next] === ']') {
                    out += c + s[next];
                    i = next;
                } else {
                    depth++;
                    out += c + newline();
                }
            } else if (c === '}' || c === ']') {
                depth--;
                out += newline() + c;
            } else if (c === ',') {
                out += ',' + newline();
            } else if (c === ':') {
                out += indent ? ': ' : ':';
            } else if (!WHITESPACE.includes(c)) {
                out += c;
            }
        }
        return out;
    };

    return {
        elementMembers,
        pretty: (s) => format(s, '  '),
        compact: (s) => format(s, ''),
    };
})();

// Must match utils.UsePayloadFromFile in Go
const usesPayloadFromFile = (mock) => {
    if(typeof mock.usePayloadFromFile === 'boolean') {
        return mock.usePayloadFromFile;
    }

    return !!mock.payloadFromFile && [undefined, null, '', 'null'].includes(mock.payload);
}

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
        const errorMsg = "FRONTEND ERROR! Could not parse json. " + typeof data + " (" + escapeHTML(e) + ")";

        console.error(errorMsg);
        
        const logMsg = '<span class="square square-red"></span> ' + errorMsg + '<br />';
        document.getElementById("footer-log").insertAdjacentHTML('afterbegin', logMsg);

        return data;
    }
}