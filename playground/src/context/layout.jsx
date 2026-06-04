import React from "react"
import MainLayout from "@layout/main-layout"
import FullPageLayout from "@layout/full-page-layout"

const layouts = {
    main: MainLayout,
    full: FullPageLayout,
};

const ContextLayout = React.createContext();

class Layout extends React.Component {
    render() {
        const { children } = this.props;
        return (
            <ContextLayout.Provider
                value={{
                    fullLayout: layouts["full"],
                    mainLayout: layouts["main"],
                }}
            >
                {children}
            </ContextLayout.Provider>
        )
    }
}

export { Layout, ContextLayout }
