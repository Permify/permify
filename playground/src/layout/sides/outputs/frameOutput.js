import React, {useState} from 'react'
import AuthorizationModel from "../../../layout/sides/authorizationModel";
import {Button, Radio} from 'antd';
import AuthorizationData from "../../../layout/sides/authorizationData";
import Visualizer from "../../../layout/sides/visualizer";
import Enforcement from "../../../layout/sides/enforcement";
import {ShareAltOutlined} from "@ant-design/icons";
import {nanoid} from "nanoid";
import yaml from "js-yaml";
import Upload from "../../../services/s3";
import {shallowEqual, useSelector} from "react-redux";
import Share from "../../components/Modals/Share";

function FrameOutput(props) {
    const [selected, setSelected] = useState('model');
    const shape = useSelector((state) => state.shape, shallowEqual);

    const [shareModalVisibility, setShareModalVisibility] = React.useState(false);
    const [id, setId] = useState("");

    const [isModelReady, setIsModelReady] = useState(false);

    function isReady(value) {
        setIsModelReady(value)
    }

    const toggleShareModalVisibility = () => {
        setShareModalVisibility(!shareModalVisibility);
    };

    const onChange = (e) => {
        setSelected(e.target.value);
    };

    const share = () => {
        let id = nanoid()
        setId(id)
        const yamlString = yaml.dump({
            schema: shape.schema,
            relationships: shape.relationships,
            assertions: shape.assertions
        })
        const file = new File([yamlString], `shapes/${id}.yaml`, {type : 'text/x-yaml'});
        Upload(file).then((res) => {
            toggleShareModalVisibility()
        })
    }

    return (
        <>
            <Share visible={shareModalVisibility} toggle={toggleShareModalVisibility} id={id}></Share>

            <div className="ml-12 mr-12">
                <div className="mt-12 mb-12">
                    <Radio.Group defaultValue="model" buttonStyle="solid" onChange={onChange}>
                        <Radio.Button value="model">Model</Radio.Button>
                        <Radio.Button value="data">Data</Radio.Button>
                        <Radio.Button value="visualizer">Visualizer</Radio.Button>
                        <Radio.Button value="enforcement">Enforcement</Radio.Button>
                    </Radio.Group>
                    <Button style={{float: 'right'}} type="secondary" onClick={() => {
                        share()
                    }} icon={<ShareAltOutlined/>}>Share</Button>
                </div>
                {!props.loading &&
                    <>
                        <AuthorizationModel title="Model" hidden={selected !== 'model'} isReady={isReady} initialValue={props.shape.schema}/>
                        <AuthorizationData title="Data" hidden={selected !== 'data'} isModelReady={isModelReady} initialValue={props.shape.relationships}/>
                        <Visualizer title="Visualizer" hidden={selected !== 'visualizer'}/>
                        <Enforcement title="Enforcement" hidden={selected !== 'enforcement'}/>
                    </>
                }
            </div>
        </>
    );
}

export default FrameOutput;
