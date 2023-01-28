export default function TupleToHumanLanguage(tuple) {
    let rv = ""
    let p = tuple.split("@")
    if (p.length !== 2) {
        return rv
    }
    let e = p[0]
    let s = p[1]
    let a = s.split("#")
    if (a.length === 1) {
        rv += a[0] + " is "
    } else {
        if (a[1] === "...") {
            rv += a[0] + " is "
        } else {
            rv += a[1] + " of " + a[0] + " is "
        }
    }
    let b = e.split("#")
    if (b.length !== 2) {
        return rv
    }
    rv += b[1] + " of " + b[0]
    return rv
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
