import React, {useEffect, useState} from 'react'
import {Button, Card, Space} from 'antd';
import {CopyOutlined, SaveOutlined} from "@ant-design/icons";
import Editor from "../../pkg/Editor";
import {shallowEqual, useDispatch, useSelector} from "react-redux";
import {setModelChangeActivity} from "../../redux/common/actions";
import {WriteSchema} from "../../services/schema";
import {setSchema} from "../../redux/shape/actions";

function AuthorizationModel(props) {

    const modelChangeTrigger = useSelector((state) => state.common.model_change_toggle, shallowEqual);

    const dispatch = useDispatch();

    const [error, setError] = useState("");
    const [model, setModel] = useState(``);
    const [isModelCopied, setIsModelCopied] = useState(false);

    const save = (m) => {
        props.isReady(false)
        setError(null)
        WriteSchema(m).then((res) => {
            if (res[1] != null) {
                let numbers = parseNumbers(res[1])
                setError({
                    line: numbers[0],
                    column: numbers[1],
                    message: res[1].replaceAll('_', ' ').toLowerCase(),
                })
            } else {
                setIsModelCopied(false)
                dispatch(setSchema(m))
                dispatch(setModelChangeActivity(!modelChangeTrigger))
                props.isReady(true)
            }
        })
    }

    useEffect(() => {
        if (props.initialValue !== '') {
            setModel(props.initialValue)
            save(props.initialValue)
        }
    }, [props.initialValue]);

    function copyModel(text) {
        if ('clipboard' in navigator) {
            setIsModelCopied(true)
            return navigator.clipboard.writeText(JSON.stringify(text));
        } else {
            return document.execCommand('copy', true, text);
        }
    }

    function parseNumbers(input) {
        const regex = /^(\d+):(\d+)/;
        const match = regex.exec(input);
        if (match) {
            const num1 = parseInt(match[1], 10);
            const num2 = parseInt(match[2], 10);
            return [num1, num2]
        } else {
            return [0, 0]
        }
    }

    return (
        <Card title={props.title} extra={<Space>

            <Button type="secondary" onClick={() => {
                copyModel(model)
            }} icon={<CopyOutlined/>}>{isModelCopied ? 'Copied!' : 'Copy'}</Button>

            <Button type="primary" onClick={() => {
                save(model)
            }} icon={<SaveOutlined/>}>Save</Button>

        </Space>} style={{display: props.hidden && 'none'}}>
            <Editor setCode={setModel} code={model} error={error}></Editor>
        </Card>
    )
}

export default AuthorizationModel
