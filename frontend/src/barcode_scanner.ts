import type _Alpine from "alpinejs"
import type { ZXing } from "zxing-v2.1.0/zxing.js"

interface Config {
    output: HTMLVideoElement
    enabled: boolean
}

interface ScanResult {
    format: string
    symbologyIdentifier: string
    text: string
    error: string
    position: {
        bottomLeft: ScanPos
        bottomRight: ScanPos
        topLeft: ScanPos
        topRight: ScanPos
    }
}

interface ScanPos {
    x: number
    y: number
}

class BarcodeScanner {
    output: HTMLVideoElement
    stream: MediaStream
    canvas: OffscreenCanvas
    ctx?: OffscreenCanvasRenderingContext2D | null
    zxing?: typeof ZXing

    private onFound: (res: ScanResult) => void

    constructor({
        output,
        stream,
        onFound,
    }: {
        output: HTMLVideoElement
        stream: MediaStream
        onFound: (res: ScanResult) => void
    }) {
        this.output = output
        this.stream = stream
        this.onFound = onFound
        this.canvas = new OffscreenCanvas(640, 480)
        this.ctx = this.canvas.getContext("2d", { willReadFrequently: true })
    }

    async start() {
        let libzxing = await import("zxing-v2.1.0/zxing.js")
        this.zxing = await libzxing.ZXing()

        this.output.srcObject = this.stream
        this.output.play()

        requestAnimationFrame(() => this._processFrame())
    }

    stop() {
        this.output.pause()
        this.output.srcObject = null
        this.ctx = null
        this.stream.getTracks().forEach((t) => {
            this.stream.removeTrack(t)
            t.stop()
        })
    }

    private _processFrame() {
        if (!this.ctx || this.stream) {
            return
        }

        this.ctx.drawImage(
            this.output,
            0,
            0,
            this.canvas.width,
            this.canvas.height,
        )

        let imageData = this.ctx.getImageData(
            0,
            0,
            this.canvas.width,
            this.canvas.height,
        )
        let sourceBuffer = imageData.data

        let buffer = this.zxing._malloc(sourceBuffer.byteLength)
        this.zxing.HEAPU8.set(sourceBuffer, buffer)
        let result = this.zxing.readBarcodeFromPixmap(
            buffer,
            this.canvas.width,
            this.canvas.height,
            false,
            "",
        )
        this.zxing._free(buffer)

        if (result.text) {
            this.onFound(result)
            return
        }

        requestAnimationFrame(() => this._processFrame())
    }
}

export function plugin(Alpine: typeof _Alpine) {
    Alpine.directive(
        "barcode-scanner",
        (_, { expression }, { evaluateLater, effect, cleanup }) => {
            let getConfig = evaluateLater<Config>(expression)

            let scanner: BarcodeScanner | undefined

            let onFound = (res: ScanResult) => {
                let path = `/assets/${res.text}`
                if (res.text.includes("://")) {
                    let u = new URL(res.text)
                    path = u.pathname
                }

                window.location.href = `${location.origin}${path}`
            }

            effect(() => {
                getConfig((config) => {
                    if (scanner) {
                        scanner.stop()
                        scanner = undefined
                    }

                    if (!config.enabled) {
                        return
                    }

                    navigator.mediaDevices
                        .getUserMedia({
                            audio: false,
                            video: {
                                facingMode: "environment",
                                frameRate: 60,
                            },
                        })
                        .then((stream) => {
                            scanner = new BarcodeScanner({
                                output: config.output,
                                stream,
                                onFound,
                            })

                            scanner.start()
                        })
                        .catch((e) => alert(e.message))
                })
            })

            cleanup(() => {
                if (scanner) {
                    scanner.stop()
                    scanner = undefined
                }
            })
        },
    )
}
