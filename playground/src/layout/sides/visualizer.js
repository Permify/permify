import React, {useEffect, useRef, useState} from 'react'
import {Alert, Card} from "antd";
import V from "../../pkg/Visualizer";
import {shallowEqual, useSelector} from "react-redux";
import {ReadSchemaGraph} from "../../services/schema";

function Visualizer(props) {
    const ref = useRef(false);

    const trigger = useSelector((state) => state.common.model_change_toggle, shallowEqual);

    const [error, setError] = useState("");
    const [graph, setGraph] = useState({nodes: [], edges: []});

    const read = () => {
        ReadSchemaGraph().then((res) => {
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
        <Card title={props.title} style={{display: props.hidden && 'none'}}>
            {error !== "" &&
                <Alert message={error} type="error" showIcon className="mb-12"/>
            }
            <div className="spot-background">
                <V graph={graph}></V>
            </div>
        </Card>
    )
}

export default Visualizer;
