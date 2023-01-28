import {combineReducers} from "redux";
import {all} from "redux-saga/effects";
import {persistReducer} from 'redux-persist'

import storage from 'redux-persist/lib/storage'

//reducers
import commonReducer from "./common/reducer";
import shapeReducer from "./shape/reducer";

// configs

const commonPersistConfig = {
    key: 'common',
    storage: storage,
};

const shapePersistConfig = {
    key: 'shape',
    storage: storage,
};

export const rootReducer = combineReducers({
    common: persistReducer(commonPersistConfig, commonReducer),
    shape: persistReducer(shapePersistConfig, shapeReducer),
});

export function* rootSaga() {
    yield all([]);
}
