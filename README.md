# 简介

`xhe-vpn` 是一款可在浏览器中运行的 `vpn`, 基于 [`WireGuard`](https://www.wireguard.com/) 和 [`WebRTC`](https://webrtcforthecurious.com/)

目的是暴露在浏览器中运行的服务. 如: 点开浏览器运行一个数据统计服务, 用户通过 `xhe-vpn` 直接访问浏览器中的数据统计服务, 跳过了 cdn 这些中间人保证了数据安全. (这是一个很美好的愿景, 然而现在并没有做出这样的应用, 因为是浏览器中的应用需要特别考虑应用离线的问题, 难度比较大)

# 使用

虽然上面说了这样的愿景, 但实际上现在在用的用途是爬虫, 所以接下来会以爬虫场景来介绍

## 配置服务端

注: 由于尚未完全开源, 建议用户在虚拟机中测试

```sh
git clone https://github.com/remoon-net/xhe-webvpn.git xhe-webvpn
# 进入仓库示例文件夹
cd xhe-webvpn/example/
# 下载 xhe-vpn 并赋予可执行权限
wget -O xhe-vpn https://github.com/remoon-net/xhe-vpn-binary/releases/download/v0.0.20240401/xhe-vpn
chmod +x xhe-vpn
# 会使用 example/xhe-vpn.yaml 配置文件进行启动
sudo ./xhe-vpn
```

## 配置浏览器端

注入 js 最好的办法就是油猴脚本, 所以将会使用油猴脚本作为示例

```js
// ==UserScript==
// @name        vpn - Xhe Vpn
// @namespace   xhe-vpn.remoon.net
// @match       https://remoon.net/
// @version     0.0.1
// @author      -
// @require     https://unpkg.com/@remoon.net/xhe-vpn@0.0.2/dist/xhe-vpn.umd.js
// @run-at      document-start
// @description 01/04/2024, 00:00:00
// @grant       none
// ==/UserScript==

void (async function main() {
  const net = await vpn.connect({
    Key: "087ec6e14bbed210e7215cdc73468dfa23f080a1bfb8665b2fd809bd99d28379",
    Route: ["192.168.4.28/24"],
    Peer: [
      {
        Pubkey:
          "c4c8e984c5322c8184c72265b92b250fdb63688705f504ba003c88f03393cf28",
        PSK: "ba3ef732682972723e233daf6daaa748a6641e4c22b0bc726d94ca03b35055bb",
        Auto: 5,
        WHIP: ["http://127.0.0.1:2233"],
        Allow: ["192.168.4.29/32"],
      },
    ],
  });
  const srv = await net.listen("0.0.0.0:80", {
    fetch(req) {
      return new Response("ok");
    },
  });
})();
```

## 测试访问浏览器中的应用

回到服务器

```sh
curl http://192.168.4.28:80
# 应该输出 ok
```
