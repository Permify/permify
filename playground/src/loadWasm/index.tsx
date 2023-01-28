import './wasm_exec.js';
import './wasmTypes.d.ts';

import React, {useEffect} from 'react';

import SVG from "react-inlinesvg";
import {toAbsoluteUrl} from "../utility/helpers/asset";

async function loadWasm(): Promise<void> {
    const goWasm = new window.Go();
    const result = await WebAssembly.instantiateStreaming(fetch('main.wasm'), goWasm.importObject);
    goWasm.run(result.instance);
}

export const LoadWasm: React.FC<React.PropsWithChildren<{}>> = (props) => {
    const [isLoading, setIsLoading] = React.useState(true);

    useEffect(() => {
        loadWasm().then(() => {
            setIsLoading(false);
        });
    }, []);

    if (isLoading) {
        return (
            <div className="center-of-screen">
                <SVG src={toAbsoluteUrl("/media/svg/rocket.svg")}/>
            </div>
        );
    } else {
        return <React.Fragment>{props.children}</React.Fragment>;
    }
};
