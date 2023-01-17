import React, {useRef} from 'react'
import MonacoEditor from "@monaco-editor/react";
import Theme from "./theme";
import SyntaxDefinition from "./syntax-definition";

function Editor(props) {

    const editorRef = useRef(null);

    function handleEditorDidMount(editor, monaco) {
        editorRef.current = editor;
    }
    function handleEditorWillMount(monaco) {
        monaco.languages.typescript.javascriptDefaults.setEagerModelSync(true);
        monaco.languages.register({id: 'perm'});
        let syntax = SyntaxDefinition()
        monaco.languages.setMonarchTokensProvider('perm', syntax)
        monaco.editor.defineTheme('perm-theme', Theme())
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
       props.setCode(value)
    }

    return (
        <MonacoEditor
            height="100vh"
            defaultLanguage="perm"
            theme="perm-theme"
            value={props.code}
            beforeMount={handleEditorWillMount}
            onMount={handleEditorDidMount}
            onChange={handleEditorChange}
        />
    )
}

export default Editor
