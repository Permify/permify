import React, {useRef, useState} from 'react';
import MonacoEditor from "@monaco-editor/react";
import Theme from "../perm/theme";

function YamlEditor() {
    const editorRef = useRef(null);
    const monacoRef = useRef(null);

    const [code, setCode] = useState(`checks:
  - entity: "repository:1"
    subject: "user:1"
    context:
    assertions:
      view: true
  - entity: "repository:1"
    subject: "user:1"
    context:
      tuples: []
      attributes: []
      data:
        day_of_week: "saturday"
    assertions:
      view: true
      delete: false
  - entity: "organization:1"
    subject: "user:1"
    context:
    assertions:
      view: true
  entity_filters:
    - entity_type: "repository"
      subject: "user:1"
      context:
      assertions:
        view : ["1"]
  subject_filters:
    - subject_reference: "user"
      entity: "repository:1"
      context:
      assertions:
        view : ["1"]
        edit : ["1"]`);

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
        fontSize: 15,
    };

    return (
        <MonacoEditor
            height="50vh"
            language="yaml"
            theme="dark-theme"
            value={code}
            options={options}
            onChange={newCode => setCode(newCode)}
            beforeMount={handleEditorWillMount}
            onMount={handleEditorDidMount}
        />
    );
}

export default YamlEditor;
