function SyntaxDefinition() {
    return {
        keywords: ["entity", "action", "relation", "and", "or", "not", "permission", "rule", "attribute"],
        escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,
        symbols:  /[=><!~?:&|+\-*\/\^%]+/,
        tokenizer: {
            root: [
                [/\B@\w+#?\w+/, 'reference'],
                [/\b(integer|double|boolean|string)\b/, 'type'],
                [/&&|==|!=|<=|>=|<|>|!|-|%|in|not in/, 'operator'],
                [/`(?:[^`\\]|\\.)*`/, 'option'],
                [/@?[a-zA-Z][\w$]*/, {
                    cases: {
                        "@keywords": 'keyword',
                        "@default": 'variable'
                    }
                }],
                { include: '@whitespace' },
                [/".*?"/, 'string'],
            ],
            comment: [
                [/(\/\/)(.+?)(?=[\n\r]|\*\))/, 'comment'],
            ],
            whitespace: [
                [/[ \t\r\n]+/, 'white'],
                [/\/\*/,       'comment', '@comment' ],
                [/\/\/.*$/,    'comment'],
            ],
        },
    };
}

export default SyntaxDefinition
