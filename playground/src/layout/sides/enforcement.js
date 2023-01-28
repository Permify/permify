import React, {useState} from 'react'
import {Card, Radio} from 'antd';
import Check from "./particials/check";
import DataFiltering from "./particials/dataFiltering";

function Enforcement(props) {

    const [selected, setSelected] = useState('check');

    const onChange = ({target: {value}}) => {
        setSelected(value);
    };

    return (
        <Card title={props.title} className="h-screen" style={{display: props.hidden && 'none'}}>
            <div className="p-12">
                <Radio.Group defaultValue="check" onChange={onChange} value={selected}>
                    <Radio value="check">Check</Radio>
                    <Radio value="data-filtering">Data Filtering</Radio>
                </Radio.Group>
                {selected === "check" ?
                    <Check/>
                    :
                    <DataFiltering/>
                }
            </div>
        </Card>
    )
}

export default Enforcement
