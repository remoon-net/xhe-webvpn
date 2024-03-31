require("./wrtc-polyfill");
require("./go_wasm_js/wasm_exec_node");
const fs = require("node:fs/promises");
const path = require("path");

const defaultWasmPath = path.join(module.path, "xhe-vpn.wasm");

/**@type {typeof connect} */
exports.connect = async function connect(config) {
  const go = new Go();
  const { WASM: wasmPath = defaultWasmPath } = config;
  const wasmBuf = await fs.readFile(wasmPath);
  const { instance } = await WebAssembly.instantiate(wasmBuf, go.importObject);
  var vpn = {
    config: config,
    connect_result: null,
  };
  go.importObject.vpn = vpn;
  go.run(instance);
  const network = await vpn.connect_result;
  return network;
};
