import React from "react";
import AppRouter from "./router/Router";
import {BrowserRouter} from 'react-router-dom';
import {Layout} from "./utility/context/Layout"
import {Provider} from 'react-redux'
import {PersistGate} from "redux-persist/integration/react";

function App({store, persistor}) {
    return (
        <BrowserRouter>
            <Provider store={store}>
                <PersistGate persistor={persistor} loading={null}>
                    <div className="App">
                        <Layout>
                            <AppRouter/>
                        </Layout>
                    </div>
                </PersistGate>
            </Provider>
        </BrowserRouter>
    );
}

export default App
