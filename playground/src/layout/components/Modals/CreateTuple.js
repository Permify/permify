import React, {useState} from 'react'
import {Alert, Button, Form, Input, Modal, Select} from "antd";
import {WriteTuple} from "../../../services/relationship";
import {useDispatch} from "react-redux";
import {addRelationships} from "../../../redux/shape/actions";
import {TupleObjectToTupleString} from "../../../utility/helpers/tuple";

const {Option} = Select;

function CreateTuple(props) {

    const definitions = props.model.entityDefinitions

    const [form] = Form.useForm();

    const dispatch = useDispatch();

    const [error, setError] = useState("");

    const [selectedEntityType, setSelectedEntityType] = useState("");
    const [selectedEntityRelation, setSelectedEntityRelation] = useState("");

    const [subjectTypeOptions, setSubjectTypeOptions] = useState([]);
    const [entityRelationOptions, setEntityRelationOptions] = useState([]);
    const [subjectRelationOptions, setSubjectRelationOptions] = useState([]);

    const writeTuple = (tuple) => {
        WriteTuple(tuple).then((res) => {
            if (res[0] != null) {
                setError(res[0])
            } else {
                props.toggle();
                dispatch(addRelationships(TupleObjectToTupleString(tuple)))
                form.resetFields();
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

    const handleEntityTypeChange = (value) => {
        if (!definitions[value]) {
            console.error("Selected entity type not found in definitions.");
            return;
        }
        clearInputs();
        setSelectedEntityType(value);
        if (definitions[value].relations){
            const relationKeys = Object.keys(definitions[value].relations);
            if (!relationKeys || !Array.isArray(relationKeys)) {
                console.error("Invalid relation keys.");
                return;
            }
            setEntityRelationOptions(relationKeys);
        }
        form.setFieldsValue({
            entity_relation: null,
            subject_type: null,
            subject_relation: null,
        });
    };

    const handleEntityRelationChange = (value) => {
        if (!definitions[selectedEntityType] || !definitions[selectedEntityType].relations[value]) {
            console.error("Selected entity type or relation not found in definitions.");
            return;
        }
        const relationReferences = definitions[selectedEntityType].relations[value].relationReferences;
        if (!relationReferences || !Array.isArray(relationReferences)) {
            console.error("Invalid relationReferences value.");
            return;
        }
        setSelectedEntityRelation(value);
        setSubjectTypeOptions([]);
        setSubjectRelationOptions([]);
        const so = relationReferences
            .filter(v => v && v.type)
            .reduce((uniqueTypes, v) => {
                if (!uniqueTypes.includes(v.type)) {
                    uniqueTypes.push(v.type);
                }
                return uniqueTypes;
            }, []);
        setSubjectTypeOptions(so);
        form.setFieldsValue({
            subject_type: null,
            subject_relation: null,
        });
    };

    const handleSubjectTypeChange = (value) => {
        if (!definitions[selectedEntityType] || !definitions[selectedEntityType].relations[selectedEntityRelation]) {
            console.error("Selected entity type or relation not found in definitions.");
            return;
        }
        const relationReferences = definitions[selectedEntityType].relations[selectedEntityRelation].relationReferences;
        if (!relationReferences || !Array.isArray(relationReferences)) {
            console.error("Invalid relationReferences value.");
            return;
        }
        const sr = relationReferences
            .filter(v => v && v.relation && v.type === value)
            .map(v => v.relation);
        setSubjectRelationOptions(sr);
        form.setFieldsValue({
            subject_relation: null,
        });
    };

    const clearInputs = () => {
        setEntityRelationOptions([])
        setSubjectTypeOptions([])
        setSubjectRelationOptions([])
    }

    return (
        <Modal title="Create Tuple" visible={props.visible} onCancel={handleCancel} destroyOnClose
               bordered={true} footer={[
            <Button type="secondary" onClick={handleCancel} key="cancel">
                Cancel
            </Button>,
            <Button form="tuple-creation-form" type="primary" key="submit" htmlType="submit">
                CreateTuple
            </Button>
        ]}>
            <Form name="tuple-creation-form" form={form} onFinish={onFinish} labelCol={{span: 4}}
                  wrapperCol={{span: 20}} className="mt-36">
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
                                {definitions && Object.keys(definitions).map((key, index) => {
                                    if (typeof key !== "string") {
                                        console.error("Invalid key: ", key);
                                        return null;
                                    }
                                    return (
                                        <Option key={key} value={key}>
                                            {key}
                                        </Option>
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
                                {entityRelationOptions &&
                                    <>
                                        {
                                            Array.isArray(entityRelationOptions) &&
                                            entityRelationOptions.map((key) => (
                                                <Option key={key} value={key}>
                                                    {key}
                                                </Option>
                                            ))
                                        }
                                    </>
                                }
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
                                {subjectTypeOptions &&
                                    <>
                                        {
                                            Array.isArray(subjectTypeOptions) &&
                                            subjectTypeOptions.map(key => (
                                                <Option key={key} value={key}>
                                                    {key}
                                                </Option>
                                            ))
                                        }
                                    </>
                                }
                            </Select>
                        </Form.Item>

                        <Form.Item
                            name="subject_id"
                            noStyle
                            rules={[{required: true, message: ''}]}
                        >
                            <Input style={{width: '20%'}} placeholder="ID"/>
                        </Form.Item>

                        {subjectRelationOptions.length > 0 &&
                            <Form.Item
                                name="optional_subject_relation"
                                noStyle
                                rules={[{required: true, message: ''}]}
                            >
                                <Select placeholder="Relation" style={{width: '35%'}}>
                                    {subjectRelationOptions &&
                                        <>
                                            {
                                                Array.isArray(subjectRelationOptions) &&
                                                subjectRelationOptions.map(item => (
                                                    <Option key={item} value={item}>
                                                        {item}
                                                    </Option>
                                                ))
                                            }
                                        </>
                                    }
                                </Select>
                            </Form.Item>
                        }
                    </Input.Group>
                </Form.Item>
            </Form>
        </Modal>
    )
}

export default CreateTuple
