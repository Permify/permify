import React, {useRef} from 'react';
import MonacoEditor from "@monaco-editor/react";
import Theme from "../perm/theme";

function YamlEditor(props) {
    const editorRef = useRef(null);
    const monacoRef = useRef(null);

    function handleEditorDidMount(editor, monaco) {
        editorRef.current = editor;
        monacoRef.current = monaco;
    }

    function handleEditorWillMount(monaco) {
        monaco.editor.defineTheme('dark-theme', Theme())
    }

    function handleEditorChange(value, event) {
        if (value !== props.code) {
            props.setCode(value);
        }
    }

    const options = {
        selectOnLineNumbers: false,
        renderIndentGuides: false,
        colorDecorators: false,
        cursorBlinking: "smooth",
        autoClosingQuotes: "always",
        suggestOnTriggerCharacters: false,
        acceptSuggestionOnEnter: "on",
        folding: false,
        lineNumbersMinChars: 3,
        fontSize: 12,
    };

    return (
        <MonacoEditor
            height="100vh"
            language="yaml"
            theme="dark-theme"
            value={props.code}
            options={options}
            beforeMount={handleEditorWillMount}
            onMount={handleEditorDidMount}
            onChange={handleEditorChange}
        />
    );
}

export default YamlEditor;
