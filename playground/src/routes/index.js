import React, {Suspense} from "react"
import {BrowserRouter, Route, Routes} from "react-router-dom"
import Play from "@pages/play";
import NotFound from "@pages/not-found";
import {ContextLayout} from "@context/layout";
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
                                            <NotFound/>
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
