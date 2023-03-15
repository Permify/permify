import { configureStore, getDefaultMiddleware } from '@reduxjs/toolkit'
import createSagaMiddleware from "redux-saga";
import {reduxBatch} from "@manaflair/redux-batch";
import {rootReducer, rootSaga} from "./rootReducer";

const sagaMiddleware = createSagaMiddleware();

const middleware = [
    ...getDefaultMiddleware({
        immutableCheck: false,
        serializableCheck: false,
        thunk: true
    }),
    sagaMiddleware,
];

const store = configureStore({
    reducer: rootReducer,
    middleware,
    devTools: process.env.NODE_ENV !== "production",
    enhancers: [reduxBatch]
});

sagaMiddleware.run(rootSaga);

export default store;
