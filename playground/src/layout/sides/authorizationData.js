import React, {useEffect, useRef, useState} from 'react'
import {Alert, Button, Card, List, Tooltip, Typography} from 'antd';
import {DeleteOutlined, InfoCircleOutlined} from "@ant-design/icons";
import Create from "../components/Modals/tuple/Create";
import {shallowEqual, useSelector} from "react-redux";
import TupleToHumanLanguage from "../../utility/helpers/tuple";

const {Text} = Typography;

function AuthorizationData() {

    const ref = useRef(false);

    // Create Modal
    const [createModalVisibility, setCreateModalVisibility] = React.useState(false);

    const toggleCreateModalVisibility = () => {
        setCreateModalVisibility(!createModalVisibility);
        readTuples()
    };

    const [error, setError] = useState("");
    const [filter, setFilter] = useState("");
    const [model, setModel] = useState({entityDefinitions: {}});
    const [tabList, setTabList] = useState([]);
    const [tuples, setTuples] = useState([]);

    const trigger = useSelector((state) => state.common.model_change_toggle, shallowEqual);

    const DeleteTupleCall = (tuple) => {
        return new Promise((resolve) => {
            let res = window.deleteTuple(JSON.stringify(tuple))
            resolve(res);
        });
    }

    const ReadSchemaCall = () => {
        return new Promise((resolve) => {
            let res = window.readSchema("")
            resolve(res);
        });
    }

    const ReadTuplesCall = (filter) => {
        return new Promise((resolve) => {
            let res = window.readTuple(JSON.stringify(filter))
            resolve(res);
        });
    }

    const readSchema = () => {
        ReadSchemaCall().then((res) => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            }
            let m = JSON.parse(res[0])
            if (res[0] !== null) {
                setModel(m)
                let en = []
                let i = 0
                for (const [_, definition] of Object.entries(m.entityDefinitions)) {
                    if (i === 0) {
                        setFilter(definition.name)
                    }
                    en.push({key: definition.name, tab: definition.name})
                    i++
                }
                setTabList(en)
            }
        })
    }

    const deleteTuple = (tuple) => {
        DeleteTupleCall(tuple).then((res) => {
            if (res[0] != null) {
                setError(res[0].replaceAll('_', ' '))
            }
            readTuples()
        })
    }

    const readTuples = () => {
        ReadTuplesCall({
            entity: {
                type: filter
            }
        }).then((res) => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            }
            let p = JSON.parse(res[0])
            if (p.tuples !== undefined) {
                setTuples(p.tuples)
            } else {
                setTuples([])
            }
        })
    }

    useEffect(() => {
        if (ref.current) {
            readSchema()
            readTuples()
        }
        ref.current = true;
    }, [trigger]);

    useEffect(() => {
        if (filter !== "") {
            readTuples()
        }
    }, [filter]);

    return (
        <>
            <Create visible={createModalVisibility} toggle={toggleCreateModalVisibility} model={model}
                    setFilter={setFilter}/>


            <Card title={
                <div>
                    <span className="mr-8">Authorization Data</span>
                    <Tooltip placement="right" color="black"
                             title={"Authorization data stored as Relation Tuples into your preferred database. These relational tuples represents your authorization data."}>
                        <InfoCircleOutlined/>
                    </Tooltip>
                </div>
            } className="mr-12 mt-12 h-screen"
                  tabList={tabList}
                  onTabChange={(key) => {
                      setFilter(key)
                  }}
                  activeTabKey={filter}
                  extra={<>
                      <Button className="ml-auto" type="primary" onClick={toggleCreateModalVisibility}>New</Button>
                  </>}>
                {error !== "" &&
                    <Alert message={error} type="error" showIcon className="mb-12 ml-12 mr-12"/>
                }
                <div className="px-12 pb-12 pt-12 mt-12">
                    <List
                        className="scroll"
                        itemLayout="horizontal"
                        dataSource={tuples}
                        renderItem={(item) => (
                            <List.Item
                                actions={[
                                    <Button type="text" danger
                                            icon={<DeleteOutlined onClick={() => deleteTuple(item)}/>}/>,
                                ]}
                            >
                                <List.Item.Meta
                                    avatar={<Text keyboard>[TUPLE]</Text>}
                                    title={TupleToHumanLanguage(`${item.entity.type}:${item.entity.id}#${item.relation}@${item.subject.type}:${item.subject.id}${item.subject.relation === undefined ? "" : "#" + item.subject.relation}`)}
                                    description={`${item.entity.type}:${item.entity.id}#${item.relation}@${item.subject.type}:${item.subject.id}${item.subject.relation === undefined ? "" : "#" + item.subject.relation}`}
                                />
                            </List.Item>
                        )}
                    />
                </div>
            </Card>

        </>
    )
}

export default AuthorizationData
