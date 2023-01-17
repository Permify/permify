import React, {useState, useEffect} from 'react'
import {Button, Card, Space, Alert, Tooltip, Modal, Input} from 'antd';
import {InfoCircleOutlined, SaveOutlined, CopyOutlined, ShareAltOutlined} from "@ant-design/icons";
import Editor from "../../pkg/Editor";
import {useLocation, useHistory} from 'react-router-dom'
import queryString from "query-string";
import axios from "axios";
import {shallowEqual, useDispatch, useSelector} from "react-redux";
import {setIsLoading, setModelChangeActivity} from "../../redux/common/actions";

const client = axios.create();

function AuthorizationModel(props) {

    const modelChangeTrigger = useSelector((state) => state.common.model_change_toggle, shallowEqual);

    let history = useHistory();
    const dispatch = useDispatch();

    const {search} = useLocation()

    const [error, setError] = useState("");
    const [model, setModel] = useState(``);
    const [isModelCopied, setIsModelCopied] = useState(false);
    const [isLinkCopied, setIsLinkCopied] = useState(false);
    const [open, setOpen] = useState(false);
    const [short, setShort] = useState("");

    const defaultModel = `entity user {}
        
entity organization {

    // organizational roles
    relation admin @user
    relation member @user
    
}

entity repository {

    // represents repositories parent organization
    relation parent @organization
    
    // represents owner of this repository
    relation owner  @user
    
    // permissions
    action edit   = parent.admin or owner
    action delete = owner
    
} `

    const showModal = (link) => {
        setShort(link);
        setOpen(true);
    };

    const call = (m) => {
        return new Promise((resolve) => {
            let res = window.writeSchema(m)
            resolve(res);
        });
    }

    const save = (m) => {
        setError("")
        call(m).then((res) => {
            if (res[1] != null) {
                setError(res[1].replaceAll('_', ' '))
            } else {
                setIsModelCopied(false)
                dispatch(setModelChangeActivity(!modelChangeTrigger))
            }
        })
    }

    useEffect(() => {
        if (search !== "") {
            dispatch(setIsLoading(true))
            const values = queryString.parse(search)
            client.get(`https://backoffice.permify.co/v1/schemas/get/${values.s}`).then((response) => {
                setModel(response.data.schema)
                save(response.data.schema)
                dispatch(setIsLoading(false))
            }).catch((error) => {
                history.push({
                    pathname: '/404',
                })
            });
        } else {
            setModel(defaultModel)
            save(defaultModel)
        }
    }, [search]);

    function copyModel(text) {
        if ('clipboard' in navigator) {
            setIsModelCopied(true)
            return navigator.clipboard.writeText(JSON.stringify(text));
        } else {
            return document.execCommand('copy', true, text);
        }
    }

    function copyLink(text) {
        if ('clipboard' in navigator) {
            setIsLinkCopied(true)
            return navigator.clipboard.writeText(text);
        } else {
            return document.execCommand('copy', true, text);
        }
    }

    function share(m) {
        client.post(`https://backoffice.permify.co/v1/schemas/store`, {
            schema: m,
            version: "-"
        }).then((response) => {
            showModal(response.data.short)
        }).catch((error) => {
            history.push({
                pathname: '/404',
            })
        });
    }

    const handleOk = () => {
        setOpen(false);
    };

    const handleCancel = () => {
        setOpen(false);
    };

    return (
        <>
            <Modal
                title="Share Your Model"
                visible={open}
                onOk={handleOk}
                onCancel={handleCancel}
                destroyOnClose
                bordered={true}
                footer={null}
            >
                <div style={{ display: "flex", gap: "8px" , marginBottom: "15px"}}>
                    <Input
                           defaultValue={`https://play.permify.co/?s=${short}`}
                    />
                    <Button
                        type="primary"
                        onClick={() => {
                            copyLink(`https://play.permify.co/?s=${short}`)
                        }}
                        icon={<CopyOutlined/>}
                    >
                        {isLinkCopied ? 'Copied!' : 'Copy'}
                    </Button>
                </div>
            </Modal>

            <Card title={<div>
                <span className="mr-8">Authorization Model</span>
                <Tooltip placement="right" color="black"
                         title={"Permify has its own language that you can model your authorization logic with it, we called it Permify Schema. You can define your entities, relations between them and access control decisions with using Permify Schema."}>
                    <InfoCircleOutlined/>
                </Tooltip>
            </div>} extra={<Space>

                <Button type="secondary" onClick={() => {
                    copyModel(model)
                }} icon={<CopyOutlined/>}>{isModelCopied ? 'Copied!' : 'Copy'}</Button>

                <Button type="secondary" onClick={() => {
                    share(model)
                }} icon={<ShareAltOutlined/>}>Share</Button>

                <Button type="primary" onClick={() => {
                    save(model)
                }} icon={<SaveOutlined/>}>Save</Button>

            </Space>} style={{marginRight: "10px", marginBottom: "10px"}}>
                {error !== "" && <Alert message={error} type="error" showIcon className="mb-12 ml-12 mr-12"/>}
                <Editor setCode={setModel} code={model}></Editor>
            </Card>
        </>
    )
}

export default AuthorizationModel
