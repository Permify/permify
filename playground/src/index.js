import React from 'react';
import ReactDOM from 'react-dom/client';
import './_assets/less/index.less';
import Main from './main';
import reportWebVitals from './report-web-vitals';
import { LoadWasm } from './wasm';

const rootElement = document.getElementById('root');

// Use createRoot instead of render
const root = ReactDOM.createRoot(rootElement);

root.render(
    <React.StrictMode>
        <LoadWasm>
            <Main />
        </LoadWasm>
    </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
