import React, {useState} from 'react'
import {Alert, Button, Input, Result, Spin, Typography} from "antd";
import {InfoCircleOutlined} from "@ant-design/icons";
import {FilterData} from "../../../services/permission";

const {Paragraph} = Typography;

function DataFiltering(props) {

    // loading
    const [loading, setLoading] = useState(false);

    // error & validations
    const [error, setError] = useState("");
    const [queryIsValid, setQueryIsValid] = useState(false);

    const isValid = (value) => {
        let s = value.split(" ")
        if (s.length !== 3) {
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
        setEntityType(s[0])

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
                        placeholder="document user:1 edit"
                        className="border-radius-right-none border-right-none"
                        size="large"
                    />
                    <Button style={{
                        width: '15%',
                    }} type="primary" size="large" className="border-radius-left-none" disabled={!queryIsValid}
                            onClick={onQuery}>Filter</Button>
                </Input.Group>
                {result != null ?
                    <>
                        <div className="pt-12">
                            <span className="text-grey">Result:</span>
                            <Paragraph className="pt-6 pl-2 text-primary font-w-600">
                                [{result.join(', ')}]
                            </Paragraph>
                        </div>
                        {/*repository user:1 edit*/}
                    </>
                    :
                    <Paragraph className="pt-6 pl-2"><blockquote>subject based access check takes form of Which resources does subject U perform an action X ? This option is best for filtering data or bulk permission checks. A real world example would be: <span className="text-grey font-w-600">document user:1 edit</span></blockquote></Paragraph>
                }
            </div>
        </Spin>)
}

export default DataFiltering
