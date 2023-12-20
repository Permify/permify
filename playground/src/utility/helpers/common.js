export default function TupleToHumanLanguage(tuple) {
    const splitChar = '@';
    const subSplitChar = '#';
    const dots = '...';

    let [entity, subject] = tuple.split(splitChar);

    if (!entity || !subject) {
        return "";
    }

    let subjectComponents = subject.split(subSplitChar);
    let entityComponents = entity.split(subSplitChar);

    if (subjectComponents.length === 0 || entityComponents.length !== 2) {
        return "";
    }

    let subjectString = constructSubjectString(subjectComponents, dots);
    let entityString = `${entityComponents[1]} of ${entityComponents[0]}`;

    return `${subjectString}${entityString}`;
}

function constructSubjectString(subjectComponents, dots) {
    if (subjectComponents.length === 1) {
        return `${subjectComponents[0]} is `;
    }

    if (subjectComponents[1] === dots) {
        return `${subjectComponents[0]} is `;
    }

    return `${subjectComponents[1]} of ${subjectComponents[0]} is `;
}

export function RelationshipObjectToKey(tuple) {
    return `${tuple.entityType}:${tuple.entityID}#${tuple.relation}@${tuple.subjectType}:${tuple.subjectID}${tuple.subjectRelation === undefined ? "" : "#" + tuple.subjectRelation}`
}

export function AttributeObjectToKey(attribute) {
    return `${attribute.entityType}:${attribute.entityID}$${attribute.attribute}|${attribute.type}:${attribute.value}`
}

export function AttributeEntityToKey(attribute) {
    return `${attribute.entityType}:${attribute.entityID}$${attribute.attribute}`
}

export function StringRelationshipsToObjects(relationships) {
    let r = []
    for (let i = 0; i < relationships.length; i++) {
        r[i] = StringRelationshipToObject(relationships[i])
    }
    return r
}

export function StringRelationshipToObject(relationship) {
    let s = relationship.trim().split("@");
    if (s.length !== 2) {
        return null;
    }
    let ear = entityAndRelation(s[0]);
    let sub = entityAndRelation(s[1]);

    return {
        key: relationship,
        entityType: ear.entity.type,
        entityID: ear.entity.id,
        relation: ear.relation,
        subjectType: sub.entity.type,
        subjectID: sub.entity.id,
        subjectRelation: sub.relation,
    }
}

export function StringAttributesToObjects(attributes) {
    let r = []
    for (let i = 0; i < attributes.length; i++) {
        r[i] = StringAttributeToObject(attributes[i])
    }
    return r
}

// StringAttributeToObject
export function StringAttributeToObject(attribute) {
    let s = attribute.trim().split("|");
    if (s.length !== 2) {
        return null;
    }
    let ear = entityAndAttribute(s[0]);
    let val = value(s[1]);

    return {
        key: attribute,
        entityType: ear.entity.type,
        entityID: ear.entity.id,
        attribute: ear.attribute,
        type: val.type,
        value: val.value,
    }
}

// EAR EntitiesAndRelation
function entityAndRelation(ear) {
    let s = ear.trim().split("#");
    if (s.length === 1) {
        let e = E(s[0]);
        return {
            entity: e,
            relation: "",
        };
    } else if (s.length === 2) {
        let e = E(s[0]);
        return {
            entity: e,
            relation: s[1],
        };
    } else {
        return null;
    }
}

// entityAndAttribute
function entityAndAttribute(ear) {
    let s = ear.trim().split("$");
    if (s.length === 1) {
        let e = E(s[0]);
        return {
            entity: e,
            attribute: "",
        };
    } else if (s.length === 2) {
        let e = E(s[0]);
        return {
            entity: e,
            attribute: s[1],
        };
    } else {
        return null;
    }
}

// value
function value(v) {
    let s = v.trim().split(":");
    if (s.length !== 2) {
        return null;
    }
    return {
        type: s[0],
        value: s[1],
    };
}

// E New Entities from string
function E(e) {
    let s = e.trim().split(":");
    if (s.length !== 2) {
        return null;
    }
    return {
        type: s[0],
        id: s[1],
    };
}
