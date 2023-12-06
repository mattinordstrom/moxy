const safelyParseJSON = (data) => {
    if (typeof data === 'number') {
        return data;
    }

    try {
        return JSON.parse(data);
    } catch (e) {
        return data;
    }
}