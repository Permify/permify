import React from "react";
import {Button, Form, Input, Modal} from "antd";
import {useShapeStore} from "@state/shape";

function NewScenario(props) {
    const [form] = Form.useForm();

    const { scenarios, setScenarios } = useShapeStore();

    const handleCreate = async () => {
        try {
            const values = await form.validateFields();

            // Add the new scenario
            const newScenario = {
                name: values.name,
                description: values.description,
                checks: [
                    {
                        entity: null,
                        subject: null,
                        context: null,
                        assertions: {}
                    }
                ],
                entity_filters: [],
                subject_filters: []
            };

            // Update the state
            setScenarios([...scenarios, newScenario]);

            // Reset form fields and close the modal
            form.resetFields();
            props.toggle();

        } catch (errorInfo) {
            console.log('Validation failed:', errorInfo);
        }
    };

    const handleCancel = () => {
        props.toggle();
        form.resetFields();
    };

    return (
        <Modal
            title="New Scenario"
            open={props.visible}
            onOk={handleCancel}
            onCancel={handleCancel}
            destroyOnClose
            bordered={true}
            footer={null}
        >
            <Form form={form} layout="vertical">
                <Form.Item
                    name="name"
                    label="Name"
                    rules={[{ required: true, message: 'Please input a name!' }]}
                >
                    <Input placeholder="Enter name" />
                </Form.Item>

                <Form.Item
                    name="description"
                    label="Description"
                    rules={[{ required: true, message: 'Please input a description!' }]}
                >
                    <Input placeholder="Enter description" />
                </Form.Item>

                <Form.Item>
                    <Button type="primary" htmlType="submit" onClick={handleCreate}>
                        Create
                    </Button>
                </Form.Item>
            </Form>
        </Modal>
    );
}

export default NewScenario
