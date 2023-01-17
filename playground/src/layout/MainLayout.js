import React, {useState} from "react";
import {Layout, Row, Button, Select} from 'antd';
import {withRouter} from "react-router-dom"
import {toAbsoluteUrl} from "../utility/helpers/asset";
import {GithubOutlined} from "@ant-design/icons";
import {useHistory} from "react-router-dom"

const {Option, OptGroup} = Select;
const {Content, Header} = Layout;

const MainLayout = ({children, ...rest}) => {

    const [selectedSample, setSelectedSample] = useState("View an Example");

    const history = useHistory()

    const handleSampleChange = (value) => {
        setSelectedSample(value)
        const params = new URLSearchParams()
        params.append("s", value)
        history.push({search: params.toString()})
    };

    return (
        <Layout className="App h-screen flex flex-col">
            <Header className="header">
                <Row justify="center" type="flex">
                    <div className="logo">
                        <a href="/">
                            <img alt="logo"
                                 src={toAbsoluteUrl("/media/svg/permify.svg")}/>
                        </a>
                    </div>
                    <div className="ml-12">
                        <Select value={selectedSample} style={{width: 220}} onChange={handleSampleChange} showArrow={false}>
                            <OptGroup label="Use Cases">
                                <Option key="i" value="i">Empty</Option>
                                <Option key="p" value="p">Organizations
                                    & Hierarchies</Option>
                                <Option key="a" value="a">RBAC</Option>
                                <Option key="s" value="s">User Groups</Option>
                                <Option key="d" value="d">Parent Child
                                    Relationships</Option>
                            </OptGroup>
                            <OptGroup label="Sample Apps">
                                <Option key="f" value="f">Github</Option>
                                <Option key="g" value="g">Mercury</Option>
                            </OptGroup>
                        </Select>
                    </div>
                    <div className="ml-auto">
                        <Button className="mr-8 text-white" type="link" target="_blank"
                                href="https://www.permify.co/change-log/permify-playground">
                            How to use playground?
                        </Button>
                        <Button className="mr-8" type="secondary" target="_blank" icon={<GithubOutlined/>}
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

export default withRouter(MainLayout);
