import React, {useEffect} from "react";
import AppRouter from "./router/Router";
import {Layout} from "./utility/context/Layout"
import {Provider} from 'react-redux'
import {ConfigProvider, theme} from 'antd';

function App({store}) {

    const params = new URLSearchParams(window.location.search);

    useEffect(() => {
        let p = params.get('t')
        if (p && p === "f") {
            document.documentElement.style.setProperty('--background-base', "#000000");
        }
    }, []);

    return (
        <ConfigProvider
            theme={{
                token: {
                    colorPrimary: '#6318FF',
                    fontSizeBase: '14px',
                    borderRadius: '2px',
                },
                components: {
                    divider:{
                        backgroundColor: '#DEDBE6',
                    }
                },
                algorithm: theme.darkAlgorithm,
            }}
        >
            <Provider store={store}>
                <Layout>
                    <AppRouter/>
                </Layout>
            </Provider>
        </ConfigProvider>
    );
}

export default App
