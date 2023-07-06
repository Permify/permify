import React, {useState} from "react";
import {Layout, Row, Button, Select} from 'antd';
import {toAbsoluteUrl} from "../utility/helpers/asset";
import {GithubOutlined, ShareAltOutlined} from "@ant-design/icons";
import {shallowEqual, useSelector} from "react-redux";
import yaml from "js-yaml";
import Upload from "../services/s3";
import Share from "./components/Modals/Share";
import {nanoid} from "nanoid";

const {Option, OptGroup} = Select;
const {Content, Header} = Layout;

const MainLayout = ({children, ...rest}) => {

    const [selectedSample, setSelectedSample] = useState("View an Example");
    const [shareModalVisibility, setShareModalVisibility] = React.useState(false);

    const toggleShareModalVisibility = () => {
        setShareModalVisibility(!shareModalVisibility);
    };

    const [id, setId] = useState("");

    const shape = useSelector((state) => state.shape, shallowEqual);

    const handleSampleChange = (value) => {
        setSelectedSample(value)
        const params = new URLSearchParams()
        params.append("s", value)
        window.location = window.location.href.split('?')[0] + `?${params.toString()}`
    };

    const share = () => {
        let id = nanoid()
        setId(id)
        const yamlString = yaml.dump({
            schema: shape.schema,
            relationships: shape.relationships,
            assertions: shape.assertions
        })
        const file = new File([yamlString], `shapes/${id}.yaml`, {type : 'text/x-yaml'});
        Upload(file).then((res) => {
            toggleShareModalVisibility()
        })
    }

    return (
        <Layout className="App h-screen flex flex-col">
            <Header className="header">

                <Share visible={shareModalVisibility} toggle={toggleShareModalVisibility} id={id}></Share>

                <Row justify="center" type="flex">
                    <div className="logo">
                        <a href="/">
                            <img alt="logo"
                                 src={toAbsoluteUrl("/media/svg/permify.svg")}/>
                        </a>
                    </div>
                    <div className="ml-12">
                        <Select className="mr-8" value={selectedSample} style={{width: 220}} onChange={handleSampleChange} showArrow={true}>
                            <OptGroup label="Use Cases">
                                <Option key="empty" value="empty">Empty</Option>
                                <Option key="organizations-hierarchies" value="organizations-hierarchies">Organizations & Hierarchies</Option>
                                <Option key="rbac" value="rbac">RBAC</Option>
                                <Option key="custom-roles" value="custom-roles">Custom Roles</Option>
                                <Option key="user-groups" value="user-groups">User Groups</Option>
                            </OptGroup>
                            <OptGroup label="Sample Apps">
                                <Option key="google-docs-simplified" value="google-docs-simplified">Google Docs Simplified</Option>
                                <Option key="facebook-groups" value="facebook-groups">Facebook Groups</Option>
                                <Option key="notion" value="notion">Notion</Option>
                            </OptGroup>
                        </Select>
                        <Button onClick={() => {
                            share()
                        }} icon={<ShareAltOutlined/>}>Share</Button>
                    </div>
                    <div className="ml-auto">
                        <Button className="mr-8 text-white" type="link" target="_blank"
                                href="https://www.permify.co/change-log/permify-playground">
                            How to use playground?
                        </Button>
                        <Button className="mr-8" target="_blank" icon={<GithubOutlined/>}
                                href="https://github.com/Permify/permify">
                            Get Started
                        </Button>
                        <Button type="primary" href="https://discord.com/invite/MJbUjwskdH" target="_blank">Join
                            Community</Button>
                    </div>
                </Row>
            </Header>
            <Layout className="m-10">
                <Content className="h-full flex flex-col max-h-full">
                    <div className="flex-auto overflow-hidden">
                        {children}
                    </div>
                </Content>
            </Layout>
        </Layout>
    );
};

export default MainLayout;
