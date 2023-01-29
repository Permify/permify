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
                 {/* Load google fonts */}
                <link rel="preconnect" href="https://fonts.googleapis.com" />
                <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="true" />
                <link href="https://fonts.googleapis.com/css2?family=Caveat&family=Rubik&display=swap" rel="stylesheet"/>
            </Head>
            <Landing />
        </>
    );
}

export default Home;
