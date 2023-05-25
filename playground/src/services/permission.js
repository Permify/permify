export function CheckPermission(entity, permission, subject) {
    return new Promise((resolve) => {
        let q = JSON.stringify({
            entity: entity,
            permission: permission,
            subject: subject,
        })
        let res = window.check(q)
        resolve(res);
    });
}

export function FilterEntity(entityType, permission, subject) {
    return new Promise((resolve) => {
        let q = JSON.stringify({
            entity_type: entityType,
            permission: permission,
            subject: subject,
        })
        let res = window.lookupEntity(q)
        resolve(res);
    });
}

export function FilterSubject(entity, permission, subject_reference) {
    return new Promise((resolve) => {
        let q = JSON.stringify({
            entity: entity,
            permission: permission,
            subject_reference: subject_reference,
        })
        let res = window.lookupSubject(q)
        resolve(res);
    });
}
