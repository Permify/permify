import React, {useEffect, useState} from 'react'
import YamlEditor from "@lib/editor/yaml";
import "allotment/dist/style.css";
import {useShapeStore} from "@state/shape";
import yaml, {dump} from 'js-yaml';
import {Alert} from "antd";

function Enforcement() {

    const {scenarios, setScenarios, scenariosError, assertionCount, runLoading, yamlValidationErrors} = useShapeStore();
    const [yamlData, setYamlData] = useState("");

    const handleYamlChange = (newCode) => {
        // While typing the editor text is the source of truth, so keep it as it is
        // and only feed the parsed result into the store.
        setYamlData(newCode);
        try {
            const updatedData = yaml.load(newCode);
            setScenarios(updatedData);
        } catch (error) {
            console.error("Error updating scenario with new YAML:", error);
        }
    };

    useEffect(() => {
        const dumped = dump(scenarios);
        setYamlData((current) => {
            try {
                if (dump(yaml.load(current)) === dumped) {
                    return current;
                }
            } catch (error) {
                // Not valid YAML yet, the user is still typing it.
                return current;
            }
            // The scenarios changed outside of the editor (example, import, new scenario).
            return dumped;
        });
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
