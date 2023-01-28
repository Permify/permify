import React from 'react'
import {Allotment} from "allotment";
import "allotment/dist/style.css";
import Enforcement from "../enforcement";
import AuthorizationData from "../authorizationData";
import AuthorizationModel from "../authorizationModel";
import Visualizer from "../visualizer";
import {Spin, Tooltip} from "antd";
import {InfoCircleOutlined} from "@ant-design/icons";

function PlayOutput(props) {
    const ref = React.useRef(null);

    return (
        <Allotment defaultSizes={[130, 120]}>
            <Allotment.Pane>
                <Allotment vertical defaultSizes={[180, 120]}>
                    <Allotment.Pane snap>
                        <Spin spinning={props.loading}>
                            <div style={{marginRight: "10px", marginBottom: "10px"}}>
                                <AuthorizationModel title={
                                    <div>
                                        <span className="mr-8">Authorization Model</span>
                                        <Tooltip placement="right" color="black"
                                                 title={"Permify has its own language that you can model your authorization logic with it, we called it Permify Schema. You can define your entities, relations between them and access control decisions with using Permify Schema."}>
                                            <InfoCircleOutlined/>
                                        </Tooltip>
                                    </div>
                                } initialValue={props.shape.schema} hidden={false}/>
                            </div>
                        </Spin>
                    </Allotment.Pane>
                    <Allotment.Pane snap>
                        <Spin spinning={props.loading}>
                            <div style={{marginRight: "12px", marginTop: "12px"}}>
                                <AuthorizationData title={
                                    <div>
                                        <span className="mr-8">Authorization Data</span>
                                        <Tooltip placement="right" color="black"
                                                 title={"Authorization data stored as Relation Tuples into your preferred database. These relational tuples represents your authorization data."}>
                                            <InfoCircleOutlined/>
                                        </Tooltip>
                                    </div>
                                } initialValue={props.shape.relationships} hidden={false}/>
                            </div>
                        </Spin>
                    </Allotment.Pane>
                </Allotment>
            </Allotment.Pane>
            <Allotment.Pane snap>
                <Allotment vertical defaultSizes={[180, 120]} ref={ref}>
                    <Allotment.Pane snap>
                        <Spin spinning={props.loading}>
                            <div style={{marginLeft: "12px"}}>
                                <Visualizer title={"Visualizer"} hidden={false}/>
                            </div>
                        </Spin>
                    </Allotment.Pane>
                    <Allotment.Pane snap>
                        <div style={{marginLeft: "12px", marginTop: "12px"}}>
                            <Enforcement title={""} hidden={false}/>
                        </div>
                    </Allotment.Pane>
                </Allotment>
            </Allotment.Pane>
        </Allotment>
    );
}


export default PlayOutput;
