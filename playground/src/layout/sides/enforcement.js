import React, {useEffect, useState} from 'react'
import YamlEditor from "@pkg/Editor/yaml";
import "allotment/dist/style.css";
import {useShapeStore} from "@state/shape";
import yaml, {dump} from 'js-yaml';
import {Alert} from "antd";

function Enforcement() {

    const {scenarios, setScenarios, scenariosError, assertionCount, runLoading, yamlValidationErrors} = useShapeStore();
    const [yamlData, setYamlData] = useState("");

    const handleYamlChange = (newCode) => {
        try {
            const updatedData = yaml.load(newCode);
            setScenarios(updatedData);
        } catch (error) {
            console.error("Error updating scenario with new YAML:", error);
        }
    };

    useEffect(() => {
        setYamlData(dump(scenarios))
    }, [scenarios]);

    return (
        <>
            {!runLoading && scenariosError && scenariosError.length > 0 && (
                scenariosError.map((error, index) => (
                    <Alert type="error" message={error.message} banner closable/>
                ))
            )}
            {!runLoading && assertionCount > 0 && scenariosError.length === 0 && yamlValidationErrors.length === 0 &&  (
                <Alert type="success" message="Success" banner closable/>
            )}
            <YamlEditor
                code={yamlData}
                setCode={(newCode) => handleYamlChange(newCode)}
            />
        </>

    );
}

export default Enforcement
