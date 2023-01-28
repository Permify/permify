export function WriteTuple(tuple) {
    return new Promise((resolve) => {
        let res = window.writeTuple(JSON.stringify(tuple), "")
        resolve(res);
    });
}

export function ReadTuples(filter) {
    return new Promise((resolve) => {
        let res = window.readTuple(JSON.stringify(filter))
        resolve(res);
    });
}

export function DeleteTuple(tuple) {
    return new Promise((resolve) => {
        let res = window.deleteTuple(JSON.stringify(tuple))
        resolve(res);
    });
}