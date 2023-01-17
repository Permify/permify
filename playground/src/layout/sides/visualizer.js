import React, {useEffect, useRef, useState} from 'react'
import {Alert, Card} from "antd";
import V from "../../pkg/Visualizer";
import {shallowEqual, useSelector} from "react-redux";

function Visualizer(props) {
    const ref = useRef(false);

    const trigger = useSelector((state) => state.common.model_change_toggle, shallowEqual);

    const [error, setError] = useState("");
    const [graph, setGraph] = useState({nodes: [], edges: []});

    const ReadSchemaCall = () => {
        return new Promise((resolve) => {
            let res = window.readSchemaGraph("")
            resolve(res);
        });
    }

    const read = () => {
        ReadSchemaCall().then((res) => {
            if (res[1] != null) {
                setError(res[1])
            }
            setGraph(JSON.parse(res[0]))
        })
    }

    useEffect(() => {
        if (ref.current) {
            read()
        }
        ref.current = true;
    }, [trigger]);

    return (
        <Card title="Visualizer" className="ml-12">
            {error !== "" &&
                <Alert message={error} type="error" showIcon className="mb-12"/>
            }
            <div className="visualizer-background">
                <V graph={graph}></V>
            </div>
        </Card>
    )
}

export default Visualizer;
