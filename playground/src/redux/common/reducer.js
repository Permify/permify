import {
    SET_IS_LOADING,
    SET_MODEL_CHANGE_TRIGGER,
} from "./types";

const INIT_STATE = {
    model_change_toggle: false,
    is_loading: true,
};

export default (state = INIT_STATE, action) => {
    switch (action.type) {
        case SET_MODEL_CHANGE_TRIGGER: {
            return {
                ...state,
                model_change_toggle: action.payload,
            };
        }
        case SET_IS_LOADING: {
            return {
                ...state,
                is_loading: action.payload,
            };
        }
        default:
            return state;
    }
};
