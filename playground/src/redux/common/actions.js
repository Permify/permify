import {SET_MODEL_CHANGE_TRIGGER} from "./types";

export const setModelChangeActivity = (payload) => {
    return {
        type: SET_MODEL_CHANGE_TRIGGER,
        payload: payload,
    };
};
