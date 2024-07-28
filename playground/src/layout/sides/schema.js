import React from 'react'
import PermEditor from "@pkg/Editor/perm";
import {useShapeStore} from "@state/shape";

function Schema() {
    const {schema, setSchema, schemaError} = useShapeStore();

    return (
        <PermEditor setCode={setSchema} code={schema} error={schemaError}></PermEditor>
    )
}

export default Schema
