import React from "react";
import AppRouter from "./router/Router";
import {Layout} from "./utility/context/Layout"
import {Provider} from 'react-redux'
import {PersistGate} from "redux-persist/integration/react";

function App({store, persistor}) {
    return (
        <Provider store={store}>
            <PersistGate persistor={persistor} loading={null}>
                <Layout>
                    <AppRouter/>
                </Layout>
            </PersistGate>
        </Provider>
    );
}

export default App
