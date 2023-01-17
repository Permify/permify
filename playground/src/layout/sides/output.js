import React from 'react'
import {Allotment} from "allotment";
import "allotment/dist/style.css";
import Enforcement from "./enforcement";
import AuthorizationData from "./authorizationData";
import AuthorizationModel from "./authorizationModel";
import Visualizer from "./visualizer";
import {shallowEqual, useSelector} from "react-redux";
import {Spin} from "antd";

function Output() {

    const ref = React.useRef(null);
    const loading = useSelector((state) => state.common.is_loading, shallowEqual);

    return (

        <Allotment defaultSizes={[130, 120]}>
            <Allotment.Pane>
                <Allotment vertical defaultSizes={[180, 120]}>
                    <Allotment.Pane snap>
                        <Spin spinning={loading}>
                            <AuthorizationModel/>
                        </Spin>
                    </Allotment.Pane>
                    <Allotment.Pane snap>
                        <Spin spinning={loading}>
                            <AuthorizationData/>
                        </Spin>
                    </Allotment.Pane>
                </Allotment>
            </Allotment.Pane>
            <Allotment.Pane snap>
                <Allotment vertical defaultSizes={[180, 120]} ref={ref}>
                    <Allotment.Pane snap>
                        <Spin spinning={loading}>
                            <Visualizer/>
                        </Spin>
                    </Allotment.Pane>
                    <Allotment.Pane snap>
                            <Enforcement/>
                    </Allotment.Pane>
                </Allotment>
            </Allotment.Pane>
        </Allotment>
    );
}


export default Output;
