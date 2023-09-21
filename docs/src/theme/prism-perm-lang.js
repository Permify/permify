(function (Prism) {
    Prism.languages.perm = Prism.languages.extend('clike', {

        'comment': {
            pattern: /\/\/.*|\/\*[\s\S]*?(?:\*\/|$)/,
            greedy: true
        },
        'number':{
            pattern: /\\B@\\w+#?\\w+/,
            greedy: true
        },
        'keyword': /\b(?:entity|permission|relation|action|or|not|and)\b(?!\s*=\s*\d)/,
    });
}(Prism));