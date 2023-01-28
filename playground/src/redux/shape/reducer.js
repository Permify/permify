import {ADD_ASSERTIONS, ADD_RELATIONSHIPS, SET_ASSERTIONS, SET_RELATIONSHIPS, SET_SCHEMA} from "./types";

const INIT_STATE = {
    schema: ``,
    relationships: [],
    assertions: []
};

export default (state = INIT_STATE, action) => {
    switch (action.type) {
        case SET_SCHEMA: {
            return {
                ...state,
                schema: action.payload,
            };
        }
        case SET_RELATIONSHIPS: {
            return {
                ...state,
                relationships: action.payload,
            };
        }
        case ADD_RELATIONSHIPS: {
            return {
                ...state,
                relationships: [...state.relationships, action.payload],
            };
        }
        case SET_ASSERTIONS: {
            return {
                ...state,
                assertions: action.payload,
            };
        }
        case ADD_ASSERTIONS: {
            return {
                ...state,
                assertions: [...state.assertions, action.payload],
            };
        }
        default:
            return state;
    }
};
