import React, {useEffect, useRef, useState} from "react";
import {Button, Grid, Layout, Row, Select, Typography} from 'antd';
import {toAbsoluteUrl} from "@utility/helpers/asset";
import {ExportOutlined, GithubOutlined, ShareAltOutlined, UploadOutlined} from "@ant-design/icons";
import {put} from '@vercel/blob';
import yaml from "js-yaml";
import {useShapeStore} from "@state/shape";
import Share from "@layout/components/Modals/Share";
import toast, {Toaster} from 'react-hot-toast';

const {Option, OptGroup} = Select;
const {Content} = Layout;
const {useBreakpoint} = Grid;
const {Text, Title} = Typography;

const MainLayout = ({children, ...rest}) => {
        const fileInputRef = useRef(null);

        const screens = useBreakpoint();

        const {
            schema,
            relationships,
            attributes,
            scenarios,
            setRelationships,
            setSchema,
            setAttributes,
            setScenarios,
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
            const yamlString = yaml.dump({
                schema: schema,
                relationships: relationships,
                attributes: attributes,
                scenarios: scenarios
            })
            const file = new File([yamlString], `s.yaml`, {type: 'text/x-yaml'});
            put("s.yaml", file, {access: 'public', token: process.env.REACT_APP_BLOB_READ_WRITE_TOKEN}).then((result) => {
                let fileName = result.url.split('/').pop();
                setId(fileName.replace('.yaml', ''))
            }).catch((error) => {
                console.error('Upload error:', error);
                toast.error('Failed to share. Please try again.');
            }).then(() => {
                toggleShareModalVisibility()
            });
        }
        const exp = () => {
            try {
                const yamlString = yaml.dump({
                    schema: schema,
                    relationships: relationships,
                    attributes: attributes,
                    scenarios: scenarios,
                });

                const blob = new Blob([yamlString], {type: 'text/x-yaml'});
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');

                a.href = url;
                a.download = 'shape.yaml';
                a.style.display = 'none';

                document.body.appendChild(a);
                a.click();

                // Clean up to avoid memory leaks
                document.body.removeChild(a);
                URL.revokeObjectURL(url);

                toast.success('Export successful! 🎉', {style: {borderRadius: '2px'}});
            } catch (err) {
                console.error('Export failed:', err);
                toast.error('Export failed. Please try again.', {style: {borderRadius: '2px'}});
            }
        }
        const imp = (event) => {
            const file = event.target.files[0];
            if (!file) {
                toast.error('No file selected.', {style: {borderRadius: '2px'}});
                return;
            }

            // Validate file type
            if (!file.name.endsWith('.yaml') && !file.name.endsWith('.yml')) {
                toast.error('Invalid file type. Please select a YAML file.', {style: {borderRadius: '2px'}});
                return;
            }

            // Validate file size (e.g., 10MB limit)
            if (file.size > 10 * 1024 * 1024) {
                toast.error('File size too large. Please select a smaller file.', {style: {borderRadius: '2px'}});
                return;
            }

            const reader = new FileReader();
            reader.onload = (e) => {
                try {
                    const yamlContent = e.target.result;
                    const result = yaml.load(yamlContent);

                    // Check if the schema exists
                    if (!result?.schema) {
                        toast.error('Missing schema in the YAML file.', {style: {borderRadius: '2px'}});
                        return;
                    }

                    // Update state with parsed YAML data
                    setSchema(result.schema ?? ``);
                    setRelationships(result.relationships ?? []);
                    setAttributes(result.attributes ?? []);
                    setScenarios(result.scenarios ?? []);

                    toast.success('File imported successfully! 🎉', {style: {borderRadius: '2px'}});
                } catch (err) {
                    console.error('Error parsing YAML file:', err);
                    toast.error('Failed to import shape: Invalid YAML format.', {style: {borderRadius: '2px'}});
                }
            };
            reader.readAsText(file);

            event.target.value = '';
        };

        const triggerFileInput = () => {
            if (fileInputRef.current) {
                fileInputRef.current.click(); // Safely trigger the file input
            } else {
                console.error('File input reference is null.');
            }
        };


        return (
            <Layout className="App h-screen flex flex-col">

                <Toaster position="top-right" reverseOrder={false}/>

                <Share visible={shareModalVisibility} toggle={toggleShareModalVisibility} id={id}></Share>

                {/* Conditional message for small screens */}
                {!screens.md ?
                    <div className="center-of-screen text-center">
                        <Title level={5}>Screen size too small!</Title>
                        <Text>This application is optimized for larger screens. Please switch to a larger screen for the
                            best experience.</Text>
                    </div>
                    :
                    <>
                        <Row justify="center" type="flex" className="header">
                            <div className="logo">
                                <a href="/">
                                    <img alt="logo"
                                         src={toAbsoluteUrl("/media/svg/permify.svg")}/>
                                </a>
                            </div>
                            <div className="ml-12">
                                <Select className="mr-8" value={label} style={{width: 220}}
                                        onChange={handleSampleChange}>
                                    <OptGroup label="Use Cases">
                                        <Option key="empty" value="empty">Empty</Option>
                                        <Option key="organizations-hierarchies" value="organizations-hierarchies">Organizations
                                            &
                                            Hierarchies</Option>
                                        <Option key="rbac" value="rbac">RBAC</Option>
                                        <Option key="custom-roles" value="custom-roles">Custom Roles</Option>
                                        <Option key="user-groups" value="user-groups">User Groups</Option>
                                        <Option key="banking-system" value="banking-system">Banking System</Option>
                                    </OptGroup>
                                    <OptGroup label="Sample Apps">
                                        <Option key="google-docs-simplified" value="google-docs-simplified">Google Docs
                                            Simplified</Option>
                                        <Option key="facebook-groups" value="facebook-groups">Facebook Groups</Option>
                                        <Option key="notion" value="notion">Notion</Option>
                                        <Option key="mercury" value="mercury">Mercury</Option>
                                        <Option key="instagram" value="instagram">Instagram</Option>
                                        <Option key="disney-plus" value="disney-plus">Disney Plus</Option>
                                    </OptGroup>
                                </Select>
                                <Button
                                    className="mr-8"
                                    icon={<UploadOutlined/>}
                                    onClick={triggerFileInput}
                                >
                                    Import
                                </Button>
                                <input
                                    type="file"
                                    ref={fileInputRef}
                                    accept=".yaml"
                                    style={{display: 'none'}}
                                    onChange={imp}
                                />
                                <Button className="mr-8" onClick={() => {
                                    exp()
                                }} icon={<ExportOutlined/>}>Export</Button>
                                <Button onClick={() => {
                                    share()
                                }} icon={<ShareAltOutlined/>}>Share</Button>
                            </div>
                            <div className="ml-auto">
                                {screens.xl &&
                                    <>
                                        <Button className="mr-8 text-white" type="link" target="_blank"
                                                href="https://docs.permify.co/docs/playground">
                                            How to use playground?
                                        </Button>
                                        <Button className="mr-8" target="_blank" icon={<GithubOutlined/>}
                                                href="https://github.com/Permify/permify">
                                            Star us on GitHub
                                        </Button>
                                        <Button type="primary" href="https://discord.com/invite/MJbUjwskdH" target="_blank">
                                            Join Community
                                        </Button>
                                    </>
                                }
                            </div>
                        </Row>

                        <Layout>
                            <Content className="h-full flex flex-col max-h-full">
                                <div className="flex-auto overflow-hidden">
                                    {children}
                                </div>
                            </Content>
                        </Layout>
                    </>
                }
            </Layout>
        );
    }
;

export default MainLayout;
