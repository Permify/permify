export default function TupleToHumanLanguage(tuple) {
    let rv = ""
    let p = tuple.split("@")
    if (p.length !== 2){
        return rv
    }
    let e = p[0]
    let s = p[1]
    let a = s.split("#")
    if (a.length === 1) {
        rv += a[0] + " is "
    }else {
        if (a[1] === "...") {
            rv += a[0] + " is "
        }else {
            rv += a[1] + " of " + a[0] + " is "
        }
    }
    let b = e.split("#")
    if (b.length !== 2){
        return rv
    }
    rv += b[1] + " of " + b[0]
    return rv
}
