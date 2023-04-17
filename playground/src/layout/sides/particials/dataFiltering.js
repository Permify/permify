import React, {useState} from 'react'
import {Alert, Button, Input, Result, Spin} from "antd";
import {InfoCircleOutlined} from "@ant-design/icons";
import {FilterData} from "../../../services/permission";

function DataFiltering(props) {

    // loading
    const [loading, setLoading] = useState(false);

    // error & validations
    const [error, setError] = useState("");
    const [queryIsValid, setQueryIsValid] = useState(false);

    const isValid = (value) => {
        let s = value.split(" ")
        if (s.length !== 5) {
            return false
        }

        let sIsValid = false
        let sb = s[2].split(":")
        if (sb.length !== 2) {
            return false
        } else {
            if (sb[0] === "" || sb[1] === "") {
                return false
            } else {
                sIsValid = true
            }
        }

        return sIsValid;
    }

    const onQueryChange = (e) => {
        setResult(null)
        setSubject({})
        setEntityType("")
        setPermission("")
        setQueryIsValid(false)
        if (isValid(e.target.value.trim())) {
            parseQuery(e.target.value.trim())
            setQueryIsValid(true)
        }
    }

    const parseQuery = (value) => {
        let s = value.split(" ")
        setEntityType(s[1])

        let sb = s[2].split(":")
        let userSet = sb[1].split("#")
        if (userSet.length === 2) {
            setSubject({
                type: sb[0],
                id: userSet[0],
                relation: userSet[1],
            })
        } else {
            setSubject({
                type: sb[0],
                id: sb[1],
            })
        }
        setPermission(s[4])
    }

    const [result, setResult] = useState(null);

    // tuple
    const [subject, setSubject] = useState({});
    const [permission, setPermission] = useState("");
    const [entityType, setEntityType] = useState("");

    const onQuery = () => {
        setError("")
        setLoading(true)
        FilterData(entityType,permission, subject ).then(res => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            }
            setResult(res[0])
            setLoading(false)
        })
    }

    return (
        <Spin spinning={loading}>
            <div className="pt-12">
                {error !== "" &&
                    <Alert message={error} type="error" showIcon className="mb-12"/>
                }
                <Input.Group>
                    <Input
                        style={{
                            width: '85%',
                        }}
                        onChange={onQueryChange}
                        placeholder="which repository user:1 can edit"
                        className="border-radius-right-none border-right-none"
                        size="large"
                    />
                    <Button style={{
                        width: '15%',
                    }} type="primary" size="large" className="border-radius-left-none" disabled={!queryIsValid}
                            onClick={onQuery}>Run</Button>
                </Input.Group>
                <span>which <span className="text-grey">entity</span> <span
                    className="text-grey">subject:id</span> <span>can</span> <span
                    className="text-grey">permission (permission or relation)</span></span>
                {result != null ?
                    <>
                        <div className="pt-12">
                            <span className="text-grey">Result:</span>
                            <p className="text-primary font-w-600">
                                [{result.join(', ')}]
                            </p>
                        </div>
                    </>
                    :
                    <Result className="pt-12"
                            icon={<InfoCircleOutlined/>}
                            title="Get filtered response"
                    />
                }
            </div>
        </Spin>)
}

export default DataFiltering
