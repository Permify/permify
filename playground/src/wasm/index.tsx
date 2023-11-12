import React, {useEffect} from 'react';
import {toAbsoluteUrl} from "../utility/helpers/asset";
import './wasm_exec.js';
import './wasm_types.d.ts';
async function loadWasm(): Promise<void> {
    const goWasm = new window.Go();
    const result = await WebAssembly.instantiateStreaming(fetch('play.wasm'), goWasm.importObject);
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
            <div className="wasm-loader-background h-screen">
                <div className="center-of-screen">
                    <img alt="rocket loader" src={toAbsoluteUrl("/media/svg/rocket.svg")}/>
                </div>
            </div>
        );
    } else {
        return <React.Fragment>{props.children}</React.Fragment>;
    }
};
