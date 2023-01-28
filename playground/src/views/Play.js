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


const def = {
    schema: `entity user {}
            
entity organization {
    
    // organizational roles
    relation admin @user
    relation member @user
        
}
    
entity repository {
    
    // represents repositories parent organization
    relation parent @organization
        
    // represents owner of this repository
    relation owner  @user
        
    // permissions
    action edit   = parent.admin or owner
    action delete = owner
        
}`,
    relationships: [
        "repository:1#owner@user:1"
    ],
    assertions: {
        "can user:2 push repo:1": true
    }
}

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

        dispatch(setSchema(``))
        dispatch(setRelationships([]))

        const params = new URLSearchParams(location.search);
        if (params.has('s')) {
            let search = params.get('s');
            client.get(`https://s3.amazonaws.com/permify.playground.storage/shapes/${search}.yaml`).then((response) => {
                let data = yaml.load(response.data, null)
                if (data.schema !== null) {
                    dispatch(setSchema(data.schema))
                }else {
                    dispatch(setSchema(def.schema))
                    dispatch(setRelationships(def.relationships))
                }
                if (data.relationships !== null){
                    dispatch(setRelationships(data.relationships))
                }else {
                    dispatch(setRelationships([]))
                }
                setLoading(false)
            }).catch((error) => {
                navigate('/404')
            });
        } else {
            dispatch(setSchema(def.schema))
            dispatch(setRelationships(def.relationships))
            setLoading(false)
        }
    }, [location.search]);

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