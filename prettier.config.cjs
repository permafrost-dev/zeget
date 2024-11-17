const overrides = [
    {
        files: ['*.yml', '*.yaml'],
        options: {
            tabWidth: 2,
        },
    },
    {
        files: '*.json',
        options: {
            parser: 'json',
            tabWidth: 4,
            printWidth: 160,
        },
    },
];

module.exports = {
    arrowParens: 'avoid',
    bracketSameLine: true,
    bracketSpacing: true,
    htmlWhitespaceSensitivity: 'css',
    insertPragma: false,
    jsxSingleQuote: false,
    overrides,
    printWidth: 160,
    proseWrap: 'preserve',
    quoteProps: 'as-needed',
    requirePragma: false,
    semi: true,
    singleQuote: true,
    tabWidth: 4,
    trailingComma: 'all',
    useTabs: false,
};
