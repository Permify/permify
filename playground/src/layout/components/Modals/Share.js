import React, {useState} from "react";
import {Button, Input, Modal} from "antd";
import {CopyOutlined} from "@ant-design/icons";

function Share(props) {

    const [isLinkCopied, setIsLinkCopied] = useState(false);

    const handleOk = () => {
        props.toggle();
    };

    const handleCancel = () => {
        props.toggle();
    };

    function copyLink(text) {
        if ('clipboard' in navigator) {
            setIsLinkCopied(true)
            return navigator.clipboard.writeText(text);
        } else {
            return document.execCommand('copy', true, text);
        }
    }

    return (
        <Modal
            title="Share"
            open={props.visible}
            onOk={handleOk}
            onCancel={handleCancel}
            destroyOnClose
            bordered={true}
            footer={null}
        >
            <div style={{display: "flex", gap: "8px", marginBottom: "15px", marginTop: "20px"}}>
                <Input
                    defaultValue={`https://play.permify.co/?s=${props.id}`}
                />
                <Button
                    type="primary"
                    onClick={() => {
                        copyLink(`https://play.permify.co/?s=${props.id}`)
                    }}
                    icon={<CopyOutlined/>}
                >
                    {isLinkCopied ? 'Copied!' : 'Copy'}
                </Button>
            </div>
        </Modal>
    )
}

export default Share


