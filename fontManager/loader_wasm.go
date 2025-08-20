//go:build wasm

package fontManager

import (
	"syscall/js"

	"github.com/cdvelop/tinystring"
)

// getFontData
func (fm *FontManager) getFontData(path string) ([]byte, error) {
	return fetchFontBytes(path)
}

func fetchFontBytes(url string) ([]byte, error) {
	global := js.Global()
	fetch := global.Get("fetch")

	promise := fetch.Invoke(url)
	response := await(promise)

	if !response.Get("ok").Bool() {
		return nil, tinystring.Errf("failed to fetch font from %s", url)
	}

	bufferPromise := response.Call("arrayBuffer")
	arrayBuffer := await(bufferPromise)

	uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
	length := uint8Array.Get("length").Int()

	data := make([]byte, length)
	js.CopyBytesToGo(data, uint8Array)

	return data, nil
}

func await(promise js.Value) js.Value {
	done := make(chan js.Value)
	var result js.Value

	success := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		result = args[0]
		done <- result
		return nil
	})
	defer success.Release()

	failure := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		result = js.Null()
		done <- result
		return nil
	})
	defer failure.Release()

	promise.Call("then", success).Call("catch", failure)
	<-done
	return result
}
