import React, {useState} from 'react'
import {
    Button,
    Collapse, Empty,
} from "antd";
import YamlEditor from "../../pkg/Editor/yaml";
import "allotment/dist/style.css";
import {useShapeStore} from "../../state/shape";
import yaml, {dump} from 'js-yaml';
import {DeleteOutlined} from "@ant-design/icons";

const {Panel} = Collapse;

export const omitKeys = (obj, keys = []) => {
    const newObj = {...obj};
    keys.forEach(key => delete newObj[key]);
    return newObj;
};

export const convertDataToYAML = (data) => {
    return dump(data);
};

function Enforcement() {

    const {scenarios, setScenarios, scenariosError, removeScenario} = useShapeStore();

    const [activeKey, setActiveKey] = useState(null);

    const handleFormatClick = (event, index) => {
        event.stopPropagation();

        const scenario = scenarios[index];
        const dataWithoutKeys = omitKeys(scenario, ['name', 'description']);
        const formattedYaml = convertDataToYAML(dataWithoutKeys); // Use your convertDataToYAML function.

        // Update the scenarios array with the modified scenario
        const updatedScenarios = [...scenarios];
        updatedScenarios[index] = {...scenario, ...yaml.load(formattedYaml)};

        // Assuming you have a function called setScenarios to update your state
        setScenarios(updatedScenarios);
    };

    const handleRemoveClick = (event, index) => {
        event.stopPropagation();
        removeScenario(index);
    };

    const handleYamlChange = (index, newCode) => {
        try {

            let name = scenarios[index].name
            let description = scenarios[index].description

            const updatedData = yaml.load(newCode);

            // Keep the name and description fields from the original scenario
            const updatedScenario = {
                name, description,
                ...updatedData
            };

            // Update the scenarios array with the updated scenario
            const updatedScenarios = [...scenarios];
            updatedScenarios[index] = updatedScenario;

            // Set the updated scenarios (assuming you have a method for this)
            setScenarios(updatedScenarios);

        } catch (error) {
            console.error("Error updating scenario with new YAML:", error);
        }
    };

    return (
        <>
            {scenarios.length === 0 ? (
                <div style={{
                    overflow: 'auto',
                    height: 'calc(100vh - 140px)',
                    display: 'flex',
                    justifyContent: 'center',
                    alignItems: 'center'
                }}>
                    <Empty
                        image={Empty.PRESENTED_IMAGE_SIMPLE}
                        imageStyle={{
                            height: 60,
                        }}
                        description={
                            <>
                                <div>
                                    Need help creating a scenario?
                                </div>
                                <div>
                                    Check out our guidelines and examples in the <a
                                    href="https://docs.permify.co/docs/playground">docs</a>.
                                </div>
                            </>
                        }
                    >
                    </Empty>
                </div>
            ) : (
                <Collapse accordion activeKey={activeKey} onChange={(key) => setActiveKey(key)}
                          style={{overflow: 'auto', height: 'calc(100vh - 140px)'}} size="large">
                    {scenarios.map((data, index) => {
                        const yamlData = convertDataToYAML(omitKeys(data, ['name', 'description']));

                        const errorForCurrentIndex = Array.isArray(scenariosError)
                            ? scenariosError.find(error => error.key === index)
                            : null;

                        const errorMessage = errorForCurrentIndex ? errorForCurrentIndex.message : null;
                        const hasErrors = Boolean(errorMessage);

                        return (
                            <Panel
                                className={hasErrors ? 'error-row' : 'success-row'}
                                header={
                                    <div style={{
                                        display: 'flex',
                                        justifyContent: 'space-between',
                                        alignItems: 'center'
                                    }}>
                                        <div>
                                            <h4 style={{margin: 0}}>{data.name}</h4>
                                            <p style={{
                                                margin: 0,
                                                fontSize: '12px',
                                                color: 'grey'
                                            }}>{data.description}</p>
                                        </div>
                                        <div>
                                            <Button className="mr-8"
                                                    onClick={(event) => handleFormatClick(event, index)}>Format</Button>
                                            <Button type="text" danger icon={<DeleteOutlined
                                                onClick={(event) => handleRemoveClick(event, index)}/>}/>
                                        </div>
                                    </div>
                                }
                                key={index}
                            >
                                <YamlEditor
                                    code={yamlData}
                                    setCode={(newCode) => handleYamlChange(index, newCode)}
                                />
                                {hasErrors &&
                                    <div style={{
                                        color: 'red',
                                        marginTop: '10px',
                                        padding: '5px',
                                        borderRadius: '5px',
                                        background: 'rgba(255,0,0,0.1)'
                                    }}>
                                        Error: {errorMessage}
                                    </div>
                                }
                            </Panel>
                        );
                    })}
                </Collapse>
            )}
        </>
    );
}

export default Enforcement
