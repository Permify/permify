import React, {useState} from 'react'
import {Alert, Button, Input, Result, Spin} from "antd";
import {InfoCircleOutlined} from "@ant-design/icons";
import {CheckPermission} from "../../../services/permission";

function Check(props) {

    // loading
    const [loading, setLoading] = useState(false);

    // error & validations
    const [error, setError] = useState("");
    const [queryIsValid, setQueryIsValid] = useState(false);

    const isValid = (value) => {
        let s = value.split(" ")
        if (s.length !== 4) {
            return false
        }

        let sIsValid = false
        let sb = s[1].split(":")
        if (sb.length !== 2) {
            return false
        } else {
            if (sb[0] === "" || sb[1] === "") {
                return false
            } else {
                sIsValid = true
            }
        }

        let oIsValid = false
        let ent = s[3].split(":")
        if (ent.length !== 2) {
            return false
        } else {
            if (ent[0] === "" || ent[1] === "") {
                return false
            } else {
                oIsValid = true
            }
        }

        return sIsValid && oIsValid;
    }

    const onQueryChange = (e) => {
        setResult(null)
        setSubject({})
        setEntity({})
        setPermission("")
        setQueryIsValid(false)
        if (isValid(e.target.value.trim())) {
            parseQuery(e.target.value.trim())
            setQueryIsValid(true)
        }
    }

    const parseQuery = (value) => {
        let s = value.split(" ")
        let sb = s[1].split(":")
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
        setPermission(s[2])
        let ent = s[3].split(":")
        setEntity({
            type: ent[0],
            id: ent[1],
        })
    }

    const [result, setResult] = useState(null);

    // tuple
    const [subject, setSubject] = useState({});
    const [permission, setPermission] = useState("");
    const [entity, setEntity] = useState({});

    const onCheck = () => {
        setError("")
        setLoading(true)
        CheckPermission(entity, permission, subject).then(res => {
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
                        placeholder="can user:1 push repository:1"
                        className="border-radius-right-none border-right-none"
                        size="large"
                    />
                    <Button style={{
                        width: '15%',
                    }} type="primary" size="large" className="border-radius-left-none" disabled={!queryIsValid}
                            onClick={onCheck}>Check</Button>
                </Input.Group>
                <span>can <span className="text-grey">subject:id</span> <span className="text-grey">permission (action or relation)</span> <span
                    className="text-grey">entity:id</span></span>
                {result != null ?
                    <>
                        {result ?
                            <Result className="pt-12"
                                    status="success"
                                    title={subject.relation ?
                                        <>
                                            {subject.type}:{subject.id}#{subject.relation} can {permission} {entity.type}:{entity.id}
                                        </>
                                        :
                                        <>
                                            {subject.type}:{subject.id} can {permission} {entity.type}:{entity.id}
                                        </>
                                    }
                            />
                            :
                            <Result className="pt-12"
                                    status="error"
                                    title={subject.relation ?
                                        <>
                                            {subject.type}:{subject.id}#{subject.relation} can
                                            not {permission} {entity.type}:{entity.id}
                                        </>
                                        :
                                        <>
                                            {subject.type}:{subject.id} can not {permission} {entity.type}:{entity.id}
                                        </>
                                    }
                            />
                        }
                    </>
                    :
                    <Result className="pt-12"
                            icon={<InfoCircleOutlined/>}
                            title="Perform your access checks"
                    />
                }
            </div>
        </Spin>)
}

export default Check
