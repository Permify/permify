export function WriteSchema(schema) {
    return new Promise((resolve) => {
        let res = window.writeSchema(schema)
        resolve(res);
    });
}

export function ReadSchema() {
    return new Promise((resolve) => {
        let res = window.readSchema("")
        resolve(res);
    });
}

export function ReadSchemaGraph() {
    return new Promise((resolve) => {
        let res = window.readSchemaGraph("")
        resolve(res);
    });
}