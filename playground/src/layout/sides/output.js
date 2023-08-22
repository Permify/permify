import React, {useState, useEffect} from 'react'
import "allotment/dist/style.css";
import Schema from "./schema";
import Visualizer from "./visualizer";
import {Button, Radio, Tabs, Typography} from "antd";
import {CopyOutlined} from "@ant-design/icons";
import Relationships from "./particials/data/relationships";
import Attributes from "./particials/data/attributes";
import {useSearchParams} from 'react-router-dom';
import {useShapeStore} from "../../state/shape";
import Enforcement from "./enforcement";
import NewScenario from "../components/modals/newScenario";

const {Text} = Typography;

function Output(props) {
    const [searchParams, setSearchParams] = useSearchParams();
    const initialTab = searchParams.get("tab") || "schema";
    const [selected, setSelected] = useState(initialTab);
    const [dataSelected, setDataSelected] = useState('relationships');

    const [newScenarioModalVisibility, setNewScenarioModalVisibility] = React.useState(false);

    const { run } = useShapeStore();

    const toggleNewScenarioModalVisibility = () => {
        setNewScenarioModalVisibility(!newScenarioModalVisibility);
    };

    const {schema} = useShapeStore();

    const onDataSelectedChange = ({target: {value}}) => {
        setDataSelected(value);
    };

    useEffect(() => {
        if (selected) {
            setSearchParams({...Object.fromEntries(searchParams), tab: selected});
        }
    }, [selected, setSearchParams, searchParams]);

    const handleTabChange = (key) => {
        setSelected(key);
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

    let tabBarExtra;
    switch (selected) {
        case "schema":
            tabBarExtra = (
                <Button className="mr-12" onClick={() => {
                    copySchema(schema)
                }} icon={<CopyOutlined/>}>{isSchemaCopied ? 'Copied!' : 'Copy'}</Button>
            );
            break;
        case "data":
            tabBarExtra = (
                <Radio.Group defaultValue="relationships" buttonStyle="solid" onChange={onDataSelectedChange}
                             value={dataSelected}>
                    <Radio value="relationships">Relationships</Radio>
                    <Radio value="attributes">Attributes <Text type="danger">(beta)</Text></Radio>
                </Radio.Group>
            );
            break;
        case "visualizer":
            tabBarExtra = (
                <></>
            );
            break;
        case "enforcement":
            tabBarExtra = (
                <div style={{ display: 'flex', alignItems: 'center' }}>
                    <Button className="mr-12" onClick={() => {
                        toggleNewScenarioModalVisibility()
                    }}>New Scenario</Button>
                    <Button className="mr-12" type="primary" onClick={() => {
                       run()
                    }}>Run</Button>
                </div>
            );
            break;
        default:
            tabBarExtra = null;
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

    const tabs = [
        {
            key: 'schema',
            label: 'Schema',
            children: <Schema />
        },
        {
            key: 'data',
            label: 'Data',
            children: renderDataComponent()
        },
        {
            key: 'visualizer',
            label: 'Visualizer',
            children: <Visualizer />
        },
        {
            key: 'enforcement',
            label: 'Enforcement',
            children: <Enforcement />
        }
    ];

    return (
        <div>
            {!props.loading &&
                <>
                    <NewScenario visible={newScenarioModalVisibility} toggle={toggleNewScenarioModalVisibility}></NewScenario>
                    <Tabs
                        className="custom-card-tabs"
                        activeKey={selected}
                        onChange={handleTabChange}
                        type="card"
                        items={tabs}
                        tabBarExtraContent={tabBarExtra}
                    />
                </>
            }
        </div>
    );
}


export default Output;
