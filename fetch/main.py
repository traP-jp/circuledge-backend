import os
import requests
import uuid
import time

url = 'https://md.trap.jp/notes'
cookie = os.environ.get("COOKIE")

headers = {
    "Cookie": cookie,
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.0.0",
    "Accept": "application/json"
}

#notesを取得し，idだけを抽出する
response = requests.get(url, headers=headers)
if response.status_code == 200:
    data = response.json()
    # "notes"キーがある場合はそこから取得
    notes = data["notes"] if "notes" in data else data
    note_idx = [note['id'] for note in notes]
else:
    print(f"Failed to fetch notes: {response.status_code} - {response.text}")

# 1.notesごとにPOSTリクエストを送信する
# 2.https://md.trap.jp/{{id}}/downloadでnoteのbody(.md)をダウンロードする
# 3.bodyを含めてPUTリクエストを送信する
MAX_RETRIES = 5

# 処理時間を計測
start_time = time.time()
for idx in note_idx:
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            # 安全のため10ミリ秒waitを入れる
            time.sleep(0.01)

            # 1. POSTリクエストを送信し，ノートを作成
            post_url = f"http://circuledge.ramdos.net:8080/api/v1/notes"
            post_response = requests.post(post_url, headers=headers)
            
            if post_response.status_code == 200:
                post_data = post_response.json()
                note_id = post_data.get("id")
                note_permission = post_data.get("permission")
                note_revision = post_data.get("revision")
                if not note_id:
                    print(f"Failed to create note for {idx}: No ID returned")
                    break

                # 2. bodyをダウンロード
                download_url = f"https://md.trap.jp/{idx}/download"
                download_response = requests.get(download_url, headers=headers)
                if download_response.status_code == 200:
                    body = download_response.text
                    
                else:
                    print(f"Failed to download note {idx}: {download_response.status_code} - {download_response.text}")
                    break
                
                # 3. PUTリクエストを送信
                put_url = f"http://circuledge.ramdos.net:8080/api/v1/notes/{note_id}"
                
                channel_uuid = str(uuid.uuid5(uuid.NAMESPACE_DNS, "trap_channel"))

                put_data = {
                    "permission": note_permission,
                    "body": body,
                    "channel": channel_uuid,
                    "revision": note_revision
                }
                put_response = requests.put(put_url, headers=headers, json=put_data)
                
                if put_response.status_code == 204:
                    print(f"Successfully updated note {idx}")
                    break
                else:
                    print(f"Failed to update note {idx}: {put_response.status_code} - {put_response.text}")
                    break
            else:
                print(f"Failed to create note for {idx}: {post_response.status_code} - {post_response.text}")
                break
        except Exception as e:
            print(f"Error processing note {idx} (attempt {attempt}): {e}")
            if attempt < MAX_RETRIES:
                print("Retrying...")
                time.sleep(3)
            else:
                print("Max retries reached. Skipping this note.")
                break
# 処理時間を表示
end_time = time.time()
elapsed_time = end_time - start_time
print(f"Total processing time: {elapsed_time:.2f} seconds")

