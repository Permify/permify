import React, {useRef} from 'react';
import MonacoEditor from "@monaco-editor/react";
import { configureMonacoYaml } from 'monaco-yaml';
import 'monaco-editor';

import Theme from "./theme";

function YamlEditor(props) {
    const editorRef = useRef(null);
    const monacoRef = useRef(null);

    const handleEditorDidMount = (editor, monaco) => {
        editorRef.current = editor;
        monacoRef.current = monaco;

        configureMonacoYaml(monaco, {
            completion: true,
            validate: true,
            format: true,
            hover: true,
            enableSchemaRequest: true,
        });
    };

    const handleEditorWillMount = (monaco) => {
        monaco.editor.defineTheme('dark-theme', Theme());
    };

    function handleEditorChange(value, event) {
        try {
            props.setCode(value);
        } catch (error) {
            console.error("Error while editing YAML: ", error);
        }
    }

    const options = {
        selectOnLineNumbers: true,
        renderIndentGuides: true,
        colorDecorators: true,
        cursorBlinking: 'smooth',
        autoClosingQuotes: 'always',
        suggestOnTriggerCharacters: true,
        acceptSuggestionOnEnter: 'on',
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
