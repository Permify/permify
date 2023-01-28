export function CheckPermission(entity, permission, subject){
    return new Promise((resolve) => {
        let q = JSON.stringify({
            entity: entity,
            permission: permission,
            subject: subject,
        })
        let res = window.check(q, "")
        resolve(res);
    });
}

export function FilterData(entityType, permission, subject) {
    return new Promise((resolve) => {
        let q = JSON.stringify({
            entity_type: entityType,
            permission: permission,
            subject: subject,
        })
        let res = window.lookupEntity(q, "")
        resolve(res);
    });
}