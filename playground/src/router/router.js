import React, {Suspense} from "react"
import {BrowserRouter, Route, Routes} from "react-router-dom"
import Play from "@views/Play";
import P404 from "@views/404";
import {ContextLayout} from "@utility/context/Layout";
import {Skeleton} from "antd";

export default function AppRouter() {
    return (
        <BrowserRouter>
            <Routes>
                <Route
                    path="/"
                    element={<Play/>}
                />
                <Route
                    path="*"
                    element={
                        <ContextLayout.Consumer>
                            {context => {
                                let LayoutTag = context.fullLayout;
                                return (
                                    <LayoutTag>
                                        <Suspense fallback={<Skeleton active/>}>
                                            <P404/>
                                        </Suspense>
                                    </LayoutTag>
                                )
                            }}
                        </ContextLayout.Consumer>
                    }
                />
            </Routes>
        </BrowserRouter>
    );
}
