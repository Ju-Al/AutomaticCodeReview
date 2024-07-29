let data;

function getConfig() {
    if (data) return Promise.resolve(data);
    return fetch('config.json?nocache=' + new Date().getUTCMilliseconds()).then(response => {
        data = response.json();
        return data;
    }).catch(error => {
        console.warn('web config file is missing so the template will be used');
        return getDefaultConfig();
    });
}

function getDefaultConfig() {
    return fetch('config.template.json').then(function (response) {
        data = response.json();
        return data;
    } catch (error) {
        console.error('failed to fetch the default web config file:', error);
    }
}

export function getMultiServer() {
    return getConfig().then(config => {
        return config.multiserver;
    }).catch(error => {
        console.log('cannot get web config:', error);
        return false;
    });
}

export function getThemes() {
    return getConfig().then(config => {
        console.warn(config.themes);
        return config.themes;
    }).catch(error => {
        console.log('cannot get web config:', error);
        return [];
    });
}

export function getPlugins() {
    return getConfig().then(config => {
        return config.plugins;
    }).catch(error => {
        console.log('cannot get web config:', error);
        return [];
    });
}
