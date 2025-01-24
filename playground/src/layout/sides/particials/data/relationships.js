import React, {useEffect, useState, useRef} from "react";
import {Button, Table, Input, Form, Space, Select} from "antd";
import {CloseOutlined, DeleteOutlined, MenuOutlined} from "@ant-design/icons";
import {useShapeStore} from "@state/shape";
import {RelationshipObjectToKey, StringRelationshipsToObjects} from "@utility/helpers/common";
import {nanoid} from "nanoid";
import {
    arrayMove,
    SortableContext,
    useSortable,
    verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import {DndContext, useSensor, useSensors, PointerSensor, KeyboardSensor} from '@dnd-kit/core';

const {Option} = Select;

const RelationshipErrorContext = React.createContext([]);

function Relationships() {

    // React hooks for state management and refs
    const [form] = Form.useForm();
    const [dataSource, setDataSource] = useState([]);
    const tableBodyRef = useRef(null);

    // Custom hooks for shape store interactions
    const {
        relationships,
        addRelationships,
        removeRelationship,
        getEntityTypes,
        getRelations,
        getSubjectTypes,
        getSubjectRelations,
        relationshipErrors,
    } = useShapeStore();

    // React hooks for state management of selected entity types, relations, and subject types
    const [selectedEntityTypes, setSelectedEntityTypes] = useState({});
    const [selectedRelations, setSelectedRelations] = useState({});
    const [selectedSubjectTypes, setSelectedSubjectTypes] = useState({});

    const [editingKeys, setEditingKeys] = useState([]);

    // Determine if a given record is currently being edited
    const isEditing = (record) => editingKeys.includes(record.key);

    // Cancel the current editing process
    const cancel = (key) => {
        const newData = dataSource.filter(item => !(item.isNew && item.key === key));
        setDataSource(newData);
        setEditingKeys(prevKeys => prevKeys.filter(k => k !== key));
    };

    // Remove a specific relationship from the data source
    const remove = (relationship) => {
        const updatedDataSource = dataSource.filter(item => item.key !== relationship.key);
        setDataSource(updatedDataSource);
        removeRelationship(relationship.key)
    }

    // Save the current edits
    const save = async (key) => {
        try {
            const currentRowValues = await form.validateFields([
                `entityType_${key}`,
                `entityID_${key}`,
                `relation_${key}`,
                `subjectType_${key}`,
                `subjectID_${key}`,
                `subjectRelation_${key}`
            ]);

            const newData = [...dataSource];
            const index = newData.findIndex((item) => key === item.key);

            const updatedRow = Object.keys(currentRowValues).reduce((acc, currKey) => {
                const newKey = currKey.split('_')[0];
                acc[newKey] = currentRowValues[currKey];
                return acc;
            }, {});

            let updatedKey = RelationshipObjectToKey(updatedRow);
            updatedRow.key = updatedKey;

            // Check if the updated key exists in the data source
            if (newData.some(item => item.key === updatedRow.key)) {
                console.log('This key already exists:', updatedRow.key);
                return; // If the key exists, we exit the function early
            }

            if (index > -1) {
                newData[index] = {...newData[index], ...updatedRow, isNew: false};
                setDataSource(newData);
                setEditingKeys(prevKeys => prevKeys.filter(k => k !== key));
                form.resetFields();
                addRelationships([updatedKey]);
            }
        } catch (errInfo) {
            console.log('Validate Failed:', errInfo);
        }
    };

    // Column configurations for the table
    const columns = [
        {
            key: 'sort',
        },
        {
            title: "Entity Type",
            dataIndex: 'entityType',
            key: 'entityType',
            editable: true,
        },
        {
            title: "Entity ID",
            dataIndex: 'entityID',
            key: 'entityID',
            editable: true,
        },
        {
            title: "Relation",
            dataIndex: 'relation',
            key: 'relation',
            editable: true,
        },
        {
            title: "Subject Type",
            key: 'subjectType',
            dataIndex: 'subjectType',
            editable: true,
        },
        {
            title: "Subject ID",
            key: 'subjectID',
            dataIndex: 'subjectID',
            editable: true,
        },
        {
            title: "Subject Relation",
            key: 'subjectRelation',
            dataIndex: 'subjectRelation',
            editable: true,
        },
        {
            title: '',
            dataIndex: 'operation',
            render: (_, record) => {
                const editable = isEditing(record);
                return editable ? (
                    <span className="flex flex-row items-center gap-2" style={{width: "fit-content"}}>
                        <Button type="primary" onClick={() => save(record.key)}>Save</Button>
                        <Button className="text-white ml-4" type="link" icon={<CloseOutlined/>} onClick={() => cancel(record.key)}/>
                    </span>
                ) : (
                    <Space size="middle">
                        <Button type="text" danger icon={<DeleteOutlined onClick={() => remove(record)}/>}/>
                    </Space>
                );
            },
        },
    ];

    // Merge columns with editing logic
    const mergedColumns = columns.map((col) => {
        if (!col.editable) {
            return col;
        }

        return {
            ...col,
            onCell: (record) => ({
                record,
                inputType: col.dataIndex === 'sort' ? 'number' : 'text',
                dataIndex: col.dataIndex,
                title: col.title,
                editing: isEditing(record),
            }),
        };
    });

    // Handle the end of the drag-and-drop event
    const onDragEnd = ({active, over}) => {
        if (active.id !== over?.id) {
            setDataSource((previous) => {
                const activeIndex = previous.findIndex((i) => i.key === active.id);
                const overIndex = previous.findIndex((i) => i.key === over?.id);
                return arrayMove(previous, activeIndex, overIndex);
            });
        }
    };

    // A table row component for drag-and-drop functionality
    const Row = ({children, ...props}) => {
        const {
            attributes,
            listeners,
            setNodeRef,
            setActivatorNodeRef,
            transform,
            transition,
            isDragging,
        } = useSortable({
            id: props['data-row-key'],
        });

        const relationshipErrors = React.useContext(RelationshipErrorContext);

        // Find if there's an error for this row
        const relationshipError = relationshipErrors.find(
            error => error.key === props['data-row-key']
        );

        const transformStyle = transform
            ? `scaleY(1) translate(${transform.x}px, ${transform.y}px)`
            : '';

        const style = {
            ...props.style,
            transform: transformStyle,
            transition,
            ...(isDragging ? {position: 'relative'} : {}),
        };

        return (
            <tr {...props} ref={setNodeRef} className={relationshipError ? 'error-row' : ''}
                style={style} {...attributes}>
                {React.Children.map(children, (child) => {
                    if (child.key === 'sort') {
                        return React.cloneElement(child, {
                            children: (
                                <MenuOutlined
                                    ref={setActivatorNodeRef}
                                    style={{
                                        touchAction: 'none',
                                        cursor: 'move',
                                    }}
                                    {...listeners}
                                />
                            ),
                        });
                    }
                    return child;
                })}
            </tr>
        );
    };

    // Populate the data source when the component is mounted
    useEffect(() => {
        setDataSource(StringRelationshipsToObjects(relationships))
    }, [relationships]);

    // Scroll to the latest data row after adding a new relationship
    useEffect(() => {
        if (tableBodyRef.current) {
            const scrollableArea = tableBodyRef.current.querySelector('.ant-table-body');
            if (scrollableArea) {
                scrollableArea.scrollTop = scrollableArea.scrollHeight;
            }
        }
    }, [dataSource]);

    // Sensor hooks for drag-and-drop functionality
    const sensors = useSensors(
        useSensor(PointerSensor),
        useSensor(KeyboardSensor)
    );

    // Add a new relationship row for editing
    const addRow = () => {
        const newRow = {
            key: nanoid(),
            entityType: "",
            entityID: "",
            relation: "",
            subjectType: "",
            subjectID: "",
            subjectRelation: "",
            isNew: true,
        };
        setDataSource((prevData) => [...prevData, newRow]);
        setEditingKeys(prevKeys => [...prevKeys, newRow.key]);
    };

    // A cell component for editing
    const EditableCell = ({editing, dataIndex, title, inputType, record, index, children, ...restProps}) => {
        let inputElement;
        if (editing) {
            switch (dataIndex) {
                case 'entityType':
                    inputElement = (
                        <Select
                            showSearch
                            allowClear
                            placeholder="Entity Type"
                            notFoundContent={null}
                            onChange={(value) => handleEntityTypeChange(value, record.key)}
                        >
                            {getEntityTypes().map(option => (
                                <Option key={option} value={option}>
                                    {option}
                                </Option>
                            ))}
                        </Select>
                    );
                    break;
                case 'relation':
                    inputElement = (
                        <Select
                            showSearch
                            allowClear
                            placeholder="Relation"
                            notFoundContent={null}
                            onChange={(value) => handleRelationChange(value, record.key)}
                        >
                            {getRelationOptionsForRow(record.key)}
                        </Select>
                    );
                    break;
                case 'subjectType':
                    inputElement = (
                        <Select
                            showSearch
                            allowClear
                            placeholder="Subject Type"
                            notFoundContent={null}
                            onChange={(value) => handleSubjectTypeChange(value, record.key)}
                        >
                            {getSubjectTypeOptionsForRow(record.key)}
                        </Select>
                    );
                    break;
                case 'subjectRelation':
                    inputElement = (
                        <Select
                            showSearch
                            allowClear
                            placeholder="Subject Relation"
                            notFoundContent={null}
                        >
                            {getSubjectRelationOptionsForRow(record.key)}
                        </Select>
                    );
                    break;
                case 'entityID':
                    inputElement = (
                        <Input placeholder="Entity ID"/>
                    );
                    break;
                case 'subjectID':
                    inputElement = (
                        <Input placeholder="Subject ID"/>
                    );
                    break;
                default:
                    inputElement = inputType === 'number'
                        ? <Input.Number/>
                        : <Input/>;
                    break;
            }

            return (
                <td {...restProps}>
                    <Form.Item
                        name={`${dataIndex}_${record.key}`}
                        style={{margin: 0}}
                        rules={[
                            {
                                required: dataIndex !== "subjectRelation",
                                message: ``
                            }
                        ]}
                    >
                        {inputElement}
                    </Form.Item>
                </td>
            );
        } else {
            return (
                <td {...restProps}>
                    {children}
                </td>
            );
        }
    };

    // Handlers to set selected entity types, relations, and subject types when they are changed

    const handleEntityTypeChange = (value, rowKey) => {
        setSelectedEntityTypes(prev => ({
            ...prev,
            [rowKey]: value
        }));
    };

    const handleRelationChange = (value, rowKey) => {
        setSelectedRelations(prev => ({
            ...prev,
            [rowKey]: value
        }));
    };

    const handleSubjectTypeChange = (value, rowKey) => {
        setSelectedSubjectTypes(prev => ({
            ...prev,
            [rowKey]: value
        }));
    };

    // Utility functions to get options for relation, subject type, and subject relation
    function safelyGetOptions(entityType, selectorFunction, mapFunction) {
        if (!entityType) return [];

        const results = selectorFunction(entityType);
        if (!results || !results.length) return [];

        return results.map(mapFunction);
    }

    function getRelationOptionsForRow(rowKey) {
        const entityType = selectedEntityTypes[rowKey];
        return safelyGetOptions(entityType, getRelations, relation => (
            <Option key={relation} value={relation}>
                {relation}
            </Option>
        ));
    }

    function getSubjectTypeOptionsForRow(rowKey) {
        const entityType = selectedEntityTypes[rowKey];
        const relation = selectedRelations[rowKey];

        return safelyGetOptions(entityType, (entity) => {
            if (!relation) return [];
            return getSubjectTypes(entity, relation);
        }, type => (
            <Option key={type} value={type}>
                {type}
            </Option>
        ));
    }

    function getSubjectRelationOptionsForRow(rowKey) {
        const entityType = selectedEntityTypes[rowKey];
        const relation = selectedRelations[rowKey];
        const subjectType = selectedSubjectTypes[rowKey];

        return safelyGetOptions(entityType, (entity) => {
            if (!relation || !subjectType) return [];
            return getSubjectRelations(entity, relation, subjectType);
        }, type => (
            <Option key={type} value={type}>
                {type}
            </Option>
        ));
    }

    // The main render method
    return (
        <div style={{height: '100vh'}} ref={tableBodyRef}>
            <DndContext sensors={sensors} onDragEnd={onDragEnd}>
                <SortableContext
                    items={dataSource.map((i) => i.key.toString())}
                    strategy={verticalListSortingStrategy}
                >
                    <Form form={form} component={false}>
                        <RelationshipErrorContext.Provider value={relationshipErrors}>
                            <Table
                                components={{
                                    body: {
                                        cell: EditableCell,
                                        row: Row,
                                    },
                                }}
                                rowKey="key"
                                columns={mergedColumns}
                                dataSource={dataSource}
                                pagination={false}
                                footer={() => (
                                    <div style={{textAlign: 'left'}}>
                                        <Button type="primary" onClick={addRow}>
                                            Add Relationship
                                        </Button>
                                    </div>
                                )}
                                scroll={{y: 'calc(100vh - 270px)'}}
                            />
                        </RelationshipErrorContext.Provider>
                    </Form>
                </SortableContext>
            </DndContext>
        </div>
    )
}

export default Relationships
