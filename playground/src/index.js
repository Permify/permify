import React from 'react';
import ReactDOM from 'react-dom';
import './_assets/less/index.less';
import App from './App';
import reportWebVitals from './reportWebVitals';
import store, {persistor} from "./redux/store";
import {LoadWasm} from './loadWasm';

ReactDOM.render(
    <React.StrictMode>
        <LoadWasm>
            <App store={store} />
        </LoadWasm>
    </React.StrictMode>,
    document.getElementById('root')
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
