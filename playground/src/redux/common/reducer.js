import {
    SET_MODEL_CHANGE_TRIGGER,
} from "./types";

const INIT_STATE = {
    model_change_toggle: false,
};

export default (state = INIT_STATE, action) => {
    switch (action.type) {
        case SET_MODEL_CHANGE_TRIGGER: {
            return {
                ...state,
                model_change_toggle: action.payload,
            };
        }
        default:
            return state;
    }
};
