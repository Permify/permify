(function (Prism) {
    Prism.languages.perm = Prism.languages.extend('clike', {
        'class-name': [
            {
                pattern: /(\b(?:entity)\s+)(\w+\/)?\w*(?=\s*\{)/,
                lookbehind: true
            },
            {
                pattern: /(\b(?:action)\s+)\w+(?=\s*\=)/,
                lookbehind: true
            },
            {
                pattern: /(\b(?:relation)\s+)\w+(?=\s*\:)/,
                lookbehind: true
            },
            {
                pattern: /(\b(?:or)\s+)\w+(?=\s*\=)/,
                lookbehind: true
            },
            {
                pattern: /(\b(?:not)\s+)\w+(?=\s*\:)/,
                lookbehind: true
            },
            {
                pattern: /(\b(?:and)\s+)\w+(?=\s*\:)/,
                lookbehind: true
            },
        ],
        'keyword': /\b(?:entity|relation|action|or|not|and)\b(?!\s*=\s*\d)/,
    });
}(Prism));