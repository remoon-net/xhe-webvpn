wasm:
	GOOS=js GOARCH=wasm go build -o xhe-vpn.wasm
cp_execjs:
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js ./go_wasm_js/
