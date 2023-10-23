import type _Alpine from "alpinejs"
import flatpickr from "flatpickr"

export function plugin(Alpine: typeof _Alpine) {
    Alpine.directive(
        "datepicker",
        (el, { expression }, { evaluate, cleanup }) => {
            let { value, format = "Y-m-d" } =
                (evaluate(expression) as
                    | { value?: string; format?: string }
                    | undefined) ?? {}

            if (value === "<nil>" || value === "") {
                value = undefined
            }

            let fp = flatpickr(el, {
                allowInput: true,
                dateFormat: format,
                defaultDate: value,
            })

            cleanup(() => {
                fp.destroy()
            })
        },
    )
}
