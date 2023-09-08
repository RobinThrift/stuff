import type _Alpine from "alpinejs"
import { quickScore, RangeTuple } from "quick-score"

interface Cmd {
    name: string
    icon: string
    url: string
}

type CmdCategroy = [string, Cmd[]]

function highlightMatches(str: string, matches: RangeTuple[]): string {
    if (!matches) {
        return str
    }

    let substrings = []
    let previousEnd = 0

    for (let [start, end] of matches) {
        const prefix = str.substring(previousEnd, start)
        const match = `<strong>${str.substring(start, end)}</strong>`

        substrings.push(prefix, match)
        previousEnd = end
    }

    substrings.push(str.substring(previousEnd))

    return substrings.join("")
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.data("commandpalette", () => ({
        init() {
            this.shown = this.commands
            this.$watch("search", () => this.onSearch())
        },

        show: false,

        commands: [
            [
                "Assets",
                [
                    {
                        name: "Add Asset",
                        icon: "plus",
                        url: "/assets/new",
                    },
                    {
                        name: "All Assets",
                        icon: "list",
                        url: "/assets",
                    },
                ],
            ],
            [
                "Tags",
                [
                    {
                        name: "All Tags",
                        icon: "list",
                        url: "/tags",
                    },
                ],
            ],
        ] as CmdCategroy[],
        shown: [] as CmdCategroy[],
        curr: [0, 0] as [number, number],
        search: "",

        onSearch() {
            if (this.search.length === 0) {
                this.curr = [0, 0]
                this.shown = this.commands
                return
            }

            this.shown = this.commands
                .map((cmds: CmdCategroy) => [
                    cmds[0],
                    cmds[1]
                        .map((c) => {
                            let matches: RangeTuple[] = []
                            let score = quickScore(c.name, this.search, matches)
                            return {
                                ...c,
                                name: highlightMatches(c.name, matches),
                                score,
                            }
                        })
                        .filter(({ score }) => score > 0),
                ])
                .filter((cmds: CmdCategroy) => cmds[1].length)

            if (this.shown.length === 0) {
                this.curr = undefined
            } else {
                this.curr = [0, 0]
            }
        },

        selectNext() {
            if (!this.curr) {
                return
            }

            let [cat, cmd] = this.curr
            if (cmd + 1 < this.shown[cat][1].length) {
                this.curr = [cat, cmd + 1]
                this.scrollToActive()
                return
            }

            if (cat + 1 < this.shown.length) {
                this.curr = [cat + 1, 0]
                this.scrollToActive()
            }
        },

        selectPrev() {
            if (!this.curr) {
                return
            }
            let [cat, cmd] = this.curr
            if (cmd - 1 >= 0) {
                this.curr = [cat, cmd - 1]
                this.scrollToActive()
                return
            }

            if (cmd - 1 < 0 && cat != 0) {
                this.curr = [cat - 1, this.shown[cat - 1][1].length - 1]
                this.scrollToActive()
            }
        },

        exec() {
            if (!this.curr) {
                if (!this.search) {
                    return
                }
                window.location.href =
                    location.origin + "/assets?query=" + this.search
                return
            }

            let [cat, cmd] = this.curr
            let execCmd = this.shown[cat][1][cmd]
            if (execCmd) {
                window.location.href = location.origin + execCmd.url
            }
        },

        scrollToActive() {
            if (!this.curr) {
                return
            }

            let el: HTMLElement | null = this.$refs.cmdsList.querySelector(
                ".command-palette-active",
            )
            if (!el) {
                return
            }

            let offset =
                el.offsetTop +
                el.offsetHeight * 2 -
                this.$refs.cmdsList.offsetHeight
            if (offset > 0) {
                this.$refs.cmdsList.scrollTo({ top: offset })
            } else {
                this.$refs.cmdsList.scrollTo({ top: 0 })
            }
        },
    }))
}
