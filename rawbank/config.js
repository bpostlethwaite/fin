const toml = require('toml');
const fs = require('fs');
const path = require('path');

module.exports = function() {

    const configPath = process.env.FIN_CONFIG_PATH;

    if (!configPath) {
        throw new Error('FIN_CONFIG_PATH env variable unset or empty');
    }

    const tomlStr = fs.readFileSync(configPath, {encoding: 'utf8'});

    return toml.parse(tomlStr);
};
