import React, {useState, useEffect} from "react";
import {Layout, Row, Button, Select} from 'antd';
import {toAbsoluteUrl} from "../utility/helpers/asset";
import {GithubOutlined, ShareAltOutlined, ExportOutlined} from "@ant-design/icons";
import Upload from "../services/s3";
import {nanoid} from "nanoid";
import yaml from "js-yaml";
import {useShapeStore} from "../state/shape";
import Share from "./components/Modals/Share";

const {Option, OptGroup} = Select;
const {Content} = Layout;

const MainLayout = ({children, ...rest}) => {

    const {
        schema,
        relationships,
        attributes,
        scenarios,
    } = useShapeStore();

    const [label, setLabel] = useState("View an Example");
    const [shareModalVisibility, setShareModalVisibility] = React.useState(false);

    const toggleShareModalVisibility = () => {
        setShareModalVisibility(!shareModalVisibility);
    };

    const [id, setId] = useState("");

    const handleSampleChange = (value) => {
        const params = new URLSearchParams()
        params.append("s", value)
        window.location = window.location.href.split('?')[0] + `?${params.toString()}`
    };

    useEffect(() => {
        const searchParams = new URLSearchParams(window.location.search);
        const sParam = searchParams.get('s');
        if (sParam) {
            setLabel(sParam);
        }
    }, []);

    const share = () => {
        let id = nanoid()
        setId(id)
        const yamlString = yaml.dump({
            schema: schema,
            relationships: relationships,
            attributes: attributes,
            scenarios: scenarios
        })
        const file = new File([yamlString], `shapes/${id}.yaml`, {type: 'text/x-yaml'});
        Upload(file).then((res) => {
            toggleShareModalVisibility()
        })
    }

    const exp = () => {
        const yamlString = yaml.dump({
            schema: schema,
            relationships: relationships,
            attributes: attributes,
            scenarios: scenarios
        });

        const blob = new Blob([yamlString], { type: 'text/x-yaml' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');

        a.href = url;
        a.download = 'shape.yaml';  // or any other name you want
        a.style.display = 'none';

        document.body.appendChild(a);
        a.click();

        // Clean up to avoid memory leaks
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    return (
        <Layout className="App h-screen flex flex-col">
            <Share visible={shareModalVisibility} toggle={toggleShareModalVisibility} id={id}></Share>

            <Row justify="center" type="flex" className="header">
                <div className="logo">
                    <a href="/">
                        <img alt="logo"
                             src={toAbsoluteUrl("/media/svg/permify.svg")}/>
                    </a>
                </div>
                <div className="ml-12">
                    <Select className="mr-8" value={label} style={{width: 220}}
                            onChange={handleSampleChange} showArrow={true}>
                        <OptGroup label="Use Cases">
                            <Option key="empty" value="empty">Empty</Option>
                            <Option key="organizations-hierarchies" value="organizations-hierarchies">Organizations & Hierarchies</Option>
                            <Option key="rbac" value="rbac">RBAC</Option>
                            <Option key="custom-roles" value="custom-roles">Custom Roles</Option>
                            <Option key="user-groups" value="user-groups">User Groups</Option>
                            <Option key="weekday" value="weekday">Weekday</Option>
                            <Option key="banking-system" value="banking-system">Banking System</Option>
                        </OptGroup>
                        <OptGroup label="Sample Apps">
                            <Option key="google-docs-simplified" value="google-docs-simplified">Google Docs Simplified</Option>
                            <Option key="facebook-groups" value="facebook-groups">Facebook Groups</Option>
                            <Option key="notion" value="notion">Notion</Option>
                            <Option key="mercury" value="mercury">Mercury</Option>
                            <Option key="instagram" value="instagram">Instagram</Option>
                            <Option key="disney-plus" value="disney-plus">Disney Plus</Option>
                        </OptGroup>
                    </Select>
                    <Button className="mr-8" onClick={() => {
                        exp()
                    }} icon={<ExportOutlined/>}>Export</Button>
                    <Button  onClick={() => {
                        share()
                    }} icon={<ShareAltOutlined/>}>Share</Button>
                </div>
                <div className="ml-auto">
                    <Button className="mr-8 text-white" type="link" target="_blank"
                            href="https://docs.permify.co/docs/playground">
                        How to use playground?
                    </Button>
                    <Button className="mr-8" target="_blank" icon={<GithubOutlined/>}
                            href="https://github.com/Permify/permify">
                        Star us on GitHub
                    </Button>
                    <Button type="primary" href="https://discord.com/invite/MJbUjwskdH" target="_blank">Join
                        Community</Button>
                </div>
            </Row>

            <Layout>
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
