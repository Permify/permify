import React, {useState} from 'react'
import {Alert, Button, Input, Result, Spin, Typography} from "antd";
import {CheckPermission} from "../../../services/permission";

const {Paragraph} = Typography;

function Check(props) {

    // loading
    const [loading, setLoading] = useState(false);

    // error & validations
    const [error, setError] = useState("");
    const [queryIsValid, setQueryIsValid] = useState(false);

    const isValid = (value) => {
        if (typeof value !== 'string') {
            throw new Error('Input must be a string.');
        }

        let s = value.split(' ');
        if (s.length !== 3) {
            return false;
        }

        const isValidSection = (section) => {
            let splitSection = section.split(':');
            if (splitSection.length !== 2) {
                return false;
            }
            let [, value] = splitSection;
            return value !== '';
        };

        let sIsValid = isValidSection(s[0]);
        let oIsValid = isValidSection(s[2]);

        return sIsValid && oIsValid;
    };

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
        if (typeof value !== 'string') {
            throw new Error('Input must be a string.');
        }

        let parts = value.split(' ');
        if (parts.length !== 3) {
            throw new Error('Invalid input format.');
        }

        let [subjectPart, permissionPart, entityPart] = parts;

        let subjectTokens = subjectPart.split(':');
        if (subjectTokens.length !== 2) {
            throw new Error('Invalid subject format.');
        }

        let [subjectType, subjectId] = subjectTokens;
        let userSet = subjectId.split('#');
        if (userSet.length === 2) {
            setSubject({
                type: subjectType,
                id: userSet[0],
                relation: userSet[1],
            });
        } else {
            setSubject({
                type: subjectType,
                id: subjectId,
            });
        }

        setPermission(permissionPart);

        let entityTokens = entityPart.split(':');
        if (entityTokens.length !== 2) {
            throw new Error('Invalid entity format.');
        }

        let [entityType, entityId] = entityTokens;
        setEntity({
            type: entityType,
            id: entityId,
        });
    };

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
                        placeholder="user:1 edit document:2"
                        className="border-radius-right-none border-right-none"
                        size="large"
                    />
                    <Button style={{
                        width: '15%',
                    }} type="primary" size="large" className="border-radius-left-none" disabled={!queryIsValid}
                            onClick={onCheck}>Check</Button>
                </Input.Group>
                {result != null ?
                    <>
                        {result ?
                            <Result className="pt-20"
                                    status="success"
                                    title={"ALLOWED"}
                            />
                            :
                            <Result className="pt-20"
                                    status="error"
                                    title={"DENIED"}
                            />
                        }
                    </>
                    :
                    <Paragraph className="pt-6 pl-2"><blockquote> simple resource based access check takes form of Can the subject U perform action X on a resource Y ?. A real world example would be: <span className="text-grey">user:1 edit document:2</span> where the right side of the ":" represents identifier of the entity.</blockquote></Paragraph>
                }
            </div>
        </Spin>)
}

export default Check
