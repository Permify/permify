import React from 'react'
import {Alert} from "antd";
import V from "@pkg/Visualizer";
import {useShapeStore} from "@state/shape";

function Visualizer() {
    const { graph, visualizerError } = useShapeStore();
    return (
        <>
            {visualizerError !== "" &&
                <Alert message={visualizerError} type="error" showIcon className="mb-12"/>
            }
            <div className="spot-background">
                <V graph={graph}></V>
            </div>
        </>
    )
}

export default Visualizer;
