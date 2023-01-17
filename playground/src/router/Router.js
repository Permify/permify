import React, {lazy} from "react"
import {Router, Switch, Redirect} from "react-router-dom"
import AppRoute from "../router/RouteConfig"
import {history} from "../utility/shared/history"

// Route-based code splitting
const Play = lazy(() =>
    import("../views/Play")
);

const P404 = lazy(() =>
    import("../views/404")
);

export default function AppRouter() {
    return (
        <Router history={history}>
            <Switch>
                <AppRoute
                    exact
                    path="/"
                    component={Play}
                    privateRoute
                />
                <AppRoute
                    exact
                    path="/404"
                    component={P404}
                    fullLayout
                />
                <Redirect to="/404"/>
            </Switch>
        </Router>
    );
}
