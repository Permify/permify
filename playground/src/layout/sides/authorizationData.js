import React, {useEffect, useRef, useState} from 'react'
import {Alert, Button, Card, List, Typography} from 'antd';
import {DeleteOutlined} from "@ant-design/icons";
import CreateTuple from "../components/Modals/CreateTuple";
import {shallowEqual, useSelector} from "react-redux";
import TupleToHumanLanguage, {Tuple, TupleObjectToTupleString} from "../../utility/helpers/tuple";
import {DeleteTuple, ReadTuples, WriteTuple} from "../../services/relationship";
import {ReadSchema} from "../../services/schema";

const {Text} = Typography;

function AuthorizationData(props) {

    const ref = useRef(false);

    // CreateTuple Modal
    const [createModalVisibility, setCreateModalVisibility] = React.useState(false);

    const toggleCreateModalVisibility = () => {
        setCreateModalVisibility(!createModalVisibility);
        readTuples()
    };

    const [error, setError] = useState("");
    const [model, setModel] = useState({entityDefinitions: {}});
    const [tuples, setTuples] = useState([]);

    const trigger = useSelector((state) => state.common.model_change_toggle, shallowEqual);

    const readSchema = () => {
        ReadSchema().then((res) => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            }
            let m = JSON.parse(res[0])
            if (res[0] !== null) {
                setModel(m)
            }
        })
    }

    const deleteTuple = (tuple) => {
        DeleteTuple(tuple).then((res) => {
            if (res[0] != null) {
                setError(res[0].replaceAll('_', ' '))
            }
            readTuples()
        })
    }

    const readTuples = () => {
        ReadTuples({}).then((res) => {
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
        for (let i = 0; i < props.initialValue.length; i++) {
            WriteTuple(Tuple(props.initialValue[i])).then((res) => {
                if (res[0] != null) {
                    setError(res[0])
                }
            })
        }
    }, [props.initialValue]);

    return (
        <>
            <CreateTuple visible={createModalVisibility} toggle={toggleCreateModalVisibility} model={model}/>

            <Card title={props.title} className="h-screen"
                  extra={<>
                      <Button className="ml-auto" type="primary" onClick={toggleCreateModalVisibility}>New</Button>
                  </>} style={{display: props.hidden && 'none'}}>
                {error !== "" &&
                    <Alert message={error} type="error" showIcon className="mb-12 ml-12 mr-12"/>
                }
                <div className="px-12 pb-12 pt-12">
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
                                    title={TupleToHumanLanguage(TupleObjectToTupleString(item))}
                                    description={TupleObjectToTupleString(item)}
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
