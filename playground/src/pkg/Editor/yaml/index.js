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
        fontSize: 13,
    };

    function handleEditorChange(value, event) {
        props.setCode(value);
    }

    return (
        <MonacoEditor
            height="50vh"
            language="yaml"
            theme="dark-theme"
            value={props.code}
            options={options}
            onChange={handleEditorChange}
            beforeMount={handleEditorWillMount}
            onMount={handleEditorDidMount}
        />
    );
}

export default YamlEditor;
