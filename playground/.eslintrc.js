module.exports = {
    settings: {
        'import/resolver': {
            alias: {
                map: [
                    ['@utility', './src/utility'],
                    ['@layout', './src/layout'],
                    ['@components', './src/components'],
                    ['@context', './src/context'],
                    ['@features', './src/features'],
                    ['@routes', './src/routes'],
                    ['@pages', './src/pages'],
                    ['@state', './src/state'],
                    ['@services', './src/services'],
                    ['@lib', './src/lib'],
                ],
                extensions: ['.ts', '.js', '.jsx', '.tsx', '.json'],
            },
        },
    },
};
