import React, {useState} from 'react'
import {Alert, Button, Input, Spin, Typography} from "antd";
import {FilterEntity, FilterSubject} from "../../../services/permission";

const {Paragraph} = Typography;

function SubjectFilter(props) {

    // loading
    const [loading, setLoading] = useState(false);

    // error & validations
    const [error, setError] = useState("");
    const [queryIsValid, setQueryIsValid] = useState(false);

    const isValid = (value) => {
        // Ensure that value is a string to avoid unexpected type errors.
        if (typeof value !== 'string') {
            return false;
        }

        let parts = value.split(" ");

        // Value should have exactly 3 parts separated by space.
        if (parts.length !== 3) {
            return false;
        }

        let subpart = parts[0].split("#");

        // The second part of the string should be split into two by a "#" symbol.
        if (subpart.length === 2) {
            return subpart[1] !== "";
        }

        let entitypart = parts[1].split(":");

        // The second part of the string should also be split into two by a ":" symbol.
        return !(entitypart.length !== 2 || entitypart[0] === "" || entitypart[1] === "");
    }

    const [result, setResult] = useState(null);

    // tuple
    const [entity, setEntity] = useState({});
    const [permission, setPermission] = useState("");
    const [subjectReference, setSubjectReference] = useState({});

    const onQueryChange = (e) => {
        setResult(null)
        setEntity({})
        setSubjectReference({})
        setPermission("")
        setQueryIsValid(false)
        if (isValid(e.target.value.trim())) {
            parseQuery(e.target.value.trim())
            setQueryIsValid(true)
        }
    }

    const onQuery = () => {
        setError("")
        setLoading(true)
        FilterSubject(entity, permission, subjectReference).then(res => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            }
            setResult(res[0])
            setLoading(false)
        })
    }

    const parseQuery = (value) => {
        // Ensure value is a string to avoid type errors.
        if (typeof value !== 'string') {
            throw new Error('Value must be a string');
        }

        // Split the value string into parts.
        const parts = value.split(" ");

        // Ensure the string has exactly 3 parts.
        if (parts.length !== 3) {
            throw new Error('Value must be composed of exactly three parts');
        }

        // Handle the first part (SubjectReference).
        const subjectPart = parts[0].split("#");
        if (subjectPart.length === 2) {
            setSubjectReference({
                type: subjectPart[0],
                relation: subjectPart[1],
            });
        } else if (subjectPart.length === 1) {
            setSubjectReference({
                type: subjectPart[0],
                relation: "",
            });
        } else {
            throw new Error('First part must be one or two terms separated by a "#" symbol');
        }

        // Handle the second part (Entity).
        const entityPart = parts[1].split(":");
        if (entityPart.length !== 2) {
            throw new Error('Second part must be two terms separated by a ":" symbol');
        } else {
            setEntity({
                type: entityPart[0],
                id: entityPart[1],
            });
        }

        // Handle the third part (Permission).
        setPermission(parts[2]);
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
                        placeholder="user document:1 edit"
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
                        {/*user repository:1 edit*/}
                    </>
                    :
                    <Paragraph className="pt-6 pl-2"><blockquote>entity based access check takes form of Which subjects does perform an action X to entity E? This option is best for filtering data or bulk permission checks. A real world example would be: <span className="text-grey font-w-600">user document:1 edit</span></blockquote></Paragraph>
                }
            </div>
        </Spin>
    )
}

export default SubjectFilter
