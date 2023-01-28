import React, {useEffect, useState} from 'react'
import {Alert, Button, Card, Space} from 'antd';
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
        setError("")
        WriteSchema(m).then((res) => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            } else {
                setIsModelCopied(false)
                dispatch(setSchema(m))
                dispatch(setModelChangeActivity(!modelChangeTrigger))
            }
        })
    }

    useEffect(() => {
        if (props.initialValue !== '') {
            setModel(props.initialValue)
            save(props.initialValue)
        }
    }, []);

    function copyModel(text) {
        if ('clipboard' in navigator) {
            setIsModelCopied(true)
            return navigator.clipboard.writeText(JSON.stringify(text));
        } else {
            return document.execCommand('copy', true, text);
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
            {error !== "" && <Alert message={error} type="error" showIcon className="mb-12 ml-12 mr-12"/>}
            <Editor setCode={setModel} code={model}></Editor>
        </Card>
    )
}

export default AuthorizationModel
