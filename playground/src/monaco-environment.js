import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import yamlWorker from 'monaco-yaml/yaml.worker?worker';

if (typeof self !== 'undefined') {
  self.MonacoEnvironment = {
    getWorker(_, label) {
      if (label === 'yaml') {
        return new yamlWorker();
      }

      return new editorWorker();
    },
  };
}
