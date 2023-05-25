import React, {useState} from 'react'
import {Alert, Button, Input, Spin, Typography} from "antd";
import {FilterEntity} from "../../../services/permission";

const {Paragraph} = Typography;

function EntityFilter(props) {

    // loading
    const [loading, setLoading] = useState(false);

    // error & validations
    const [error, setError] = useState("");
    const [queryIsValid, setQueryIsValid] = useState(false);

    const isValid = (value) => {
        // Check if value is a string
        if (typeof value !== 'string') {
            return false;
        }

        // Split value by space
        const parts = value.split(" ");

        // Ensure there are exactly three parts
        if (parts.length !== 3) {
            return false;
        }

        // Split the second part by colon
        const subParts = parts[1].split(":");

        // Ensure there are exactly two sub-parts and neither are empty strings
        if (subParts.length !== 2 || subParts[0] === "" || subParts[1] === "") {
            return false;
        }

        // If all checks pass, the input is valid
        return true;
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
        // Ensure that value is a string to avoid unexpected type errors.
        if (typeof value !== 'string') {
            throw new Error('Value must be a string');
        }

        // Split the value string into parts.
        const parts = value.split(" ");

        // Ensure the string has exactly 3 parts.
        if (parts.length !== 3) {
            throw new Error('Value must be composed of exactly three parts');
        }

        // Set the entity type using the first part.
        setEntityType(parts[0]);

        // Split the second part on ":"
        const subjectpart = parts[1].split(":");

        // Ensure the second part has exactly 2 parts when split on ":".
        if (subjectpart.length !== 2) {
            throw new Error('Second part must be two terms separated by a ":" symbol');
        }

        // Split the user set on "#"
        const userset = subjectpart[1].split("#");
        if (userset.length === 2) {
            setSubject({
                type: subjectpart[0],
                id: userset[0],
                relation: userset[1],
            });
        } else if (userset.length === 1) {
            setSubject({
                type: subjectpart[0],
                id: subjectpart[1],
            });
        } else {
            throw new Error('Invalid format for user set');
        }

        // Set the permission using the third part.
        setPermission(parts[2]);
    }

    const [result, setResult] = useState(null);

    // tuple
    const [subject, setSubject] = useState({});
    const [permission, setPermission] = useState("");
    const [entityType, setEntityType] = useState("");

    const onQuery = () => {
        setError("")
        setLoading(true)
        FilterEntity(entityType, permission, subject).then(res => {
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

export default EntityFilter
