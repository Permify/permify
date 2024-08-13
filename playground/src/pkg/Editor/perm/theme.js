function Theme() {
    return {
        base: 'vs-dark',
        inherit: true,
        rules: [
            {token: 'keyword', foreground: 'A274FF', fontStyle: 'bold'},
            {token: 'option', foreground: 'BCE089FF'},
            {token: 'comment', foreground: '21A65F'},
            {token: 'string', foreground: 'F7F3FF'},
            {token: 'variable', foreground: 'F7F3FF'},
            {token: 'reference', foreground: '93F1EE'},
            {token: 'type', foreground: 'FFA500'},
            {token: 'operator', foreground: 'A274FF'},
        ],
        colors: {
            "editor.background": '#141517',
        }
    }
}

export default Theme
