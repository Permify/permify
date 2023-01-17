import { useEffect, useState } from 'react';
import  {AsBind} from "as-bind";

export const useWasm = () => {
    const [state, setState] = useState(null);
    useEffect(() => {
        const fetchWasm = async () => {
            const wasm = await fetch('main.wasm');
            const ins = await AsBind.instantiate(wasm, {});
            setState(ins);
        };
        fetchWasm();
    }, []);
    return state;
}
