import React, {Suspense, useEffect, useState} from 'react'
import {useLocation} from "react-router-dom";
import {Skeleton} from "antd";
import {ContextLayout} from "@utility/context/Layout";
import Output from "@layout/sides/output";
import {useShapeStore} from "@state/shape";

function Play() {
    const { fetchShape } = useShapeStore();

    let location = useLocation();

    const [type, setType] = useState("");

    const [loading, setLoading] = useState(true);

    useEffect(() => {
        setLoading(true);

        const params = new URLSearchParams(location.search);

        if (params.has("t")) {
            setType(params.get('t'));
        }

        if (params.has('s')) {
            const search = params.get('s');
            fetchShape(search).finally(() => setLoading(false));
        } else {
            const baseUrl = window.location.href.split('?')[0];
            window.location = `${baseUrl}?s=organizations-hierarchies`;
        }
    }, []);

    return (
        <ContextLayout.Consumer>
            {
                context => {
                    const LayoutTag = (type === 'f') ? context.fullLayout : context.mainLayout;
                    return (
                        <LayoutTag>
                            <Suspense fallback={<Skeleton active />}>
                                <Output loading={loading} type={type} />
                            </Suspense>
                        </LayoutTag>
                    );
                }
            }
        </ContextLayout.Consumer>
    )
}

export default Play;
