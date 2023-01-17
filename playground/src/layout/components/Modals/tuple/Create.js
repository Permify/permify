import React, {useState} from 'react'
import {Modal, Form, Button, Select, Input, Alert} from "antd";

const {Option} = Select;

function Create(props) {

    const [error, setError] = useState("");

    const [isSubjectUser, setIsSubjectUser] = useState(false);

    const [selectedEntity, setSelectedEntity] = useState({});
    const [subjectOptions, setSubjectOptions] = useState([]);
    const [entityRelationOptions, setEntityRelationOptions] = useState([]);
    const [subjectRelationOptions, setSubjectRelationOptions] = useState([]);

    const [form] = Form.useForm();

    const WriteTupleCall = (tuple) => {
        return new Promise((resolve) => {
            let res = window.writeTuple(JSON.stringify(tuple), "")
            resolve(res);
        });
    }

    const writeTuple = (tuple) => {
        WriteTupleCall(tuple).then((res) => {
            if (res[0] != null) {
                setError(res[0])
            } else {
                props.toggle();
                setIsSubjectUser(false)
                form.resetFields();
                props.setFilter(tuple.entity.type)
            }
        })
    }

    const onFinish = (values) => {
        writeTuple({
            entity: {
                type: values.entity_type,
                id: values.entity_id,
            },
            relation: values.entity_relation,
            subject: {
                type: values.subject_type,
                id: values.subject_id,
                relation: values.optional_subject_relation === undefined ? "" : values.optional_subject_relation
            }
        })
    };

    const handleCancel = () => {
        props.toggle();
        form.resetFields();
    };

    const handleSubjectTypeChange = (value) => {
        setIsSubjectUser(false)
        if (value === "user") {
            setIsSubjectUser(true)
        }
    }

    const clearInputs = () => {
        setSelectedEntity({})
        setEntityRelationOptions([])
        setSubjectOptions([])
        setSubjectRelationOptions([])
    }

    const handleEntityTypeChange = (value) => {
        clearInputs()
        for (const [_, definition] of Object.entries(props.model.entityDefinitions)) {
            if (definition.name === value) {
                setSelectedEntity(definition)
                setEntityRelationOptions(Object.keys(definition.relations))
            }
        }
    }

    const handleEntityRelationChange = (value) => {
        setSubjectOptions([])
        setSubjectRelationOptions([])
        let so = []
        let sr = []
        for (const [_, v] of Object.entries(selectedEntity.relations[value].relationReferences)) {
            let e = v.name.split("#")
            so.push(e[0])
            if (e.length === 2) {
                sr.push(e[1])
            } else {
                if (e[0] !== "user") {
                    sr.push('...')
                }
            }
        }
        setSubjectOptions(so)
        setSubjectRelationOptions(sr)
    }

    return (
        <Modal title="Create Tuple" visible={props.visible} onCancel={handleCancel} destroyOnClose
               bordered={true} footer={[
            <Button type="secondary" onClick={handleCancel}>
                Cancel
            </Button>,
            <Button form="tuple-creation-form" type="primary" key="submit" htmlType="submit">
                Create
            </Button>
        ]}>
            <Form name="tuple-creation-form" form={form} onFinish={onFinish} labelCol={{span: 4}} wrapperCol={{span: 20}}>
                {error !== "" &&
                    <Alert message={error} type="error" showIcon className="mb-12"/>
                }
                <Form.Item label="Entity">
                    <Input.Group compact>
                        <Form.Item
                            name="entity_type"
                            noStyle
                            rules={[{required: true, message: ''}]}
                        >
                            <Select placeholder="Entity Type" style={{width: '35%'}} onChange={handleEntityTypeChange}>
                                {Object.keys(props.model.entityDefinitions).map((key, index) => {
                                    return (
                                        <Option key={key} value={key}>{key}</Option>
                                    );
                                })}
                            </Select>
                        </Form.Item>

                        <Form.Item
                            name="entity_id"
                            noStyle
                            rules={[{required: true, message: ''}]}
                        >
                            <Input style={{width: '20%'}} placeholder="ID"/>
                        </Form.Item>

                        <Form.Item
                            name="entity_relation"
                            noStyle
                            rules={[{required: true, message: ''}]}
                        >
                            <Select placeholder="Relation" style={{width: '35%'}} onChange={handleEntityRelationChange}>
                                {entityRelationOptions.map(key => (
                                    <Option key={key} value={key}>{key}</Option>
                                ))}
                            </Select>
                        </Form.Item>

                    </Input.Group>
                </Form.Item>

                <Form.Item label="Subject">
                    <Input.Group compact>
                        <Form.Item
                            name="subject_type"
                            noStyle
                            rules={[{required: true, message: ''}]}
                        >
                            <Select placeholder="Subject Type" onChange={handleSubjectTypeChange}
                                    style={{width: '35%'}}>
                                {subjectOptions.map(item => (
                                    <Option key={item} value={item}>{item}</Option>
                                ))}
                            </Select>
                        </Form.Item>

                        <Form.Item
                            name="subject_id"
                            noStyle
                            rules={[{required: true, message: ''}]}
                        >
                            <Input style={{width: '20%'}} placeholder="ID"/>
                        </Form.Item>

                        {!isSubjectUser &&
                            <Form.Item
                                name="optional_subject_relation"
                                noStyle
                                rules={[{required: true, message: ''}]}
                            >
                                <Select placeholder="Relation" style={{width: '35%'}}>
                                    {subjectRelationOptions.map(item => (
                                        <Option key={item} value={item}>{item}</Option>
                                    ))}
                                </Select>
                            </Form.Item>
                        }
                    </Input.Group>
                </Form.Item>
            </Form>
        </Modal>
    )
}

export default Create
