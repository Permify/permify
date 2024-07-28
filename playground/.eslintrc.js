module.exports = {
    settings: {
        'import/resolver': {
            alias: {
                map: [
                    // And all your import aliases
                    ['@utility', './src/utility'],
                    ['@layout', './src/layout'],
                    ['@router', './src/router'],
                ],
                extensions: ['.ts', '.js', '.jsx', '.json'],
            },
        },
    },
};
