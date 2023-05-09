import React from "react";
import AppRouter from "./router/Router";
import {Layout} from "./utility/context/Layout"
import {Provider} from 'react-redux'
import {ConfigProvider, theme} from 'antd';

function App({store}) {
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
