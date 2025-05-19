# chat-backend

## localでの開発環境
1. 自己証明書作成
```
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout key.pem -out cert.pem -days 365 \
  -subj "/CN=localhost"
```

2. 起動
```
docker compose -f 'docker-compose.yml' up -d --build
```


3. （落とす時）
```
docker-compose down -v
```


## API
### 履歴取得
#### /messages
##### req
``` bash
URL=https://localhost:8443

curl -sk $URL/messages | jq .
```

##### res
```
[
  {
    "username": "aaa",
    "content": "azzzz"
  },
  {
    "username": "aaa",
    "content": "xxxxx"
  },
...
]
```

### websocket
#### /ws
``` js
// 例
const beUrl = "wss://localhost:8443";

const socket = new WebSocket(`${beUrl}/ws`);
socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    messages.value.push(msg);
};
```


 ## Front
 ***※TLS化必須***

1.自己証明書作成
```
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout key.pem -out cert.pem -days 365 \
  -subj "/CN=localhost"
```

2. 設定(Viteの場合)
vite.config.js
```
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import fs from 'fs'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  server: {
    https: {
      key: fs.readFileSync(path.resolve(__dirname, 'cert/key.pem')),
      cert: fs.readFileSync(path.resolve(__dirname, 'cert/cert.pem')),
    },
    port: 5173,
    host: 'localhost',
  }
})
```
