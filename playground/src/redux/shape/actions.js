import {SET_SCHEMA, SET_RELATIONSHIPS, SET_ASSERTIONS, ADD_RELATIONSHIPS, ADD_ASSERTIONS} from "./types";

export const setSchema = (payload) => {
    return {
        type: SET_SCHEMA,
        payload: payload,
    };
};

export const setRelationships = (payload) => {
    return {
        type: SET_RELATIONSHIPS,
        payload: payload,
    };
};

export const addRelationships = (payload) => {
    return {
        type: ADD_RELATIONSHIPS,
        payload: payload,
    };
};

export const setAssertions = (payload) => {
    return {
        type: SET_ASSERTIONS,
        payload: payload,
    };
};

export const addAssertions = (payload) => {
    return {
        type: ADD_ASSERTIONS,
        payload: payload,
    };
};