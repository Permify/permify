const CracoLessPlugin = require("craco-less");
const {
    getThemeVariables
} = require('antd/dist/theme');

module.exports = {
    plugins: [{
        plugin: CracoLessPlugin,
        options: {
            lessLoaderOptions: {
                lessOptions: {
                    modifyVars: {
                        ...getThemeVariables({
                            dark: true
                        }),
                        '@primary-color': '#6318FF', '@font-size-base': '14px', '@text-color': 'rgba(255, 255, 255, 0.85)',
                    },
                    javascriptEnabled: true
                }
            }
        }
    }],
};
