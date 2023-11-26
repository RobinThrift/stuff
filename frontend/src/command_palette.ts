import type _Alpine from "alpinejs"
import { hasMatch } from "fzy.js"

interface Cmd {
    name: string
    icon: string
    url: string
    tags: string[]
}

type CmdCategroy = [string, Cmd[]]

export function plugin(Alpine: typeof _Alpine) {
    Alpine.data("commandpalette", (initParam) => {
        let { isAdmin } = initParam as { isAdmin?: boolean }

        let commands: CmdCategroy[] = [
            [
                "Assets",
                [
                    {
                        name: "Add Asset",
                        icon: "plus",
                        url: "/assets/new",
                        tags: ["create", "new"],
                    },
                    {
                        name: "All Assets",
                        icon: "package",
                        url: "/assets",
                        tags: ["list"],
                    },
                    {
                        name: "List Components",
                        icon: "stack-simple",
                        url: "/assets?type=component",
                        tags: ["list", "components"],
                    },
                    {
                        name: "List Consumables",
                        icon: "receipt",
                        url: "/assets?type=consumable",
                        tags: ["list", "consumables"],
                    },
                    {
                        name: "Import Assets",
                        icon: "arrow-square-in",
                        url: "/assets/import",
                        tags: [],
                    },
                ],
            ],
            [
                "Tags",
                [
                    {
                        name: "All Tags",
                        icon: "tag",
                        url: "/tags",
                        tags: ["list"],
                    },
                ],
            ],
        ]

        if (isAdmin) {
            commands.push([
                "Users",
                [
                    {
                        name: "All Users",
                        icon: "user",
                        url: "/users",
                        tags: ["list"],
                    },
                    {
                        name: "Create User",
                        icon: "user-plus",
                        url: "/users/new",
                        tags: ["add", "new"],
                    },
                ],
            ])
        }

        return {
            init() {
                this.shown = this.commands
                this.$watch("search", () => this.onSearch())
            },

            show: false,

            commands,
            shown: [] as CmdCategroy[],
            curr: [0, 0] as [number, number] | undefined,
            search: "",

            onSearch() {
                if (this.search.length === 0) {
                    this.curr = [0, 0]
                    this.shown = this.commands
                    return
                }

                this.shown = this.commands
                    .map((cmds: CmdCategroy): [string, ScoredCmd[]] => [
                        cmds[0],
                        cmds[1]
                            .map((c) => scoreCmd(this.search, c))
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

                if (cmd - 1 < 0 && cat !== 0) {
                    this.curr = [cat - 1, this.shown[cat - 1][1].length - 1]
                    this.scrollToActive()
                }
            },

            exec() {
                if (!this.curr) {
                    if (!this.search) {
                        return
                    }
                    window.location.href = `${location.origin}/assets?query=${this.search}`
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
        }
    })
}

type ScoredCmd = Cmd & { score: number }

function scoreCmd(input: string, cmd: Cmd): ScoredCmd {
    return {
        ...cmd,
        score: hasMatch(input, cmd.tags.join(" ") + cmd.name),
    }
}
