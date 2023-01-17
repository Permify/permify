import React from "react"

const FullPageLayout = ({ children, ...rest }) => {
    return (
        <div className="center-of-screen">
            {children}
        </div>
    )
};

export default FullPageLayout
