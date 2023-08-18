import React, {useState} from 'react'
import {
    Collapse,
} from "antd";
import YamlEditor from "../../pkg/Editor/yaml";
import Terminal from './../../pkg/XTerm';
import {Allotment} from "allotment";
import "allotment/dist/style.css";

const {Panel} = Collapse;

function Enforcement() {

    const sampleDataList = [
        {name: "Scenario 1", description: "test description", content: "entity: 'repository:1'\nsubject: 'user:1'"},
        {name: "Scenario 2", description: "test description", content: "entity: 'repository:2'\nsubject: 'user:2'"},
    ];
    // const formatYaml = () => {
    //     try {
    //         const jsonObject = yaml.load(code);
    //         const formattedYaml = yaml.dump(jsonObject, { indent: 4 });
    //         setCode(formattedYaml);
    //     } catch (e) {
    //         console.error("Error formatting YAML:", e);
    //         alert("Invalid YAML format. Please check your content.");
    //     }
    // };
    return (
        <div style={{height: 'calc(100vh - 140px)'}}>
            <Allotment vertical defaultSizes={[130, 100]}>
                <Allotment.Pane >
                    <Collapse style={{overflow: 'auto', height: 'calc(100vh - 335px)'}} size="large">
                        {sampleDataList.map((data, index) => (
                            <Panel
                                header={
                                    <div>
                                        <h4 style={{margin: 0}}>{data.name}</h4>
                                        <p style={{margin: 0, fontSize: '12px', color: 'grey'}}>{data.description}</p>
                                    </div>
                                }
                                key={index}
                            >
                                <YamlEditor
                                    initialCode={data.content}
                                    onCodeChange={(newCode) => {
                                        // Handle code changes, e.g., updating your state
                                        console.log("New YAML content:", newCode);
                                    }}
                                />
                            </Panel>
                        ))}
                    </Collapse>
                </Allotment.Pane>
                <Allotment.Pane >
                    <Terminal/>
                </Allotment.Pane>
            </Allotment>
        </div>
    );
}

export default Enforcement
