if (!("window" in globalThis)) {
  globalThis["window"] = globalThis;
}
if (!("RTCPeerConnection" in globalThis)) {
  const all = require("wrtc");
  for (let k in all) {
    globalThis[k] = all[k];
  }
}
