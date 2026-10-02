#!/usr/bin/env python3
import os
import sys
import subprocess
import urllib.request
import urllib.parse
import json
import time
import re

REPO = "giangsamne/TailRouter"

RELEASES = {
    "2.1.1": 401608330,
    "2.1.0": 400945564,
    "2.0.1": 398224481,
    "2.0.0": 394797416,
    "1.1.1": 401608406,
    "1.1.0": 400947289,
    "1.0.1": 398225324,
    "1.0.0": 394806550,
}

BASE_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

def get_token():
    token = os.environ.get("GITHUB_TOKEN")
    if not token:
        try:
            out = subprocess.check_output(["git", "config", "--get", "remote.origin.url"], cwd=BASE_DIR, text=True).strip()
            m = re.search(r'://[^:]+:([^@]+)@', out)
            if m:
                token = m.group(1)
        except Exception:
            pass
    if not token:
        print("❌ Cannot find GitHub token in GITHUB_TOKEN or git remote.origin.url")
        sys.exit(1)
    return token

def run_cmd(cmd, cwd=BASE_DIR):
    print(f"==> Running: {cmd}")
    res = subprocess.run(cmd, shell=True, cwd=cwd, text=True, capture_output=True)
    if res.returncode != 0:
        print(f"❌ Error (code {res.returncode}):\nSTDOUT: {res.stdout}\nSTDERR: {res.stderr}")
        sys.exit(res.returncode)
    return res.stdout

def get_headers():
    return {
        "Authorization": f"token {get_token()}",
        "Accept": "application/vnd.github.v3+json",
        "User-Agent": "TailRouter-Release-Script",
    }

def delete_existing_assets(release_id, target_names):
    url = f"https://api.github.com/repos/{REPO}/releases/{release_id}/assets"
    req = urllib.request.Request(url, headers=get_headers())
    with urllib.request.urlopen(req) as resp:
        assets = json.loads(resp.read().decode())
    
    for a in assets:
        if a["name"] in target_names:
            print(f"🗑️ Deleting old asset: {a['name']} (ID: {a['id']})")
            del_url = f"https://api.github.com/repos/{REPO}/releases/assets/{a['id']}"
            del_req = urllib.request.Request(del_url, headers=get_headers(), method="DELETE")
            try:
                with urllib.request.urlopen(del_req) as del_resp:
                    pass
            except Exception as e:
                print(f"⚠️ Failed to delete asset {a['id']}: {e}")
            time.sleep(0.5)

def upload_asset(release_id, file_path, name, content_type):
    file_size = os.path.getsize(file_path)
    print(f"🚀 Uploading {name} ({file_size} bytes) to Release {release_id}...")
    upload_url = f"https://uploads.github.com/repos/{REPO}/releases/{release_id}/assets?name={urllib.parse.quote(name)}"
    
    headers = get_headers()
    headers["Content-Type"] = content_type
    headers["Content-Length"] = str(file_size)

    with open(file_path, "rb") as f:
        data = f.read()

    req = urllib.request.Request(upload_url, data=data, headers=headers, method="POST")
    with urllib.request.urlopen(req) as resp:
        res = json.loads(resp.read().decode())
        print(f"✅ Uploaded {name} -> {res.get('browser_download_url')}")

def update_version_in_source(ver):
    server_go = os.path.join(BASE_DIR, "internal", "gateway", "server.go")
    with open(server_go, "r", encoding="utf-8") as f:
        content = f.read()
    content = re.sub(r'Version\s*=\s*"[^"]*"', f'Version = "{ver}-go-native"', content)
    with open(server_go, "w", encoding="utf-8") as f:
        f.write(content)

    main_go = os.path.join(BASE_DIR, "cmd", "tailrouter", "main.go")
    with open(main_go, "r", encoding="utf-8") as f:
        content = f.read()
    content = re.sub(r'⚡ TailRouter Native Engine v[^\n]+', f'⚡ TailRouter Native Engine v{ver}', content)
    with open(main_go, "w", encoding="utf-8") as f:
        f.write(content)

    index_html = os.path.join(BASE_DIR, "web", "index.html")
    with open(index_html, "r", encoding="utf-8") as f:
        content = f.read()
    content = re.sub(r'id="app-version-badge"[^>]*>[^<]*<', f'id="app-version-badge" style="background: rgba(56, 189, 248, 0.18); color: #38bdf8; border: 1px solid rgba(56, 189, 248, 0.4); font-size: 0.8rem; font-weight: 700; padding: 0.15rem 0.6rem; border-radius: 9999px; vertical-align: middle; margin-left: 6px;">v{ver}<', content)
    with open(index_html, "w", encoding="utf-8") as f:
        f.write(content)

def release(ver):
    if ver not in RELEASES:
        print(f"❌ Unknown version {ver}. Choose from {list(RELEASES.keys())}")
        sys.exit(1)
    
    rel_id = RELEASES[ver]
    tag = f"v{ver}"
    print(f"\n=======================================================")
    print(f"📦 Releasing TailRouter {tag} (Release ID: {rel_id})")
    print(f"=======================================================")

    update_version_in_source(ver)

    # 1. Compile & package all binaries
    run_cmd("./scripts/build_all.sh")

    # 2. Upload assets to GitHub Releases
    target_names = ["TailRouter-Linux.tar.gz", "TailRouter-macOS.zip", "TailRouter-Windows.zip"]
    delete_existing_assets(rel_id, target_names)

    upload_asset(rel_id, os.path.join(BASE_DIR, "release", "TailRouter-Linux.tar.gz"), "TailRouter-Linux.tar.gz", "application/gzip")
    upload_asset(rel_id, os.path.join(BASE_DIR, "release", "TailRouter-macOS.zip"), "TailRouter-macOS.zip", "application/zip")
    upload_asset(rel_id, os.path.join(BASE_DIR, "release", "TailRouter-Windows.zip"), "TailRouter-Windows.zip", "application/zip")

    # 3. Create snapshot tag & push
    run_cmd(f"./scripts/snapshot_tag.sh {tag}")
    print(f"🎉 Successfully released {tag}!")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: ./release_version.py <version|all>")
        sys.exit(1)
    target = sys.argv[1]
    if target == "all":
        for v in ["1.0.0", "1.0.1", "2.0.0", "2.0.1"]:
            release(v)
    else:
        release(target)
