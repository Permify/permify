import React from 'react'
import {toAbsoluteUrl} from "../utility/helpers/asset";
import SVG from "react-inlinesvg";

function P404() {
    return (
        <div className="center-of-screen">
            <SVG src={toAbsoluteUrl("/media/svg/bg/404.svg")}/>

            <p className="mt-12 text-primary font-w-500" style={{fontSize: "45px"}}>
                Oopps, Not Found!
            </p>
        </div>
    );
}

export default P404;
