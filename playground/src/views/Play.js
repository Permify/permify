import React, {Suspense, useEffect, useState} from 'react'
import {useLocation, useNavigate, useParams} from "react-router-dom";
import yaml from "js-yaml";
import axios from "axios";
import {shallowEqual, useDispatch, useSelector} from "react-redux";
import {Skeleton} from "antd";
import {ContextLayout} from "../utility/context/Layout";
import PlayOutput from "../layout/sides/outputs/playOutput";
import FrameOutput from "../layout/sides/outputs/frameOutput";
import {setRelationships, setSchema} from "../redux/shape/actions";

const client = axios.create();

function Play(props) {
    let {typ} = useParams();

    let location = useLocation();
    let navigate = useNavigate();

    const dispatch = useDispatch();
    const shape = useSelector((state) => state.shape, shallowEqual);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        setLoading(true)
        const params = new URLSearchParams(location.search);
        if (params.has('s')) {
            let search = params.get('s');
            client.get(`https://s3.amazonaws.com/permify.playground.storage/shapes/${search}.yaml`).then((response) => {
                return yaml.load(response.data, null)
            }).then((result) => {
                if (result.schema !== null) {
                    dispatch(setSchema(result.schema))
                }else {
                    dispatch(setSchema(``))
                    dispatch(setRelationships([]))
                }
                if (result.relationships !== null){
                    dispatch(setRelationships(result.relationships))
                }else {
                    dispatch(setRelationships([]))
                }
                setLoading(false)
            }).catch((error) => {
                navigate('/404')
            });
        } else {
            window.location = window.location.href.split('?')[0] + `?s=p`
        }
    }, []);

    return (
        <ContextLayout.Consumer>
            {context => {
                if (typ === 'f') {
                    let LayoutTag = context.fullLayout;
                    return (
                        <LayoutTag>
                            <Suspense fallback={<Skeleton active/>}>
                                <FrameOutput loading={loading} shape={shape}/>
                            </Suspense>
                        </LayoutTag>
                    )
                } else {
                    let LayoutTag = context.mainLayout;
                    return (
                        <LayoutTag>
                            <Suspense fallback={<Skeleton active/>}>
                                <PlayOutput loading={loading} shape={shape}/>
                            </Suspense>
                        </LayoutTag>
                    )
                }
            }}
        </ContextLayout.Consumer>
    )
}

export default Play;