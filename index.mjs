if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer();
    return await WebAssembly.instantiate(source, importObject);
  };
}

import { version } from "./package.json";
const defaultWasmUrl = `https://unpkg.com/@remoon.net/xhe-vpn@${version}/xhe-vpn.wasm`;

/**
 * @param {import("./index").Config} config
 * @returns {import("./index").Network}
 */
export async function connect(config) {
  const go = new Go();
  const { WASM: wasmLink = defaultWasmUrl } = config;
  const { instance } = await WebAssembly.instantiateStreaming(
    fetch(wasmLink),
    go.importObject
  );
  var vpn = {
    config: config,
    connect_result: null,
  };
  go.importObject.vpn = vpn;
  go.run(instance);
  const network = await vpn.connect_result;
  return network;
}
