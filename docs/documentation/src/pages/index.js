import React from "react";
import Layout from "@theme/Layout";
import Head from "@docusaurus/Head";

import { Landing } from "../components/landing";

function Home() {
    React.useEffect(() => {
        return () => {
            // scroll to top after unmount with set timeout
            setTimeout(() => {
                window.scrollTo(0, 0);
            }, 0);
        };
    }, []);

    return (
        <>
            <Head>
                <html data-page="index" data-customized="true" />
                <title>Permify | Open Source Authorization Service Based on Google Zanzibar</title>
                <meta name="Description" content="Open Source Authorization Service Based on Google Zanzibar" />
            </Head>
            <Landing />
        </>
    );
}

export default Home;
