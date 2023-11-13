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
        try {
            if (value !== props.code) {
                props.setCode(value);
            }
        } catch (error) {
            console.error("Error while editing YAML: ", error);
        }
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
