import React, {useState, useRef} from 'react' // React core imports
import {Allotment} from 'allotment' // Layout component
import "allotment/dist/style.css"; // Allotment styles
import Schema from "@features/schema/schema"; // Schema editor
import Visualizer from "@features/schema/visualizer"; // Schema visualizer
import {Button, Card, Radio} from "antd"; // Ant Design components
import { // Ant Design icons
    CheckCircleOutlined, // Success icon
    CopyOutlined, // Copy icon
    ExclamationCircleOutlined, // Error icon
    ExpandOutlined, FullscreenExitOutlined, // Expand/collapse icons
} from "@ant-design/icons";
import Relationships from "@features/data/relationships"; // Relationships component
import Attributes from "@features/data/attributes"; // Attributes component
import {useShapeStore} from "@state/shape"; // Shape state management
import Enforcement from "@features/enforcement/enforcement"; // Enforcement panel
import GuidedTour from '@components/guided-tour'; // Guided tour component
3
function Output(props) { // Main output component
    const schemaEditorRef = useRef(null); // Schema editor reference
    const relationshipsAndAttributesRef = useRef(null); // Relationships and attributes reference
    const enforcementRef = useRef(null); // Enforcement panel reference
 // Component state
    const [dataSelected, setDataSelected] = useState('relationships'); // Selected data tab
    const [schemaSelected, setSchemaSelected] = useState('schema');
    const [isOpen, setIsOpen] = useState(false);

    const {runAssertions, runLoading, scenariosError, assertionCount} = useShapeStore();

    const {schema, yamlValidationErrors} = useShapeStore();

    const onDataSelectedChange = ({target: {value}}) => {
        setDataSelected(value);
    };

    const onSchemaSelectedChange = ({target: {value}}) => {
        setSchemaSelected(value);
    };

    const [isSchemaCopied, setIsSchemaCopied] = useState(false);

    function copySchema(text) {
        if ('clipboard' in navigator) {
            setIsSchemaCopied(true)
            return navigator.clipboard.writeText(JSON.stringify(text));
        } else {
            return document.execCommand('copy', true, text);
        }
    }

    const renderDataComponent = () => {
        switch (dataSelected) {
            case "relationships":
                return <Relationships/>;
            case "attributes":
                return <Attributes/>;
            default:
                return null;
        }
    }

    const renderSchemaComponent = () => {
        switch (schemaSelected) {
            case "schema":
                return <Schema/>;
            case "visualizer":
                return <Visualizer/>;
            default:
                return null;
        }
    }

    const [allotmentStatus, setAllotmentStatus] = React.useState("default");

    const open = () => {
        setAllotmentStatus("open")
        setIsOpen(!isOpen)
    };

    const reset = () => {
        setAllotmentStatus("default")
        setIsOpen(!isOpen)
    };

    return (
        <div>
            {!props.loading &&
                <>
                    { props.type === 'f' ?
                        <Card style={{border: 'none'}} title={
                            <Radio.Group defaultValue="schema" buttonStyle="solid"
                                         onChange={onSchemaSelectedChange}
                                         value={schemaSelected} optionType="button">
                                <Radio.Button value="schema">Schema</Radio.Button>
                                <Radio.Button value="visualizer">Visualizer</Radio.Button>
                            </Radio.Group>
                        }>
                            {renderSchemaComponent()}
                        </Card>
                        :
                        <>
                            <GuidedTour refSchemaEditor={schemaEditorRef}
                                        refRelationshipsAndAttributes={relationshipsAndAttributesRef}
                                        refEnforcement={enforcementRef}/>
                            <div style={{height: '100vh'}} className="ml-10 mr-10">
                                <Allotment vertical defaultSizes={[100, 100]}>
                                    <Allotment.Pane snap visible={!isOpen}>
                                        <Allotment>
                                            <Allotment.Pane snap ref={schemaEditorRef}>
                                                <Card title={
                                                    <Radio.Group defaultValue="schema" buttonStyle="solid"
                                                                 onChange={onSchemaSelectedChange}
                                                                 value={schemaSelected} optionType="button">
                                                        <Radio.Button value="schema">Schema</Radio.Button>
                                                        <Radio.Button value="visualizer">Visualizer</Radio.Button>
                                                    </Radio.Group>
                                                } className="mr-10" extra={<Button onClick={() => {
                                                    copySchema(schema)
                                                }} icon={<CopyOutlined/>}>{isSchemaCopied ? 'Copied!' : 'Copy'}</Button>}>
                                                    {renderSchemaComponent()}
                                                </Card>
                                            </Allotment.Pane>
                                            <Allotment.Pane snap ref={enforcementRef}>
                                                <Card title="Enforcement" className="ml-10"
                                                      extra={<div style={{display: 'flex', alignItems: 'center'}}>
                                                          <Button
                                                              disabled={yamlValidationErrors.length > 0}
                                                              icon={assertionCount === 0 ? null : scenariosError.length > 0 ?
                                                                  <ExclamationCircleOutlined/> :
                                                                  <CheckCircleOutlined/>}
                                                              type="primary"
                                                              loading={runLoading}
                                                              onClick={() => {
                                                                  runAssertions()
                                                              }}>Run</Button>
                                                      </div>}>
                                                    <Enforcement/>
                                                </Card>
                                            </Allotment.Pane>
                                        </Allotment>
                                    </Allotment.Pane>
                                    <Allotment.Pane snap>
                                        <Card title={
                                            <Radio.Group
                                                defaultValue="relationships"
                                                buttonStyle="solid"
                                                onChange={onDataSelectedChange}
                                                value={dataSelected}
                                                ref={relationshipsAndAttributesRef}
                                            >
                                                <Radio.Button value="relationships">Relationships</Radio.Button>
                                                <Radio.Button value="attributes">Attributes</Radio.Button>
                                            </Radio.Group>} className="mt-10" extra={
                                            allotmentStatus === "default" ?
                                                <Button className="ml-auto" icon={<ExpandOutlined/>} onClick={open}/>
                                                :
                                                <Button className="ml-auto" icon={<FullscreenExitOutlined/>}
                                                        onClick={reset}/>
                                        }>
                                            {renderDataComponent()}
                                        </Card>
                                    </Allotment.Pane>
                                </Allotment>
                            </div>
                        </>
                    }
                </>
            }
        </div>
    );
}


export default Output;
