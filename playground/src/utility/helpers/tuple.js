export default function TupleToHumanLanguage(tuple) {
    const splitChar = '@';
    const subSplitChar = '#';
    const continuationIndicator = '...';

    let [entity, subject] = tuple.split(splitChar);

    if (!entity || !subject) {
        return "";
    }

    let subjectComponents = subject.split(subSplitChar);
    let entityComponents = entity.split(subSplitChar);

    if (subjectComponents.length === 0 || entityComponents.length !== 2) {
        return "";
    }

    let subjectString = constructSubjectString(subjectComponents, continuationIndicator);
    let entityString = `${entityComponents[1]} of ${entityComponents[0]}`;

    return `${subjectString}${entityString}`;
}

function constructSubjectString(subjectComponents, continuationIndicator) {
    if (subjectComponents.length === 1) {
        return `${subjectComponents[0]} is `;
    }

    if (subjectComponents[1] === continuationIndicator) {
        return `${subjectComponents[0]} is `;
    }

    return `${subjectComponents[1]} of ${subjectComponents[0]} is `;
}

export function TupleObjectToTupleString(tuple) {
    return `${tuple.entity.type}:${tuple.entity.id}#${tuple.relation}@${tuple.subject.type}:${tuple.subject.id}${tuple.subject.relation === undefined ? "" : "#" + tuple.subject.relation}`
}

// Tuple -
export function Tuple(tuple) {
    let s = tuple.trim().split("@");
    if (s.length !== 2) {
        return null;
    }
    let ear = EAR(s[0]);
    let sub = EAR(s[1]);
    return {
        entity: ear.entity,
        relation: ear.relation,
        subject: {
            type: sub.entity.type,
            id: sub.entity.id,
            relation: sub.relation,
        },
    };
}

// EAR EntitiesAndRelation
function EAR(ear) {
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
