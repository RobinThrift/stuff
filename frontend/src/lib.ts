export function get(path: string, obj: Record<string, unknown>) {
    // biome-ignore lint/suspicious/noExplicitAny: Can't really be known beforehand
    let res: any = obj
    let parts = path.split(".")
    for (let i = 0; i <= parts.length; i++) {
        let p = parts[i]
        if (Array.isArray(res)) {
            res = res.map((r) => get(parts.slice(i).join("."), r))
            continue
        }

        if (res[p]) {
            res = res[p]
        }
    }

    return res
}
