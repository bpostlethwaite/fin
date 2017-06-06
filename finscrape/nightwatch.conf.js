module.exports = (function(settings) {

    const config = require('./config')();

    if (!config.data.raw_data) {
        throw new Error('config.data.raw_data is a required option.');
    }

    settings
        .test_settings
        .default
        .desiredCapabilities
        .chromeOptions
        .prefs['download.default_directory'] = config.data.raw_data;

    return settings;
}(require('./nightwatch.json')));
