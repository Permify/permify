import React from 'react';
import ReactDOM from 'react-dom/client';
import './_assets/less/index.less';
import './monaco-environment';
import Main from './main';
import reportWebVitals from './report-web-vitals';
import {LoadWasm} from './wasm';

const rootElement = document.getElementById('root');
const root = ReactDOM.createRoot(rootElement);

root.render(
  <React.StrictMode>
    <LoadWasm>
      <Main />
    </LoadWasm>
  </React.StrictMode>
);

reportWebVitals();
