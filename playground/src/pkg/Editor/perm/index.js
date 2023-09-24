import React, {useEffect, useRef} from 'react'
import MonacoEditor from "@monaco-editor/react";
import Theme from "./theme";
import SyntaxDefinition from "./syntax-definition";

function PermEditor(props) {

    const editorRef = useRef(null);
    const monacoRef = useRef(null);

    function handleEditorDidMount(editor, monaco) {
        editorRef.current = editor;
        monacoRef.current = monaco;
    }

    function handleEditorWillMount(monaco) {
        monaco.languages.typescript.javascriptDefaults.setEagerModelSync(true);
        monaco.languages.register({id: 'perm'});
        let syntax = SyntaxDefinition()
        monaco.languages.setMonarchTokensProvider('perm', syntax)
        monaco.editor.defineTheme('perm-theme', Theme())
        monaco.languages.setLanguageConfiguration('perm', {
            comments: {
                lineComment: '//',
                blockComment: ['/*', '*/'],
            },
            brackets: [
                ['{', '}'],
                ['(', ')'],
            ],
            autoClosingPairs: [
                { open: '{', close: '}' },
                { open: '(', close: ')' },
            ],
            surroundingPairs: [
                { open: '{', close: '}' },
                { open: '(', close: ')' },
            ],
        });
        monaco.languages.registerCompletionItemProvider('perm', {
            provideCompletionItems(model, position) {
                const suggestions = [
                    ...syntax.keywords.map(k => {
                        return {
                            label: k,
                            kind: monaco.languages.CompletionItemKind.Keyword,
                            insertText: k,
                        }
                    }),
                ];
                return {suggestions: suggestions};
            }
        })
    }

    function handleEditorChange(value, event) {
        props.setCode(value);
    }

    useEffect(() => {
        if (editorRef.current !== null && monacoRef.current !== null){
            const model = editorRef.current.getModel();
            if (props.error !== null) {
                updateMarkers(editorRef.current, model, [props.error.schemaError], monacoRef.current);
            }else{
                updateMarkers(editorRef.current, model, [], monacoRef.current);
            }
        }
    }, [props.error]);

    function updateMarkers(editor, model, errors, monaco) {
        const markers = errors.map((error) => ({
            severity: monaco.MarkerSeverity.Error,
            message: error.message,
            startLineNumber: error.line,
            startColumn: error.column,
            endLineNumber: error.line,
            endColumn: error.column + 5,
        }));
        monaco.editor.setModelMarkers(model, "perm", markers);
    }

    const options = {
        selectOnLineNumbers: true,
        renderIndentGuides: true,
        colorDecorators: true,
        cursorBlinking: "smooth",
        autoClosingQuotes: "always",
        suggestOnTriggerCharacters: true,
        acceptSuggestionOnEnter: "on",
        folding: true,
        lineNumbersMinChars: 3,
        fontSize: 12,
    };

    return (
        <MonacoEditor
            height="100vh"
            defaultLanguage="perm"
            theme="perm-theme"
            value={props.code}
            options={options}
            beforeMount={handleEditorWillMount}
            onMount={handleEditorDidMount}
            onChange={handleEditorChange}
        />
    )
}

export default PermEditor
