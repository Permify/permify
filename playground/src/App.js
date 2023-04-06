import React, {useEffect} from "react";
import AppRouter from "./router/Router";
import {Layout} from "./utility/context/Layout"
import {Provider} from 'react-redux'

function App({store}) {

    useEffect(
        () => {
            localStorage.clear()
        }
    )

    return (
        <Provider store={store}>
            <Layout>
                <AppRouter/>
            </Layout>
        </Provider>
    );
}

export default App
