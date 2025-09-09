import React, {useEffect, useRef, useState, useMemo, useCallback} from "react";
import {Button, Form, Input, Select, Space, Table, message} from "antd";
import {CloseOutlined, DeleteOutlined, MenuOutlined, EditOutlined} from "@ant-design/icons";
import {useShapeStore} from "@state/shape";
import {arrayMove, SortableContext, useSortable, verticalListSortingStrategy} from "@dnd-kit/sortable";
import {AttributeEntityToKey, AttributeObjectToKey, StringAttributesToObjects,} from "@utility/helpers/common";
import {DndContext, KeyboardSensor, PointerSensor, useSensor, useSensors} from "@dnd-kit/core";
import {nanoid} from "nanoid";

const {Option} = Select;

const AttributeErrorContext = React.createContext([]);

function Attributes() {

    const [form] = Form.useForm();
    const [dataSource, setDataSource] = useState([]);
    const [errorMessage, setErrorMessage] = useState('');
    const tableBodyRef = useRef(null);

    const {
        attributes,
        addAttributes,
        removeAttribute,
        getEntityTypes,
        getAttributes,
        getTypeValueBasedOnAttribute,
        attributeErrors
    } = useShapeStore();

    const [selectedEntityTypes, setSelectedEntityTypes] = useState({});
    const [selectedAttributes, setSelectedAttributes] = useState({});

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

    // Remove a specific attribute from the data source
    const remove = useCallback((attribute) => {
        const updatedDataSource = dataSource.filter(item => item.key !== attribute.key);
        setDataSource(updatedDataSource);
        removeAttribute(attribute.key);
        setErrorMessage('');
    }, [dataSource, removeAttribute]);

    // Save the current edits
    const save = useCallback(async (key) => {
        try {
            setErrorMessage('');
            const currentRowValues = await form.validateFields([
                `entityType_${key}`,
                `entityID_${key}`,
                `attribute_${key}`,
                `type_${key}`,
                `value_${key}`,
            ]);

            const newData = [...dataSource];
            const index = newData.findIndex((item) => key === item.key);
            const originalItem = newData[index];

            const updatedRow = Object.keys(currentRowValues).reduce((acc, currKey) => {
                const newKey = currKey.split('_')[0];
                acc[newKey] = currentRowValues[currKey];
                return acc;
            }, {});

            const updatedKey = AttributeObjectToKey(updatedRow);
            updatedRow.key = updatedKey;

            // Check if the updated key exists in the data source (excluding current item)
            const existingItem = newData.find(item => AttributeEntityToKey(item) === AttributeEntityToKey(updatedRow) && item.key !== key);
            if (existingItem) {
                const errorMsg = 'This attribute already exists. Please check your input.';
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
                    // Add new attribute
                    addAttributes([updatedKey]);
                } else {
                    // Update existing attribute
                    removeAttribute(key);
                    addAttributes([updatedKey]);
                }
                
                message.success(isNewItem ? 'Attribute added successfully' : 'Attribute updated successfully');
            }
        } catch (errInfo) {
            const errorMsg = 'Validation failed. Please check all required fields.';
            setErrorMessage(errorMsg);
            message.error(errorMsg);
        }
    }, [dataSource, form, addAttributes, removeAttribute]);

    // Edit an existing attribute
    const editRow = useCallback((record) => {
        setEditingKeys(prevKeys => [...prevKeys, record.key]);
        
        // Populate form fields with existing data
        form.setFieldsValue({
            [`entityType_${record.key}`]: record.entityType,
            [`entityID_${record.key}`]: record.entityID,
            [`attribute_${record.key}`]: record.attribute,
            [`type_${record.key}`]: record.type,
            [`value_${record.key}`]: record.value,
        });
        
        // Set selected values for dropdowns
        setSelectedEntityTypes(prev => ({
            ...prev,
            [record.key]: record.entityType
        }));
        setSelectedAttributes(prev => ({
            ...prev,
            [record.key]: record.attribute
        }));
        
        setErrorMessage('');
    }, [form]);

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
            title: 'Attribute',
            dataIndex: 'attribute',
            key: 'attribute',
            editable: true,
        },
        {
            title: 'Type',
            key: 'type',
            dataIndex: 'type',
            editable: true,
        },
        {
            title: 'Value',
            key: 'value',
            dataIndex: 'value',
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
                        <Button type="primary" onClick={() => save(record.key)} aria-label="Save attribute">Save</Button>
                        <Button className="text-white ml-4" type="link" icon={<CloseOutlined/>} onClick={() => cancel(record.key)} aria-label="Cancel editing"/>
                    </span>
                ) : (
                    <Space size="small">
                        <Button type="text" icon={<EditOutlined/>} onClick={() => editRow(record)} aria-label="Edit attribute"/>
                        <Button type="text" danger icon={<DeleteOutlined/>} onClick={() => remove(record)} aria-label="Delete attribute"/>
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

        const attributeErrors = React.useContext(AttributeErrorContext);

        // Find if there's an error for this row
        const attributeError = attributeErrors.find(
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
            <tr {...props} ref={setNodeRef} className={attributeError ? 'error-row' : ''}
                style={style} {...attributes} role="row" aria-label={`Attribute row ${props['data-row-key']}`}>
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
                                    aria-label="Drag to reorder attribute"
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

    useEffect(() => {
        setDataSource(StringAttributesToObjects(attributes))
    }, [attributes]);

    // Scroll to the latest data row after adding a new attribute
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

    // Add a new attribute row for editing
    const addRow = useCallback(() => {
        const newRow = {
            key: nanoid(),
            entityType: '',
            entityID: '',
            attribute: '',
            type: '',
            value: '',
            isNew: true,
        };
        setDataSource((prevData) => [...prevData, newRow]);
        setEditingKeys(prevKeys => [...prevKeys, newRow.key]);
        setErrorMessage('');
    }, []);


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
                case 'attribute':
                    inputElement = (
                        <Select
                            showSearch
                            allowClear
                            placeholder="Attribute"
                            notFoundContent={null}
                            onChange={(value) => handleAttributeChange(value, record.key)}
                        >
                            {getAttributeOptionsForRow(record.key)}
                        </Select>
                    );
                    break;
                case 'entityID':
                    inputElement = (
                        <Input placeholder="Entity ID"/>
                    );
                    break;
                case 'type':
                    inputElement = (
                        <Input placeholder="Type" disabled={true}/>
                    );
                    break;
                case 'value':
                    inputElement = (
                        <Input placeholder="Value"/>
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
                                required: true,
                                message: ``
                            }
                        ]}
                    >
                        {inputElement}
                    </Form.Item>
                </td>
            );
        }

        return (
            <td {...restProps}>
                {children}
            </td>
        );
    };

    // Handlers to set selected entity types, attributes, and subject types when they are changed
    const handleEntityTypeChange = useCallback((value, rowKey) => {
        setSelectedEntityTypes(prev => ({
            ...prev,
            [rowKey]: value
        }));
    }, []);

    const handleAttributeChange = useCallback((value, rowKey) => {
        try {
            // Update selectedAttributes state
            if (value) {
                setSelectedAttributes(prev => ({
                    ...prev,
                    [rowKey]: value
                }));
            } else {
                const {[rowKey]: _, ...remaining} = selectedAttributes;
                setSelectedAttributes(remaining);
            }

            // Get corresponding entityType from the selectedEntityTypes state
            const entityType = selectedEntityTypes[rowKey];

            if (entityType && value) {
                const typeValue = getTypeValueBasedOnAttribute(entityType, value);

                // Ensure the typeValue exists before updating the form
                if (typeValue !== undefined) {
                    form.setFieldsValue({
                        [`type_${rowKey}`]: typeValue
                    });
                }
            }

        } catch (error) {
            console.error('Error handling attribute change:', error);
        }
    }, [selectedEntityTypes, form, getTypeValueBasedOnAttribute]);

    // Utility functions to get options for attribute, subject type, and subject attribute
    const safelyGetOptions = useCallback((entityType, selectorFunction, mapFunction) => {
        if (!entityType) return [];

        const results = selectorFunction(entityType);
        if (!results || !results.length) return [];

        return results.map(mapFunction);
    }, []);

    const getAttributeOptionsForRow = useCallback((rowKey) => {
        const entityType = selectedEntityTypes[rowKey];
        return safelyGetOptions(entityType, getAttributes, attribute => (
            <Option key={attribute} value={attribute}>
                {attribute}
            </Option>
        ));
    }, [selectedEntityTypes, safelyGetOptions, getAttributes]);

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
                        <AttributeErrorContext.Provider value={attributeErrors}>
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
                                        <Button type="primary" onClick={addRow} aria-label="Add new attribute">
                                            Add Attribute
                                        </Button>
                                    </div>
                                )}
                                scroll={{y: 'calc(100vh - 270px)'}}
                                aria-label="Attributes table"
                            />
                        </AttributeErrorContext.Provider>
                    </Form>
                </SortableContext>
            </DndContext>
        </div>
    )
}

export default Attributes
