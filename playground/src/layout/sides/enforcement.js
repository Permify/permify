import React from 'react'
import YamlEditor from "../../pkg/Editor/yaml";
import "allotment/dist/style.css";
import {useShapeStore} from "../../state/shape";
import yaml, {dump} from 'js-yaml';
import {CheckCircleOutlined, ExclamationCircleOutlined} from "@ant-design/icons";

function Enforcement() {

    const {scenarios, setScenarios, scenariosError, assertionCount} = useShapeStore();
    const yamlData = dump(scenarios);

    const handleYamlChange = (newCode) => {
        try {
            const updatedData = yaml.load(newCode);
            setScenarios(updatedData);
        } catch (error) {
            console.error("Error updating scenario with new YAML:", error);
        }
    };

    return (
        <>
            {scenariosError && scenariosError.length > 0 && (
                <div style={{
                    padding: '5px',
                    borderRadius: '0',
                    background: 'rgba(255,0,0,0.1)'
                }}>
                    {scenariosError.map((error, index) => (
                        <div key={index} style={{color: 'red'}}>
                            <ExclamationCircleOutlined/> {error.message}
                        </div>
                    ))}
                </div>
            )}
            {assertionCount > 0 && scenariosError.length < 1 && (
                <div style={{
                    padding: '5px',
                    borderRadius: '0',
                    background: 'rgba(78,223,67,0.1)'
                }}>
                    <div style={{color: '#4edf43'}}>
                       <CheckCircleOutlined/> Success
                    </div>
                </div>
            )}
            <YamlEditor
                code={yamlData}
                setCode={(newCode) => handleYamlChange(newCode)}
            />
        </>
    );
}

export default Enforcement
