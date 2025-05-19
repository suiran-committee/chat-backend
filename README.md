# chat-backend

## localでの開発環境
1. mkcert(推奨) or 自己証明書作成
- mkcert
[開発環境をhttps化するmkcertの仕組み](https://qiita.com/k_kind/items/b87777efa3d29dcc4467)
```bash
# 1) mkcert をインストール
brew install mkcert nss

# 2) ローカル CA をシステムに登録  ※初回のみ
mkcert -install

# 3) バックエンド用の証明書を生成
mkcert -cert-file cert.pem -key-file key.pem localhost 127.0.0.1 ::1
```

- 自己証明書
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
null

or

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
<script setup>
import { ref, onMounted } from 'vue'

const messages = ref([])
const username = ref('')
const content = ref('')
let socket

const sendMessage = () => {
  if (!username.value || !content.value) return
  socket.send(JSON.stringify({ username: username.value, content: content.value }))
  content.value = ''
}

onMounted(async () => {
  try {
    const beUrl = "wss://localhost:8443";

    socket = new WebSocket(`${beUrl}/ws`);
    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      messages.value.push(msg);
    };

    const res = await fetch("https://localhost:8443/messages");
    const history = await res.json();
    messages.value.push(...history);
  } catch (err) {
    console.error("履歴取得失敗", err);
  }
});
</script>
```
