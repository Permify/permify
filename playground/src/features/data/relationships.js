import React, {useEffect, useState, useRef, useMemo, useCallback} from "react";
import {Button, Table, Input, Form, Space, Select, message} from "antd";
import {CloseOutlined, DeleteOutlined, MenuOutlined, EditOutlined} from "@ant-design/icons";
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
    const [errorMessage, setErrorMessage] = useState('');
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
    const cancel = useCallback((key) => {
        const newData = dataSource.filter(item => !(item.isNew && item.key === key));
        setDataSource(newData);
        setEditingKeys(prevKeys => prevKeys.filter(k => k !== key));
        setErrorMessage('');
    }, [dataSource]);

    // Remove a specific relationship from the data source
    const remove = useCallback((relationship) => {
        const updatedDataSource = dataSource.filter(item => item.key !== relationship.key);
        setDataSource(updatedDataSource);
        removeRelationship(relationship.key);
        setErrorMessage('');
    }, [dataSource, removeRelationship]);

    // Save the current edits
    const save = useCallback(async (key) => {
        try {
            setErrorMessage('');
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
            const originalItem = newData[index];

            const updatedRow = Object.keys(currentRowValues).reduce((acc, currKey) => {
                const newKey = currKey.split('_')[0];
                acc[newKey] = currentRowValues[currKey];
                return acc;
            }, {});

            const updatedKey = RelationshipObjectToKey(updatedRow);
            updatedRow.key = updatedKey;

            // Check if the updated key exists in the data source (excluding current item)
            const existingItem = newData.find(item => item.key === updatedRow.key && item.key !== key);
            if (existingItem) {
                const errorMsg = 'This relationship already exists. Please check your input.';
                setErrorMessage(errorMsg);
                message.error(errorMsg);
                return;
            }

            if (index > -1) {
                const isNewItem = originalItem.isNew;
                newData[index] = {...newData[index], ...updatedRow, isNew: false};
                setDataSource(newData);
                setEditingKeys(prevKeys => prevKeys.filter(k => k !== key));
                form.resetFields();
                
                if (isNewItem) {
                    // Add new relationship
                    addRelationships([updatedKey]);
                } else {
                    // Update existing relationship
                    removeRelationship(key);
                    addRelationships([updatedKey]);
                }
                
                message.success(isNewItem ? 'Relationship added successfully' : 'Relationship updated successfully');
            }
        } catch (errInfo) {
            const errorMsg = 'Validation failed. Please check all required fields.';
            setErrorMessage(errorMsg);
            message.error(errorMsg);
        }
    }, [dataSource, form, addRelationships, removeRelationship]);

    // Edit an existing relationship
    const editRow = useCallback((record) => {
        setEditingKeys(prevKeys => [...prevKeys, record.key]);
        
        // Populate form fields with existing data
        form.setFieldsValue({
            [`entityType_${record.key}`]: record.entityType,
            [`entityID_${record.key}`]: record.entityID,
            [`relation_${record.key}`]: record.relation,
            [`subjectType_${record.key}`]: record.subjectType,
            [`subjectID_${record.key}`]: record.subjectID,
            [`subjectRelation_${record.key}`]: record.subjectRelation,
        });
        
        // Set selected values for dropdowns
        setSelectedEntityTypes(prev => ({
            ...prev,
            [record.key]: record.entityType
        }));
        setSelectedRelations(prev => ({
            ...prev,
            [record.key]: record.relation
        }));
        setSelectedSubjectTypes(prev => ({
            ...prev,
            [record.key]: record.subjectType
        }));
        
        setErrorMessage('');
    }, [form]);

    // Column configurations for the table
    const columns = useMemo(() => [
        {
            key: 'sort',
        },
        {
            title: 'Entity Type',
            dataIndex: 'entityType',
            key: 'entityType',
            editable: true,
        },
        {
            title: 'Entity ID',
            dataIndex: 'entityID',
            key: 'entityID',
            editable: true,
        },
        {
            title: 'Relation',
            dataIndex: 'relation',
            key: 'relation',
            editable: true,
        },
        {
            title: 'Subject Type',
            key: 'subjectType',
            dataIndex: 'subjectType',
            editable: true,
        },
        {
            title: 'Subject ID',
            key: 'subjectID',
            dataIndex: 'subjectID',
            editable: true,
        },
        {
            title: 'Subject Relation',
            key: 'subjectRelation',
            dataIndex: 'subjectRelation',
            editable: true,
        },
        {
            title: '',
            dataIndex: 'operation',
            align: 'right',
            render: (_, record) => {
                const editable = isEditing(record);
                return editable ? (
                    <span className="flex flex-row items-center gap-2" style={{width: 'fit-content', marginLeft: 'auto'}}>
                        <Button type="primary" onClick={() => save(record.key)} aria-label="Save relationship">Save</Button>
                        <Button className="text-white ml-4" type="link" icon={<CloseOutlined/>} onClick={() => cancel(record.key)} aria-label="Cancel editing"/>
                    </span>
                ) : (
                    <Space size="small">
                        <Button type="text" icon={<EditOutlined/>} onClick={() => editRow(record)} aria-label="Edit relationship"/>
                        <Button type="text" danger icon={<DeleteOutlined/>} onClick={() => remove(record)} aria-label="Delete relationship"/>
                    </Space>
                );
            },
        },
    ], [isEditing, save, cancel, remove, editRow]);

    // Merge columns with editing logic
    const mergedColumns = useMemo(() => columns.map((col) => {
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
    }), [columns, isEditing]);

    // Handle the end of the drag-and-drop event
    const onDragEnd = useCallback(({active, over}) => {
        if (active.id !== over?.id) {
            setDataSource((previous) => {
                const activeIndex = previous.findIndex((i) => i.key === active.id);
                const overIndex = previous.findIndex((i) => i.key === over?.id);
                return arrayMove(previous, activeIndex, overIndex);
            });
        }
    }, []);

    // A table row component for drag-and-drop functionality
    const Row = useCallback(({children, ...props}) => {
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
                style={style} {...attributes} role="row" aria-label={`Relationship row ${props['data-row-key']}`}>
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
                                    aria-label="Drag to reorder relationship"
                                    tabIndex={0}
                                />
                            ),
                        });
                    }
                    return child;
                })}
            </tr>
        );
    }, []);

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
    const addRow = useCallback(() => {
        const newRow = {
            key: nanoid(),
            entityType: '',
            entityID: '',
            relation: '',
            subjectType: '',
            subjectID: '',
            subjectRelation: '',
            isNew: true,
        };
        setDataSource((prevData) => [...prevData, newRow]);
        setEditingKeys(prevKeys => [...prevKeys, newRow.key]);
        setErrorMessage('');
    }, []);


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
    const handleEntityTypeChange = useCallback((value, rowKey) => {
        setSelectedEntityTypes(prev => ({
            ...prev,
            [rowKey]: value
        }));
    }, []);

    const handleRelationChange = useCallback((value, rowKey) => {
        setSelectedRelations(prev => ({
            ...prev,
            [rowKey]: value
        }));
    }, []);

    const handleSubjectTypeChange = useCallback((value, rowKey) => {
        setSelectedSubjectTypes(prev => ({
            ...prev,
            [rowKey]: value
        }));
    }, []);

    // Utility functions to get options for relation, subject type, and subject relation
    const safelyGetOptions = useCallback((entityType, selectorFunction, mapFunction) => {
        if (!entityType) return [];

        const results = selectorFunction(entityType);
        if (!results || !results.length) return [];

        return results.map(mapFunction);
    }, []);

    const getRelationOptionsForRow = useCallback((rowKey) => {
        const entityType = selectedEntityTypes[rowKey];
        return safelyGetOptions(entityType, getRelations, relation => (
            <Option key={relation} value={relation}>
                {relation}
            </Option>
        ));
    }, [selectedEntityTypes, safelyGetOptions, getRelations]);

    const getSubjectTypeOptionsForRow = useCallback((rowKey) => {
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
    }, [selectedEntityTypes, selectedRelations, safelyGetOptions, getSubjectTypes]);

    const getSubjectRelationOptionsForRow = useCallback((rowKey) => {
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
    }, [selectedEntityTypes, selectedRelations, selectedSubjectTypes, safelyGetOptions, getSubjectRelations]);

    // The main render method
    return (
        <div style={{height: '100vh'}} ref={tableBodyRef}>
            {errorMessage && (
                <div style={{color: 'red', marginBottom: '10px', padding: '8px', backgroundColor: '#fff2f0', border: '1px solid #ffccc7', borderRadius: '4px'}}>
                    {errorMessage}
                </div>
            )}
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
                                        <Button type="primary" onClick={addRow} aria-label="Add new relationship">
                                            Add Relationship
                                        </Button>
                                    </div>
                                )}
                                scroll={{y: 'calc(100vh - 270px)'}}
                                aria-label="Relationships table"
                            />
                        </RelationshipErrorContext.Provider>
                    </Form>
                </SortableContext>
            </DndContext>
        </div>
    )
}

export default Relationships
