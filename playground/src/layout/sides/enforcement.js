import React, {useState} from 'react'
import {Card, Radio} from 'antd';
import Check from "./particials/check";
import EntityFilter from "./particials/entityFilter";
import SubjectFilter from "./particials/subjectFilter";

function Enforcement(props) {

    const [selected, setSelected] = useState('check');

    const onChange = ({target: {value}}) => {
        setSelected(value);
    };

    const renderComponent = () => {
        switch (selected) {
            case "check":
                return <Check />;
            case "entity-filter":
                return <EntityFilter />;
            case "subject-filter":
                return <SubjectFilter />;
            default:
                return null;
        }
    }

    return (
        <Card title={props.title} className="h-screen" style={{display: props.hidden && 'none'}}>
            <div className="p-12">
                <Radio.Group defaultValue="check" onChange={onChange} value={selected}>
                    <Radio value="check">Check</Radio>
                    <Radio value="entity-filter">Entity Filter</Radio>
                    <Radio value="subject-filter">Subject Filter</Radio>
                </Radio.Group>
                {renderComponent()}
            </div>
        </Card>
    )
}

export default Enforcement
