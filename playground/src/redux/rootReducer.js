import {combineReducers} from "redux";
import {all} from "redux-saga/effects";
import {persistReducer} from 'redux-persist'

import storage from 'redux-persist/lib/storage'

//reducers
import commonReducer from "./common/reducer";

const commonPersistConfig = {
    key: 'common',
    storage: storage,
};

export const rootReducer = combineReducers({
    common: persistReducer(commonPersistConfig, commonReducer),
});

export function* rootSaga() {
    yield all([]);
}
