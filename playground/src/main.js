import AppRouter from "@routes";
import {Layout} from "@context/layout"
import {ConfigProvider, theme} from 'antd';
import {Analytics} from '@vercel/analytics/react';
import {SpeedInsights} from "@vercel/speed-insights/react"

function Main() {
    return (
        <ConfigProvider
            theme={{
                token: {
                    colorPrimary: '#6318FF',
                    fontSizeBase: '14px',
                    borderRadius: '2px',
                },
                components: {
                    divider: {
                        backgroundColor: '#DEDBE6',
                    }
                },
                algorithm: theme.darkAlgorithm,
            }}
        >
            <Layout>
                <AppRouter/>
                <Analytics/>
                <SpeedInsights/>
            </Layout>
        </ConfigProvider>
    );
}

export default Main
