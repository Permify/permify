import React, {Suspense} from "react"
import {Route} from "react-router-dom"
import {ContextLayout} from "../utility/context/Layout"
import {Skeleton} from 'antd';

const AppRoute = (
    {
        component: Component,
        fullLayout,
        privateRoute,
        ...rest
    }) => {
    return (
        <Route
            {...rest}
            render={props => {
                return (
                    <ContextLayout.Consumer>
                        {context => {
                            let LayoutTag = fullLayout === true ? context.fullLayout : context.mainLayout;
                            return (
                                <LayoutTag {...props}>
                                    <Suspense fallback={<Skeleton active  />}>
                                        <Component {...props} />
                                    </Suspense>
                                </LayoutTag>
                            )
                        }}
                    </ContextLayout.Consumer>
                )
            }}
        />
    )

};

export default AppRoute
