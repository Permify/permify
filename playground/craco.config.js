const CracoLessPlugin = require("craco-less");

const dotenv = require('dotenv');
const webpack = require('webpack');

const env = dotenv.config().parsed;
const envKeys = Object.keys(env).reduce((prev, next) => {
    prev[`process.env.${next}`] = JSON.stringify(env[next]);
    return prev;
}, {});

module.exports = {
    webpack: {
        plugins: [
            new webpack.DefinePlugin(envKeys)
        ]
    },
    plugins: [
        { plugin: CracoLessPlugin }
    ]
};
