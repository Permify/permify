import React, {Suspense} from "react"
import {BrowserRouter, Navigate, Route, Routes} from "react-router-dom"
import Play from "../views/Play";
import P404 from "../views/404";
import {ContextLayout} from "../utility/context/Layout";
import {Skeleton} from "antd";

export default function AppRouter() {
    return (
        <BrowserRouter>
            <Routes>
                <Route
                    path=":typ"
                    element={<Play/>}
                />
                <Route
                    path="/"
                    element={<Play/>}
                />
                <Route
                    path="404"
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
