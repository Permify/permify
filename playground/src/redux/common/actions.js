import {SET_IS_LOADING, SET_MODEL_CHANGE_TRIGGER} from "./types";

export const setModelChangeActivity = (payload) => {
    return {
        type: SET_MODEL_CHANGE_TRIGGER,
        payload: payload,
    };
};

export const setIsLoading = (payload) => {
    return {
        type: SET_IS_LOADING,
        payload: payload,
    };
};
