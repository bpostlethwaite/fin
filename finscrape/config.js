const toml = require('toml');
const fs = require('fs');
const path = require('path');

module.exports = function() {
    try {
        const tomlStr = fs.readFileSync(
            path.join(__dirname, '../config.toml'),
            {encoding: 'utf8'}
        );

        return toml.parse(tomlStr);

    } catch (e) {
        console.error(
            `Parsing error on line ${e.line}, column ${e.column}: ${e.message}`
        );

        process.exit(1);
    }
};
