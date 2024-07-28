module.exports = {
    settings: {
        'import/resolver': {
            alias: {
                map: [
                    // And all your import aliases
                    ['@utility', './src/utility'],
                    ['@layout', './src/layout'],
                    ['@router', './src/router'],
                    ['@views', './src/views'],
                    ['@state', './src/state'],
                    ['@services', './src/services'],
                    ['@pkg', './src/pkg'],
                ],
                extensions: ['.ts', '.js', '.jsx', '.tsx', '.json'],
            },
        },
    },
};
