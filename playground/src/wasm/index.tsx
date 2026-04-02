import React, {useEffect} from 'react';
import './wasm_exec.js';
import './wasm_types.d.ts';
import {toAbsoluteUrl} from "@utility/helpers/asset";
import {Alert} from "antd";

async function loadWasm(): Promise<void> {
    const goWasm = new window.Go();
    const response = await fetch(toAbsoluteUrl('/play.wasm'));
    const fallbackResponse = response.clone();

    if (!response.ok) {
        throw new Error(`Failed to load wasm: ${response.status} ${response.statusText}`);
    }

    let result;
    try {
        result = await WebAssembly.instantiateStreaming(response, goWasm.importObject);
    } catch (error) {
        const bytes = await fallbackResponse.arrayBuffer();
        result = await WebAssembly.instantiate(bytes, goWasm.importObject);
    }

    goWasm.run(result.instance);
}

export const LoadWasm: React.FC<React.PropsWithChildren<{}>> = (props) => {
    const [isLoading, setIsLoading] = React.useState(true);
    const [loadError, setLoadError] = React.useState("");

    useEffect(() => {
        loadWasm()
            .then(() => {
                setIsLoading(false);
            })
            .catch((error) => {
                console.error('Failed to initialize wasm', error);
                setLoadError(error instanceof Error ? error.message : 'Unknown wasm initialization error');
                setIsLoading(false);
            });
    }, []);

    if (loadError) {
        return (
            <div className="wasm-loader-background h-screen">
                <div className="center-of-screen" style={{width: 'min(520px, calc(100vw - 32px))'}}>
                    <Alert
                        type="error"
                        showIcon
                        message="Playground failed to initialize"
                        description={loadError}
                    />
                </div>
            </div>
        );
    }

    if (isLoading) {
        return (
            <div className="wasm-loader-background h-screen">
                <div className="center-of-screen">
                    <img alt="rocket loader" src={toAbsoluteUrl('/media/svg/rocket.svg')}/>
                </div>
            </div>
        );
    } else {
        return <React.Fragment>{props.children}</React.Fragment>;
    }
};
